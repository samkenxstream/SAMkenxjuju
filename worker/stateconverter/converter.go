// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package stateconverter

import (
	"github.com/juju/errors"
	"github.com/juju/names/v4"

	"github.com/juju/juju/api/agent/machiner"
	agenterrors "github.com/juju/juju/cmd/jujud/agent/errors"
	"github.com/juju/juju/core/model"
	"github.com/juju/juju/core/watcher"
)

type config struct {
	machineTag names.MachineTag
	machiner   Machiner
	logger     Logger
}

// NewConverter returns a new notify watch handler that will convert the given machine &
// agent to a controller.
func NewConverter(cfg config) watcher.NotifyHandler {
	return &converter{
		machiner:   cfg.machiner,
		machineTag: cfg.machineTag,
		logger:     cfg.logger,
	}
}

// converter is a NotifyWatchHandler that converts a unit hosting machine to a
// state machine.
type converter struct {
	machineTag names.MachineTag
	machiner   Machiner
	machine    Machine
	logger     Logger
}

// wrapper is a wrapper around api/machiner.State to match the (local) machiner
// interface.
type wrapper struct {
	m *machiner.State
}

// Machine implements machiner.Machine and returns a machine from the wrapper
// api/machiner.
func (w wrapper) Machine(tag names.MachineTag) (Machine, error) {
	m, err := w.m.Machine(tag)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// SetUp implements NotifyWatchHandler's SetUp method. It returns a watcher that
// checks for changes to the current machine.
func (c *converter) SetUp() (watcher.NotifyWatcher, error) {
	c.logger.Tracef("Calling SetUp for %s", c.machineTag)
	m, err := c.machiner.Machine(c.machineTag)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.machine = m
	return m.Watch()
}

// Handle implements NotifyWatchHandler's Handle method.  If the change means
// that the machine is now expected to manage the environment,
// we throw a fatal error to instigate agent restart.
func (c *converter) Handle(_ <-chan struct{}) error {
	c.logger.Tracef("Calling Handle for %s", c.machineTag)
	results, err := c.machine.Jobs()
	if err != nil {
		return errors.Annotate(err, "can't get jobs for machine")
	}
	if !model.AnyJobNeedsState(results.Jobs...) {
		return nil
	}

	return &agenterrors.FatalError{Err: "bounce agent to pick up new jobs"}
}

// TearDown implements NotifyWatchHandler's TearDown method.
func (c *converter) TearDown() error {
	return nil
}
