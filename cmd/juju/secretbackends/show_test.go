// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package secretbackends_test

import (
	"time"

	"github.com/golang/mock/gomock"
	"github.com/juju/cmd/v3/cmdtesting"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	apisecretbackends "github.com/juju/juju/api/client/secretbackends"
	"github.com/juju/juju/cmd/juju/secretbackends"
	"github.com/juju/juju/cmd/juju/secretbackends/mocks"
	"github.com/juju/juju/core/status"
	"github.com/juju/juju/jujuclient"
)

type ShowSuite struct {
	jujutesting.IsolationSuite
	store             *jujuclient.MemStore
	secretBackendsAPI *mocks.MockListSecretBackendsAPI
}

var _ = gc.Suite(&ShowSuite{})

func (s *ShowSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	store := jujuclient.NewMemStore()
	store.Controllers["mycontroller"] = jujuclient.ControllerDetails{}
	store.CurrentControllerName = "mycontroller"
	s.store = store
}

func (s *ShowSuite) setup(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.secretBackendsAPI = mocks.NewMockListSecretBackendsAPI(ctrl)

	return ctrl
}

func (s *ShowSuite) TestShowYAML(c *gc.C) {
	defer s.setup(c).Finish()

	s.secretBackendsAPI.EXPECT().ListSecretBackends([]string{"myvault"}, true).Return(
		[]apisecretbackends.SecretBackend{{
			ID:                  "vault-id",
			Name:                "myvault",
			BackendType:         "vault",
			TokenRotateInterval: ptr(666 * time.Minute),
			Config:              map[string]interface{}{"endpoint": "http://vault"},
			NumSecrets:          666,
			Status:              status.Error,
			Message:             "vault is sealed",
		}}, nil)

	s.secretBackendsAPI.EXPECT().Close().Return(nil)

	ctx, err := cmdtesting.RunCommand(c, secretbackends.NewShowCommandForTest(s.store, s.secretBackendsAPI), "myvault", "--reveal")
	c.Assert(err, jc.ErrorIsNil)
	out := cmdtesting.Stdout(ctx)
	c.Assert(out, gc.Equals, `
myvault:
  backend: vault
  token-rotate-interval: 11h6m0s
  config:
    endpoint: http://vault
  secrets: 666
  status: error
  message: vault is sealed
  id: vault-id
`[1:])
}
