// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package vault_test

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/secrets/provider"
	_ "github.com/juju/juju/secrets/provider/all"
	jujuvault "github.com/juju/juju/secrets/provider/vault"
)

type configSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&configSuite{})

func (s *configSuite) TestValidateConfig(c *gc.C) {
	p, err := provider.Provider(jujuvault.BackendType)
	c.Assert(err, jc.ErrorIsNil)
	configValidator, ok := p.(provider.ProviderConfig)
	c.Assert(ok, jc.IsTrue)
	for _, t := range []struct {
		cfg    map[string]interface{}
		oldCfg map[string]interface{}
		err    string
	}{{
		cfg: map[string]interface{}{},
		err: "endpoint: expected string, got nothing",
	}, {
		cfg:    map[string]interface{}{"endpoint": "newep"},
		oldCfg: map[string]interface{}{"endpoint": "oldep"},
		err:    `cannot change config "endpoint" from "oldep" to "newep"`,
	}, {
		cfg: map[string]interface{}{"endpoint": "newep", "client-cert": "aaa"},
		err: `vault config missing client key not valid`,
	}, {
		cfg: map[string]interface{}{"endpoint": "newep", "client-key": "aaa"},
		err: `vault config missing client certificate not valid`,
	}} {
		err = configValidator.ValidateConfig(t.oldCfg, t.cfg)
		c.Assert(err, gc.ErrorMatches, t.err)
	}
}
