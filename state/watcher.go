package state

import (
	"labix.org/v2/mgo"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/log"
	"launchpad.net/juju-core/state/watcher"
	"launchpad.net/tomb"
	"strings"
)

// commonWatcher is part of all client watchers.
type commonWatcher struct {
	st   *State
	tomb tomb.Tomb
}

// Stop stops the watcher, and returns any error encountered while running
// or shutting down.
func (w *commonWatcher) Stop() error {
	w.tomb.Kill(nil)
	return w.tomb.Wait()
}

// Err returns any error encountered while running or shutting down, or
// tomb.ErrStillAlive if the watcher is still running.
func (w *commonWatcher) Err() error {
	return w.tomb.Err()
}

// RelationScopeWatcher observes changes to the set of units
// in a particular relation scope.
type RelationScopeWatcher struct {
	commonWatcher
	prefix     string
	ignore     string
	knownUnits map[string]bool
	changeChan chan *RelationScopeChange
}

// RelationScopeChange contains information about units that have
// entered or left a particular scope.
type RelationScopeChange struct {
	Entered []string
	Left    []string
}

// MachinePrincipalUnitsWatcher observes the assignment and removal of units
// to and from a machine.
type MachinePrincipalUnitsWatcher struct {
	commonWatcher
	machine    *Machine
	changeChan chan *MachinePrincipalUnitsChange
	knownUnits map[string]*Unit
}

// MachinePrincipalUnitsChange contains information about units that have been
// assigned to or removed from the machine.
type MachinePrincipalUnitsChange struct {
	Added   []*Unit
	Removed []*Unit
}

func hasString(changes []string, name string) bool {
	for _, v := range changes {
		if v == name {
			return true
		}
	}
	return false
}

func hasInt(changes []int, id int) bool {
	for _, v := range changes {
		if v == id {
			return true
		}
	}
	return false
}

// LifecyclesWatcher notifies about lifecycle changes for all entities of
// a given kind. The first event emitted will contain the ids of each such
// entity, regardless of life state; subsequent events are emitted whenever
// one such entity is added, or changes its lifecycle state. After an entity
// is found to be Dead, no further event will include it.
type LifecyclesWatcher struct {
	commonWatcher
	coll *mgo.Collection
	life map[string]Life
	out  chan []string
}

// WatchMachines returns a LifecyclesWatcher that notifies of changes to
// machines in the environment.
func (st *State) WatchMachines() *LifecyclesWatcher {
	return newLifecyclesWatcher(st, st.machines)
}

// WatchServices returns a LifecyclesWatcher that notifies of changes to
// services in the environment.
func (st *State) WatchServices() *LifecyclesWatcher {
	return newLifecyclesWatcher(st, st.services)
}

func newLifecyclesWatcher(st *State, coll *mgo.Collection) *LifecyclesWatcher {
	w := &LifecyclesWatcher{
		commonWatcher: commonWatcher{st: st},
		coll:          coll,
		life:          make(map[string]Life),
		out:           make(chan []string),
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop())
	}()
	return w
}

type lifeDoc struct {
	Id   string `bson:"_id"`
	Life Life
}

var lifeFields = D{{"_id", 1}, {"life", 1}}

// Changes returns the event channel for the LifecyclesWatcher.
func (w *LifecyclesWatcher) Changes() <-chan []string {
	return w.out
}

func (w *LifecyclesWatcher) initial() (ids []string, err error) {
	iter := w.coll.Find(nil).Select(lifeFields).Iter()
	var doc lifeDoc
	for iter.Next(&doc) {
		ids = append(ids, doc.Id)
		w.life[doc.Id] = doc.Life
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (w *LifecyclesWatcher) merge(ids []string, ch watcher.Change) ([]string, error) {
	id := ch.Id.(string)
	log.Printf("changed: %s", id)
	for _, pending := range ids {
		if id == pending {
			return ids, nil
		}
	}
	if ch.Revno == -1 {
		if life, ok := w.life[id]; ok && life != Dead {
			ids = append(ids, id)
		}
		delete(w.life, id)
		return ids, nil
	}
	doc := lifeDoc{Id: id, Life: Dead}
	err := w.coll.FindId(id).Select(lifeFields).One(&doc)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	log.Printf("%s; %s; %v", w.life[id], doc.Life, err)
	if life, ok := w.life[id]; !ok || doc.Life != life {
		ids = append(ids, id)
		if err != mgo.ErrNotFound {
			w.life[id] = doc.Life
		}
	}
	return ids, nil
}

func (w *LifecyclesWatcher) loop() (err error) {
	ch := make(chan watcher.Change)
	w.st.watcher.WatchCollection(w.coll.Name, ch)
	defer w.st.watcher.UnwatchCollection(w.coll.Name, ch)
	ids, err := w.initial()
	if err != nil {
		return err
	}
	out := w.out
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c := <-ch:
			if ids, err = w.merge(ids, c); err != nil {
				return err
			}
			if len(ids) > 0 {
				out = w.out
			}
		case out <- ids:
			ids = nil
			out = nil
		}
	}
	return nil
}

// ServiceUnitsWatcher notifies about the lifecycle changes of the units
// belonging to the service. The first event returned by the watcher is the
// set of names of all units that are part of the service, irrespective of
// their life state. Subsequent events return batches of newly added units
// and units which have changed their lifecycle.  After a unit is reported
// to be Dead, no further event will include it.
type ServiceUnitsWatcher struct {
	commonWatcher
	service *Service
	out     chan []string
	known   map[string]Life
}

// Changes returns the event channel for w.
func (w *ServiceUnitsWatcher) Changes() <-chan []string {
	return w.out
}

// WatchUnits returns a new ServiceUnitsWatcher for s.
func (s *Service) WatchUnits() *ServiceUnitsWatcher {
	return newServiceUnitsWatcher(s)
}

func newServiceUnitsWatcher(svc *Service) *ServiceUnitsWatcher {
	w := &ServiceUnitsWatcher{
		commonWatcher: commonWatcher{st: svc.st},
		known:         make(map[string]Life),
		out:           make(chan []string),
		service:       &Service{svc.st, svc.doc}, // Copy so it may be freely refreshed.
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop())
	}()
	return w
}

func (w *ServiceUnitsWatcher) merge(pending []string, name string) (changes []string, err error) {
	doc := unitDoc{}
	err = w.st.units.FindId(name).One(&doc)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	life, known := w.known[name]
	if err == mgo.ErrNotFound {
		delete(w.known, name)
		if known && life != Dead && !hasString(pending, name) {
			return append(pending, name), nil
		}
		return pending, nil
	}
	w.known[name] = doc.Life
	if !known {
		return append(pending, name), nil
	}
	if life == doc.Life || hasString(pending, name) {
		return pending, nil
	}
	return append(pending, name), nil
}

func (w *ServiceUnitsWatcher) initial() (changes []string, err error) {
	doc := &unitDoc{}
	iter := w.st.units.Find(D{{"service", w.service.doc.Name}}).Select(lifeFields).Iter()
	for iter.Next(doc) {
		w.known[doc.Name] = doc.Life
		changes = append(changes, doc.Name)
	}
	if iter.Err() != nil {
		return nil, err
	}
	return changes, nil
}

func (w *ServiceUnitsWatcher) loop() (err error) {
	ch := make(chan watcher.Change)
	w.st.watcher.WatchCollection(w.st.units.Name, ch)
	defer w.st.watcher.UnwatchCollection(w.st.units.Name, ch)
	changes, err := w.initial()
	if err != nil {
		return err
	}
	prefix := w.service.doc.Name + "/"
	out := w.out
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c := <-ch:
			name := c.Id.(string)
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			changes, err = w.merge(changes, name)
			if err != nil {
				return err
			}
			if len(changes) > 0 {
				out = w.out
			}
		case out <- changes:
			out = nil
			changes = nil
		}
	}
	return nil
}

// ServiceRelationsWatcher notifies about the lifecycle changes of the
// relations the service is in. The first event returned by the watcher is
// the set of ids of all relations that the service is part of, irrespective
// of their life state. Subsequent events return batches of newly added
// relations and relations which have changed their lifecycle. After a
// relation is reported to be Dead, no further event will include it.
type ServiceRelationsWatcher struct {
	commonWatcher
	service *Service
	out     chan []int
	known   map[string]relationDoc
}

// WatchRelations returns a new ServiceRelationsWatcher for s.
func (s *Service) WatchRelations() *ServiceRelationsWatcher {
	return newServiceRelationsWatcher(s)
}

func newServiceRelationsWatcher(s *Service) *ServiceRelationsWatcher {
	w := &ServiceRelationsWatcher{
		commonWatcher: commonWatcher{st: s.st},
		out:           make(chan []int),
		known:         make(map[string]relationDoc),
		service:       &Service{s.st, s.doc}, // Copy so it may be freely refreshed
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop())
	}()
	return w
}

func (w *ServiceRelationsWatcher) Stop() error {
	w.tomb.Kill(nil)
	return w.tomb.Wait()
}

// Changes returns the event channel for w.
func (w *ServiceRelationsWatcher) Changes() <-chan []int {
	return w.out
}

func (w *ServiceRelationsWatcher) initial() (new []int, err error) {
	doc := relationDoc{}
	iter := w.st.relations.Find(D{{"endpoints.servicename", w.service.doc.Name}}).Select(append(D{{"id", 1}}, lifeFields...)).Iter()
	for iter.Next(&doc) {
		w.known[doc.Key] = doc
		new = append(new, doc.Id)
	}
	if iter.Err() != nil {
		return nil, err
	}
	return new, nil
}

func (w *ServiceRelationsWatcher) merge(pending []int, key string) (new []int, err error) {
	doc := relationDoc{}
	err = w.st.relations.FindId(key).One(&doc)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	old, known := w.known[key]
	if err == mgo.ErrNotFound {
		if known && old.Life != Dead && !hasInt(pending, old.Id) {
			delete(w.known, key)
			return append(pending, old.Id), nil
		}
		return pending, nil
	}
	w.known[key] = doc
	if !known {
		return append(pending, doc.Id), nil
	}
	if old.Life == doc.Life || hasInt(pending, old.Id) {
		return pending, nil
	}
	return append(pending, doc.Id), nil
}

func (w *ServiceRelationsWatcher) loop() (err error) {
	ch := make(chan watcher.Change)
	w.st.watcher.WatchCollection(w.st.relations.Name, ch)
	defer w.st.watcher.UnwatchCollection(w.st.relations.Name, ch)
	changes, err := w.initial()
	if err != nil {
		return err
	}
	prefix1 := w.service.doc.Name + ":"
	prefix2 := " " + w.service.doc.Name + ":"
	out := w.out
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c := <-ch:
			key := c.Id.(string)
			if !strings.HasPrefix(key, prefix1) && !strings.Contains(key, prefix2) {
				continue
			}
			changes, err = w.merge(changes, key)
			if err != nil {
				return err
			}
			if len(changes) > 0 {
				out = w.out
			}
		case out <- changes:
			out = nil
			changes = nil
		}
	}
	return nil
}

// WatchPrincipalUnits returns a watcher for observing units being
// added to or removed from the machine.
func (m *Machine) WatchPrincipalUnits() *MachinePrincipalUnitsWatcher {
	return newMachinePrincipalUnitsWatcher(m)
}

// newMachinePrincipalUnitsWatcher creates and starts a watcher to watch information
// about units being added to or deleted from the machine.
func newMachinePrincipalUnitsWatcher(m *Machine) *MachinePrincipalUnitsWatcher {
	w := &MachinePrincipalUnitsWatcher{
		changeChan:    make(chan *MachinePrincipalUnitsChange),
		machine:       m,
		knownUnits:    make(map[string]*Unit),
		commonWatcher: commonWatcher{st: m.st},
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.changeChan)
		w.tomb.Kill(w.loop())
	}()
	return w
}

// Changes returns a channel that will receive changes when units are
// added or deleted. The Added field in the first event on the channel
// holds the initial state as returned by Machine.Units.
func (w *MachinePrincipalUnitsWatcher) Changes() <-chan *MachinePrincipalUnitsChange {
	return w.changeChan
}

func (w *MachinePrincipalUnitsWatcher) mergeChange(changes *MachinePrincipalUnitsChange, ch watcher.Change) (err error) {
	err = w.machine.Refresh()
	if err != nil {
		return err
	}
	units := make(map[string]*Unit)
	for _, name := range w.machine.doc.Principals {
		var unit *Unit
		doc := &unitDoc{}
		if _, ok := w.knownUnits[name]; !ok {
			err = w.st.units.FindId(name).One(doc)
			if err == mgo.ErrNotFound {
				continue
			}
			if err != nil {
				return err
			}
			unit = newUnit(w.st, doc)
			changes.Added = append(changes.Added, unit)
			w.knownUnits[name] = unit
		}
		units[name] = unit
	}
	for name, unit := range w.knownUnits {
		if _, ok := units[name]; !ok {
			changes.Removed = append(changes.Removed, unit)
			delete(w.knownUnits, name)
		}
	}
	return nil
}

func (changes *MachinePrincipalUnitsChange) isEmpty() bool {
	return len(changes.Added)+len(changes.Removed) == 0
}

func (w *MachinePrincipalUnitsWatcher) getInitialEvent() (initial *MachinePrincipalUnitsChange, err error) {
	changes := &MachinePrincipalUnitsChange{}
	docs := []unitDoc{}
	err = w.st.units.Find(D{{"_id", D{{"$in", w.machine.doc.Principals}}}}).All(&docs)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		unit := newUnit(w.st, &doc)
		w.knownUnits[doc.Name] = unit
		changes.Added = append(changes.Added, unit)
	}
	return changes, nil
}

func (w *MachinePrincipalUnitsWatcher) loop() (err error) {
	ch := make(chan watcher.Change)
	w.st.watcher.Watch(w.st.machines.Name, w.machine.doc.Id, w.machine.doc.TxnRevno, ch)
	defer w.st.watcher.Unwatch(w.st.machines.Name, w.machine.doc.Id, ch)
	changes, err := w.getInitialEvent()
	if err != nil {
		return err
	}
	for {
		for changes != nil {
			select {
			case <-w.st.watcher.Dead():
				return watcher.MustErr(w.st.watcher)
			case <-w.tomb.Dying():
				return tomb.ErrDying
			case c := <-ch:
				err := w.mergeChange(changes, c)
				if err != nil {
					return err
				}
			case w.changeChan <- changes:
				changes = nil
			}
		}
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c := <-ch:
			changes = &MachinePrincipalUnitsChange{}
			err := w.mergeChange(changes, c)
			if err != nil {
				return err
			}
			if changes.isEmpty() {
				changes = nil
			}
		}
	}
	return nil
}

func newRelationScopeWatcher(st *State, scope, ignore string) *RelationScopeWatcher {
	w := &RelationScopeWatcher{
		commonWatcher: commonWatcher{st: st},
		prefix:        scope + "#",
		ignore:        ignore,
		changeChan:    make(chan *RelationScopeChange),
		knownUnits:    make(map[string]bool),
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.changeChan)
		w.tomb.Kill(w.loop())
	}()
	return w
}

// Changes returns a channel that will receive changes when units enter and
// leave a relation scope. The Entered field in the first event on the channel
// holds the initial state.
func (w *RelationScopeWatcher) Changes() <-chan *RelationScopeChange {
	return w.changeChan
}

func (changes *RelationScopeChange) isEmpty() bool {
	return len(changes.Entered)+len(changes.Left) == 0
}

func (w *RelationScopeWatcher) mergeChange(changes *RelationScopeChange, ch watcher.Change) (err error) {
	doc := &relationScopeDoc{ch.Id.(string)}
	if !strings.HasPrefix(doc.Key, w.prefix) {
		return nil
	}
	name := doc.unitName()
	if name == w.ignore {
		return nil
	}
	if ch.Revno == -1 {
		if w.knownUnits[name] {
			changes.Left = append(changes.Left, name)
			delete(w.knownUnits, name)
		}
		return nil
	}
	if !w.knownUnits[name] {
		changes.Entered = append(changes.Entered, name)
		w.knownUnits[name] = true
	}
	return nil
}

func (w *RelationScopeWatcher) getInitialEvent() (initial *RelationScopeChange, err error) {
	changes := &RelationScopeChange{}
	docs := []relationScopeDoc{}
	sel := D{{"_id", D{{"$regex", "^" + w.prefix}}}}
	err = w.st.relationScopes.Find(sel).All(&docs)
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		if name := doc.unitName(); name != w.ignore {
			changes.Entered = append(changes.Entered, name)
			w.knownUnits[name] = true
		}
	}
	return changes, nil
}

func (w *RelationScopeWatcher) loop() error {
	ch := make(chan watcher.Change)
	w.st.watcher.WatchCollection(w.st.relationScopes.Name, ch)
	defer w.st.watcher.UnwatchCollection(w.st.relationScopes.Name, ch)
	changes, err := w.getInitialEvent()
	if err != nil {
		return err
	}
	for {
		for changes != nil {
			select {
			case <-w.st.watcher.Dead():
				return watcher.MustErr(w.st.watcher)
			case <-w.tomb.Dying():
				return tomb.ErrDying
			case c := <-ch:
				if err := w.mergeChange(changes, c); err != nil {
					return err
				}
			case w.changeChan <- changes:
				changes = nil
			}
		}
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c := <-ch:
			changes = &RelationScopeChange{}
			if err := w.mergeChange(changes, c); err != nil {
				return err
			}
			if changes.isEmpty() {
				changes = nil
			}
		}
	}
	return nil
}

// RelationUnitsWatcher sends notifications of units entering and leaving the
// scope of a RelationUnit, and changes to the settings of those units known
// to have entered.
type RelationUnitsWatcher struct {
	commonWatcher
	sw       *RelationScopeWatcher
	watching map[string]bool
	updates  chan watcher.Change
	out      chan RelationUnitsChange
}

// RelationUnitsChange holds notifications of units entering and leaving the
// scope of a RelationUnit, and changes to the settings of those units known
// to have entered.
//
// When a counterpart first enters scope, it is/ noted in the Joined field,
// and its settings are noted in the Changed field. Subsequently, settings
// changes will be noted in the Changed field alone, until the couterpart
// leaves the scope; at that point, it will be noted in the Departed field,
// and no further events will be sent for that counterpart unit.
type RelationUnitsChange struct {
	Joined   []string
	Changed  map[string]UnitSettings
	Departed []string
}

// Watch returns a watcher that notifies of changes to conterpart units in
// the relation.
func (ru *RelationUnit) Watch() *RelationUnitsWatcher {
	return newRelationUnitsWatcher(ru)
}

func newRelationUnitsWatcher(ru *RelationUnit) *RelationUnitsWatcher {
	w := &RelationUnitsWatcher{
		commonWatcher: commonWatcher{st: ru.st},
		sw:            ru.WatchScope(),
		watching:      map[string]bool{},
		updates:       make(chan watcher.Change),
		out:           make(chan RelationUnitsChange),
	}
	go func() {
		defer w.finish()
		w.tomb.Kill(w.loop())
	}()
	return w
}

// Changes returns a channel that will receive the changes to
// counterpart units in a relation. The first event on the
// channel holds the initial state of the relation in its
// Joined and Changed fields.
func (w *RelationUnitsWatcher) Changes() <-chan RelationUnitsChange {
	return w.out
}

func (changes *RelationUnitsChange) empty() bool {
	return len(changes.Joined)+len(changes.Changed)+len(changes.Departed) == 0
}

// mergeSettings reads the relation settings node for the unit with the
// supplied id, and sets a value in the Changed field keyed on the unit's
// name. It returns the mgo/txn revision number of the settings node.
func (w *RelationUnitsWatcher) mergeSettings(changes *RelationUnitsChange, key string) (int64, error) {
	node, err := readSettings(w.st, key)
	if err != nil {
		return -1, err
	}
	name := (&relationScopeDoc{key}).unitName()
	settings := UnitSettings{node.txnRevno, node.Map()}
	if changes.Changed == nil {
		changes.Changed = map[string]UnitSettings{name: settings}
	} else {
		changes.Changed[name] = settings
	}
	return node.txnRevno, nil
}

// mergeScope starts and stops settings watches on the units entering and
// leaving the scope in the supplied RelationScopeChange event, and applies
// the expressed changes to the supplied RelationUnitsChange event.
func (w *RelationUnitsWatcher) mergeScope(changes *RelationUnitsChange, c *RelationScopeChange) error {
	for _, name := range c.Entered {
		key := w.sw.prefix + name
		revno, err := w.mergeSettings(changes, key)
		if err != nil {
			return err
		}
		changes.Joined = append(changes.Joined, name)
		changes.Departed = remove(changes.Departed, name)
		w.st.watcher.Watch(w.st.settings.Name, key, revno, w.updates)
		w.watching[key] = true
	}
	for _, name := range c.Left {
		key := w.sw.prefix + name
		changes.Departed = append(changes.Departed, name)
		if changes.Changed != nil {
			delete(changes.Changed, name)
		}
		changes.Joined = remove(changes.Joined, name)
		w.st.watcher.Unwatch(w.st.settings.Name, key, w.updates)
		delete(w.watching, key)
	}
	return nil
}

// remove removes s from strs and returns the modified slice.
func remove(strs []string, s string) []string {
	for i, v := range strs {
		if s == v {
			strs[i] = strs[len(strs)-1]
			return strs[:len(strs)-1]
		}
	}
	return strs
}

func (w *RelationUnitsWatcher) finish() {
	watcher.Stop(w.sw, &w.tomb)
	for key := range w.watching {
		w.st.watcher.Unwatch(w.st.settings.Name, key, w.updates)
	}
	close(w.updates)
	close(w.out)
	w.tomb.Done()
}

func (w *RelationUnitsWatcher) loop() (err error) {
	sentInitial := false
	changes := RelationUnitsChange{}
	out := w.out
	out = nil
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case c, ok := <-w.sw.Changes():
			if !ok {
				return watcher.MustErr(w.sw)
			}
			if err = w.mergeScope(&changes, c); err != nil {
				return err
			}
			if !sentInitial || !changes.empty() {
				out = w.out
			} else {
				out = nil
			}
		case c := <-w.updates:
			if _, err = w.mergeSettings(&changes, c.Id.(string)); err != nil {
				return err
			}
			out = w.out
		case out <- changes:
			sentInitial = true
			changes = RelationUnitsChange{}
			out = nil
		}
	}
	panic("unreachable")
}

// EnvironConfigWatcher observes changes to the
// environment configuration.
type EnvironConfigWatcher struct {
	commonWatcher
	out chan *config.Config
}

// WatchEnvironConfig returns a watcher for observing changes
// to the environment configuration.
func (s *State) WatchEnvironConfig() *EnvironConfigWatcher {
	return newEnvironConfigWatcher(s)
}

func newEnvironConfigWatcher(s *State) *EnvironConfigWatcher {
	w := &EnvironConfigWatcher{
		commonWatcher: commonWatcher{st: s},
		out:           make(chan *config.Config),
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop())
	}()
	return w
}

// Changes returns a channel that will receive the new environment
// configuration when a change is detected. Note that multiple changes may
// be observed as a single event in the channel.
func (w *EnvironConfigWatcher) Changes() <-chan *config.Config {
	return w.out
}

func (w *EnvironConfigWatcher) loop() (err error) {
	sw := w.st.watchSettings("e")
	defer sw.Stop()
	out := w.out
	out = nil
	cfg := &config.Config{}
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case settings, ok := <-sw.Changes():
			if !ok {
				return watcher.MustErr(sw)
			}
			cfg, err = config.New(settings.Map())
			if err == nil {
				out = w.out
			} else {
				out = nil
			}
		case out <- cfg:
			out = nil
		}
	}
	return nil
}

type settingsWatcher struct {
	commonWatcher
	out chan *Settings
}

// watchSettings creates a watcher for observing changes to settings.
func (s *State) watchSettings(key string) *settingsWatcher {
	return newSettingsWatcher(s, key)
}

func newSettingsWatcher(s *State, key string) *settingsWatcher {
	w := &settingsWatcher{
		commonWatcher: commonWatcher{st: s},
		out:           make(chan *Settings),
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop(key))
	}()
	return w
}

// Changes returns a channel that will receive the new settings.
// Multiple changes may be observed as a single event in the channel.
func (w *settingsWatcher) Changes() <-chan *Settings {
	return w.out
}

func (w *settingsWatcher) loop(key string) (err error) {
	ch := make(chan watcher.Change)
	revno := int64(-1)
	settings, err := readSettings(w.st, key)
	if err == nil {
		revno = settings.txnRevno
	} else if !IsNotFound(err) {
		return err
	}
	w.st.watcher.Watch(w.st.settings.Name, key, revno, ch)
	defer w.st.watcher.Unwatch(w.st.settings.Name, key, ch)
	out := w.out
	if revno == -1 {
		out = nil
	}
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case <-ch:
			settings, err = readSettings(w.st, key)
			if err != nil {
				return err
			}
			out = w.out
		case out <- settings:
			out = nil
		}
	}
	return nil
}

type ConfigWatcher struct {
	*settingsWatcher
}

func (s *Service) WatchConfig() *ConfigWatcher {
	return &ConfigWatcher{newSettingsWatcher(s.st, "s#"+s.Name())}
}

// EntityWatcher observes changes to a state entity.
type EntityWatcher struct {
	commonWatcher
	changeChan chan struct{}
}

// Watch return a watcher for observing changes to a service.
func (s *Service) Watch() *EntityWatcher {
	return newEntityWatcher(s.st, s.st.services.Name, s.doc.Name, s.doc.TxnRevno)
}

// Watch return a watcher for observing changes to a unit.
func (u *Unit) Watch() *EntityWatcher {
	return newEntityWatcher(u.st, u.st.units.Name, u.doc.Name, u.doc.TxnRevno)
}

// Watch return a watcher for observing changes to a machine.
func (m *Machine) Watch() *EntityWatcher {
	return newEntityWatcher(m.st, m.st.machines.Name, m.doc.Id, m.doc.TxnRevno)
}

func newEntityWatcher(st *State, coll string, key interface{}, revno int64) *EntityWatcher {
	w := &EntityWatcher{
		commonWatcher: commonWatcher{st: st},
		changeChan:    make(chan struct{}),
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.changeChan)
		ch := make(chan watcher.Change)
		w.st.watcher.Watch(coll, key, revno, ch)
		defer w.st.watcher.Unwatch(coll, key, ch)
		w.tomb.Kill(w.loop(ch))
	}()
	return w
}

// Changes returns the event channel for the EntityWatcher.
func (w *EntityWatcher) Changes() <-chan struct{} {
	return w.changeChan
}

func (w *EntityWatcher) loop(ch <-chan watcher.Change) (err error) {
	out := w.changeChan
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case <-ch:
			out = w.changeChan
		case out <- struct{}{}:
			out = nil
		}
	}
	return nil
}

// MachineUnitsWatcher notifies about assignments and lifecycle changes
// for all units of a machine.
//
// The first event emitted contains the unit names of all units currently
// assigned to the machine, irrespective of their life state. From then on,
// a new event is emitted whenever a unit is assigned to or unassigned from
// the machine, or the lifecycle of a unit that is currently assigned to
// the machine changes.
//
// After a unit is found to be Dead, no further event will include it.
type MachineUnitsWatcher struct {
	commonWatcher
	machine *Machine
	out     chan []string
	in      chan watcher.Change
	known   map[string]Life
}

// WatchUnits returns a new MachineUnitsWatcher for m.
func (m *Machine) WatchUnits() *MachineUnitsWatcher {
	return newMachineUnitsWatcher(m)
}

func newMachineUnitsWatcher(m *Machine) *MachineUnitsWatcher {
	w := &MachineUnitsWatcher{
		commonWatcher: commonWatcher{st: m.st},
		out:           make(chan []string),
		in:            make(chan watcher.Change),
		known:         make(map[string]Life),
		machine:       &Machine{m.st, m.doc}, // Copy so it may be freely refreshed
	}
	go func() {
		defer w.tomb.Done()
		defer close(w.out)
		w.tomb.Kill(w.loop())
	}()
	return w
}

// Changes returns the event channel for w.
func (w *MachineUnitsWatcher) Changes() <-chan []string {
	return w.out
}

func (w *MachineUnitsWatcher) updateMachine(pending []string) (new []string, err error) {
	err = w.machine.Refresh()
	if err != nil {
		return nil, err
	}
	for _, unit := range w.machine.doc.Principals {
		if _, ok := w.known[unit]; !ok {
			pending, err = w.merge(pending, unit)
			if err != nil {
				return nil, err
			}
		}
	}
	return pending, nil
}

func (w *MachineUnitsWatcher) merge(pending []string, unit string) (new []string, err error) {
	doc := unitDoc{}
	err = w.st.units.FindId(unit).One(&doc)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	life, known := w.known[unit]
	if err == mgo.ErrNotFound || doc.Principal == "" && (doc.MachineId == "" || doc.MachineId != w.machine.doc.Id) {
		// Unit was removed or unassigned from w.machine.
		if known {
			delete(w.known, unit)
			w.st.watcher.Unwatch(w.st.units.Name, unit, w.in)
			if life != Dead && !hasString(pending, unit) {
				pending = append(pending, unit)
			}
			for _, subunit := range doc.Subordinates {
				if sublife, subknown := w.known[subunit]; subknown {
					delete(w.known, subunit)
					w.st.watcher.Unwatch(w.st.units.Name, subunit, w.in)
					if sublife != Dead && !hasString(pending, subunit) {
						pending = append(pending, subunit)
					}
				}
			}
		}
		return pending, nil
	}
	if !known {
		w.st.watcher.Watch(w.st.units.Name, unit, doc.TxnRevno, w.in)
		pending = append(pending, unit)
	} else if life != doc.Life && !hasString(pending, unit) {
		pending = append(pending, unit)
	}
	w.known[unit] = doc.Life
	for _, subunit := range doc.Subordinates {
		if _, ok := w.known[subunit]; !ok {
			pending, err = w.merge(pending, subunit)
			if err != nil {
				return nil, err
			}
		}
	}
	return pending, nil
}

func (w *MachineUnitsWatcher) loop() (err error) {
	defer func() {
		for _, unit := range w.known {
			w.st.watcher.Unwatch(w.st.units.Name, unit, w.in)
		}
	}()
	machineCh := make(chan watcher.Change)
	w.st.watcher.Watch(w.st.machines.Name, w.machine.doc.Id, w.machine.doc.TxnRevno, machineCh)
	defer w.st.watcher.Unwatch(w.st.machines.Name, w.machine.doc.Id, machineCh)
	changes, err := w.updateMachine([]string(nil))
	if err != nil {
		return err
	}
	out := w.out
	for {
		select {
		case <-w.st.watcher.Dead():
			return watcher.MustErr(w.st.watcher)
		case <-w.tomb.Dying():
			return tomb.ErrDying
		case <-machineCh:
			changes, err = w.updateMachine(changes)
			if err != nil {
				return err
			}
			if len(changes) > 0 {
				out = w.out
			}
		case c := <-w.in:
			changes, err = w.merge(changes, c.Id.(string))
			if err != nil {
				return err
			}
			if len(changes) > 0 {
				out = w.out
			}
		case out <- changes:
			out = nil
			changes = nil
		}
	}
	panic("unreachable")
}
