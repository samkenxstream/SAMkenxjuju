// Copyright 2018 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// The test cases in this file do not pertain to a specific command.

package service_test

import (
	"os"
	"path"

	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/service"
	"github.com/juju/juju/service/common"
	"github.com/juju/juju/service/mocks"
	"github.com/juju/juju/testing"
	jujuversion "github.com/juju/juju/version"
)

type agentConfSuite struct {
	testing.BaseSuite

	agentConf           agent.Config
	dataDir             string
	machineName         string
	unitNames           []string
	systemdDir          string
	systemdMultiUserDir string
	systemdDataDir      string // service.SystemdDataDir
	manager             service.SystemdServiceManager

	services []*mocks.MockService
}

func (s *agentConfSuite) SetUpSuite(c *gc.C) {
	s.BaseSuite.SetUpSuite(c)
}

func (s *agentConfSuite) SetUpTest(c *gc.C) {
	s.BaseSuite.SetUpTest(c)

	s.dataDir = c.MkDir()
	s.systemdDir = path.Join(s.dataDir, "etc", "systemd", "system")
	s.systemdMultiUserDir = path.Join(s.systemdDir, "multi-user.target.wants")
	c.Assert(os.MkdirAll(s.systemdMultiUserDir, os.ModeDir|os.ModePerm), jc.ErrorIsNil)
	s.systemdDataDir = path.Join(s.dataDir, "lib", "systemd", "system")

	s.machineName = "machine-0"
	s.unitNames = []string{"unit-ubuntu-0", "unit-mysql-0"}

	s.manager = service.NewServiceManager(
		func() bool { return true },
		s.newService,
	)

	s.assertSetupAgentsForTest(c)
	s.setUpAgentConf(c)
}

func (s *agentConfSuite) TearDownTest(c *gc.C) {
	s.services = nil
	s.BaseSuite.TearDownTest(c)
}

var _ = gc.Suite(&agentConfSuite{})

func (s *agentConfSuite) setUpAgentConf(c *gc.C) {
	configParams := agent.AgentConfigParams{
		Paths:             agent.Paths{DataDir: s.dataDir},
		Tag:               names.NewMachineTag("0"),
		UpgradedToVersion: jujuversion.Current,
		APIAddresses:      []string{"localhost:17070"},
		CACert:            testing.CACert,
		Password:          "fake",
		Controller:        testing.ControllerTag,
		Model:             testing.ModelTag,
	}

	agentConf, err := agent.NewAgentConfig(configParams)
	c.Assert(err, jc.ErrorIsNil)

	err = agentConf.Write()
	c.Assert(err, jc.ErrorIsNil)

	s.agentConf = agentConf
}

func (s *agentConfSuite) setUpServices(ctrl *gomock.Controller) {
	s.addService(ctrl, "jujud-"+s.machineName)
}

func (s *agentConfSuite) addService(ctrl *gomock.Controller, name string) {
	svc := mocks.NewMockService(ctrl)
	svc.EXPECT().Name().Return(name).AnyTimes()
	s.services = append(s.services, svc)
}

func (s *agentConfSuite) newService(name string, _ common.Conf) (service.Service, error) {
	for _, svc := range s.services {
		if svc.Name() == name {
			return svc, nil
		}
	}
	return nil, errors.NotFoundf("service %q", name)
}

func (s *agentConfSuite) assertSetupAgentsForTest(c *gc.C) {
	agentsDir := path.Join(s.dataDir, "agents")
	err := os.MkdirAll(path.Join(agentsDir, s.machineName), os.ModeDir|os.ModePerm)
	c.Assert(err, jc.ErrorIsNil)
	for _, unit := range s.unitNames {
		err = os.Mkdir(path.Join(agentsDir, unit), os.ModeDir|os.ModePerm)
		c.Assert(err, jc.ErrorIsNil)
	}
}

func (s *agentConfSuite) TestFindAgents(c *gc.C) {
	machineAgent, unitAgents, errAgents, err := s.manager.FindAgents(s.dataDir)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(machineAgent, gc.Equals, s.machineName)
	c.Assert(unitAgents, jc.SameContents, s.unitNames)
	c.Assert(errAgents, gc.HasLen, 0)
}

func (s *agentConfSuite) TestFindAgentsUnexpectedTagType(c *gc.C) {
	unexpectedAgent := names.NewApplicationTag("failme").String()
	unexpectedAgentDir := path.Join(s.dataDir, "agents", unexpectedAgent)
	err := os.MkdirAll(unexpectedAgentDir, os.ModeDir|os.ModePerm)
	c.Assert(err, jc.ErrorIsNil)

	machineAgent, unitAgents, unexpectedAgents, err := s.manager.FindAgents(s.dataDir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(machineAgent, gc.Equals, s.machineName)
	c.Assert(unitAgents, jc.SameContents, s.unitNames)
	c.Assert(unexpectedAgents, gc.DeepEquals, []string{unexpectedAgent})
}

func (s *agentConfSuite) TestCreateAgentConfDesc(c *gc.C) {
	conf, err := s.manager.CreateAgentConf("machine-2", s.dataDir)
	c.Assert(err, jc.ErrorIsNil)
	// Spot check Conf
	c.Assert(conf.Desc, gc.Equals, "juju agent for machine-2")
}

func (s *agentConfSuite) TestCreateAgentConfLogPath(c *gc.C) {
	conf, err := s.manager.CreateAgentConf("machine-2", s.dataDir)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(conf.Logfile, gc.Equals, "/var/log/juju/machine-2.log")
}

func (s *agentConfSuite) TestCreateAgentConfFailAgentKind(c *gc.C) {
	_, err := s.manager.CreateAgentConf("application-fail", s.dataDir)
	c.Assert(err, gc.ErrorMatches, `agent "application-fail" is neither a machine nor a unit`)
}

func (s *agentConfSuite) TestWriteSystemdAgent(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	s.setUpServices(ctrl)
	s.services[0].EXPECT().WriteService().Return(nil)

	err := s.manager.WriteSystemdAgent(
		s.machineName, s.systemdDataDir, s.systemdMultiUserDir)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *agentConfSuite) TestWriteSystemdAgentSystemdNotRunning(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	s.setUpServices(ctrl)
	s.services[0].EXPECT().WriteService().Return(nil)

	s.manager = service.NewServiceManager(
		func() bool { return false },
		s.newService,
	)

	err := s.manager.WriteSystemdAgent(
		s.machineName, s.systemdDataDir, s.systemdMultiUserDir)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *agentConfSuite) TestWriteSystemdAgentWriteServiceFail(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	s.setUpServices(ctrl)
	s.services[0].EXPECT().WriteService().Return(errors.New("fail me"))

	err := s.manager.WriteSystemdAgent(
		s.machineName, s.systemdDataDir, s.systemdMultiUserDir)
	c.Assert(err, gc.ErrorMatches, "fail me")
}
