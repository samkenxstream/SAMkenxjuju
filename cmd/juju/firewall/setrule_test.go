// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package firewall_test

import (
	"github.com/juju/cmd/v3"
	"github.com/juju/cmd/v3/cmdtesting"
	"github.com/juju/errors"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/cmd/juju/firewall"
)

type SetRuleSuite struct {
	testing.BaseSuite

	mockAPI *mockSetRuleAPI
}

var _ = gc.Suite(&SetRuleSuite{})

func (s *SetRuleSuite) SetUpTest(c *gc.C) {
	s.mockAPI = &mockSetRuleAPI{}
}

func (s *SetRuleSuite) TestInitMissingService(c *gc.C) {
	_, err := s.runSetRule(c, "--allowlist", "10.0.0.0/8")
	c.Assert(err, gc.ErrorMatches, "no well known service specified")
}

func (s *SetRuleSuite) TestInitMissingWhitelist(c *gc.C) {
	_, err := s.runSetRule(c, "ssh")
	c.Assert(err, gc.ErrorMatches, `no allowlist subnets specified`)
}

func (s *SetRuleSuite) TestSetRuleSSH(c *gc.C) {
	_, err := s.runSetRule(c, "--allowlist", "10.2.1.0/8,192.168.1.0/8", "ssh")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.sshRule, gc.Equals, "10.2.1.0/8,192.168.1.0/8")
}

func (s *SetRuleSuite) TestSetRuleSAAS(c *gc.C) {
	_, err := s.runSetRule(c, "--allowlist", "10.2.1.0/8,192.168.1.0/8", "juju-application-offer")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.saasRule, gc.Equals, "10.2.1.0/8,192.168.1.0/8")
}

func (s *SetRuleSuite) TestWhitelistAndAllowlist(c *gc.C) {
	_, err := s.runSetRule(c, "ssh", "--allowlist", "192.168.0.0/24", "--whitelist", "192.168.1.0/24")
	c.Assert(err, gc.ErrorMatches, "cannot specify both whitelist and allowlist")
}

func (s *SetRuleSuite) TestSetError(c *gc.C) {
	s.mockAPI.err = errors.New("fail")
	_, err := s.runSetRule(c, "ssh", "--allowlist", "10.0.0.0/8")
	c.Assert(err, gc.ErrorMatches, ".*fail.*")
}

func (s *SetRuleSuite) runSetRule(c *gc.C, args ...string) (*cmd.Context, error) {
	return cmdtesting.RunCommand(c, firewall.NewSetRulesCommandForTest(s.mockAPI), args...)
}

type mockSetRuleAPI struct {
	sshRule  string
	saasRule string
	err      error
}

func (s *mockSetRuleAPI) Close() error {
	return nil
}

func (s *mockSetRuleAPI) ModelSet(cfg map[string]interface{}) error {
	if s.err != nil {
		return s.err
	}
	sshRule, ok := cfg[config.SSHAllowKey].(string)
	if ok {
		s.sshRule = sshRule
	}
	saasRule, ok := cfg[config.SAASIngressAllowKey].(string)
	if ok {
		s.saasRule = saasRule
	}

	return nil
}
