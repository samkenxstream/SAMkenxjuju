// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package secretbackends_test

import (
	"os"
	"path/filepath"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/juju/cmd/v3/cmdtesting"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	apisecretbackends "github.com/juju/juju/api/client/secretbackends"
	"github.com/juju/juju/cmd/juju/secretbackends"
	"github.com/juju/juju/cmd/juju/secretbackends/mocks"
	"github.com/juju/juju/jujuclient"
)

type UpdateSuite struct {
	jujutesting.IsolationSuite
	store                   *jujuclient.MemStore
	updateSecretBackendsAPI *mocks.MockUpdateSecretBackendsAPI
}

var _ = gc.Suite(&UpdateSuite{})

func (s *UpdateSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	store := jujuclient.NewMemStore()
	store.Controllers["mycontroller"] = jujuclient.ControllerDetails{}
	store.CurrentControllerName = "mycontroller"
	s.store = store
}

func (s *UpdateSuite) setup(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.updateSecretBackendsAPI = mocks.NewMockUpdateSecretBackendsAPI(ctrl)

	return ctrl
}

func (s *UpdateSuite) TestUpdateInitError(c *gc.C) {
	for _, t := range []struct {
		args []string
		err  string
	}{{
		args: []string{},
		err:  "must specify backend name",
	}, {
		args: []string{"myvault"},
		err:  "must specify a config file or key/reset values",
	}, {
		args: []string{"myvault", "foo=bar", "token-rotate=1s"},
		err:  `token rotate interval "1s" less than 1h not valid`,
	}, {
		args: []string{"myvault", "foo=bar", "--config", "/path/to/nowhere"},
		err:  `open /path/to/nowhere: no such file or directory`,
	}} {
		_, err := cmdtesting.RunCommand(c, secretbackends.NewUpdateCommandForTest(s.store, s.updateSecretBackendsAPI), t.args...)
		c.Check(err, gc.ErrorMatches, t.err)
	}
}

func (s *UpdateSuite) TestUpdate(c *gc.C) {
	defer s.setup(c).Finish()

	s.updateSecretBackendsAPI.EXPECT().UpdateSecretBackend(
		apisecretbackends.UpdateSecretBackend{
			Name:                "myvault",
			TokenRotateInterval: ptr(666 * time.Minute),
			Config:              map[string]interface{}{"endpoint": "http://vault"},
		}, true).Return(nil)
	s.updateSecretBackendsAPI.EXPECT().Close().Return(nil)

	_, err := cmdtesting.RunCommand(c, secretbackends.NewUpdateCommandForTest(s.store, s.updateSecretBackendsAPI),
		"myvault", "endpoint=http://vault", "token-rotate=666m", "--force",
	)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *UpdateSuite) TestUpdateName(c *gc.C) {
	defer s.setup(c).Finish()

	s.updateSecretBackendsAPI.EXPECT().UpdateSecretBackend(
		apisecretbackends.UpdateSecretBackend{
			Name:       "myvault",
			NameChange: ptr("myvault2"),
			Config:     map[string]interface{}{"endpoint": "http://vault"},
		}, false).Return(nil)
	s.updateSecretBackendsAPI.EXPECT().Close().Return(nil)

	_, err := cmdtesting.RunCommand(c, secretbackends.NewUpdateCommandForTest(s.store, s.updateSecretBackendsAPI),
		"myvault", "endpoint=http://vault", "name=myvault2",
	)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *UpdateSuite) TestUpdateResetTokenRotate(c *gc.C) {
	defer s.setup(c).Finish()

	s.updateSecretBackendsAPI.EXPECT().UpdateSecretBackend(
		apisecretbackends.UpdateSecretBackend{
			Name:                "myvault",
			TokenRotateInterval: ptr(0 * time.Second),
			Config:              map[string]interface{}{"endpoint": "http://vault"},
		}, false).Return(nil)
	s.updateSecretBackendsAPI.EXPECT().Close().Return(nil)

	_, err := cmdtesting.RunCommand(c, secretbackends.NewUpdateCommandForTest(s.store, s.updateSecretBackendsAPI),
		"myvault", "endpoint=http://vault", "token-rotate=0",
	)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *UpdateSuite) TestUpdateFromFile(c *gc.C) {
	defer s.setup(c).Finish()

	fname := filepath.Join(c.MkDir(), "cfg.yaml")
	err := os.WriteFile(fname, []byte("endpoint: http://vault"), 0644)
	c.Assert(err, jc.ErrorIsNil)
	s.updateSecretBackendsAPI.EXPECT().UpdateSecretBackend(
		apisecretbackends.UpdateSecretBackend{
			Name:                "myvault",
			TokenRotateInterval: ptr(666 * time.Minute),
			Config: map[string]interface{}{
				"endpoint": "http://vault",
				"token":    "s.666",
			},
		}, false).Return(nil)
	s.updateSecretBackendsAPI.EXPECT().Close().Return(nil)

	_, err = cmdtesting.RunCommand(c, secretbackends.NewUpdateCommandForTest(s.store, s.updateSecretBackendsAPI),
		"myvault", "token=s.666", "token-rotate=666m", "--config", fname,
	)
	c.Assert(err, jc.ErrorIsNil)
}
