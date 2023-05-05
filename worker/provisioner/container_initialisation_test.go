// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package provisioner

import (
	"errors"
	"sync"

	"github.com/golang/mock/gomock"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/v3"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	apiprovisioner "github.com/juju/juju/api/agent/provisioner"
	provisionermocks "github.com/juju/juju/api/agent/provisioner/mocks"
	apimocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/container"
	"github.com/juju/juju/container/factory"
	"github.com/juju/juju/container/testing"
	"github.com/juju/juju/core/instance"
	"github.com/juju/juju/core/machinelock"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/rpc/params"
	coretesting "github.com/juju/juju/testing"
	jujuversion "github.com/juju/juju/version"
)

type containerSetupSuite struct {
	coretesting.BaseSuite

	modelUUID      utils.UUID
	controllerUUID utils.UUID

	initialiser  *testing.MockInitialiser
	facadeCaller *apimocks.MockFacadeCaller
	machine      *provisionermocks.MockMachineProvisioner
	manager      *testing.MockManager

	machineLock *fakeMachineLock
}

func (s *containerSetupSuite) SetUpTest(c *gc.C) {
	s.BaseSuite.SetUpTest(c)

	s.modelUUID = utils.MustNewUUID()
	s.controllerUUID = utils.MustNewUUID()

	s.machineLock = &fakeMachineLock{}
}

var _ = gc.Suite(&containerSetupSuite{})

func (s *containerSetupSuite) TestInitialiseContainersKVM(c *gc.C) {
	s.testInitialiseContainers(c, instance.KVM)
}

func (s *containerSetupSuite) TestInitialiseContainersLXD(c *gc.C) {
	s.testInitialiseContainers(c, instance.LXD)
}

func (s *containerSetupSuite) testInitialiseContainers(c *gc.C, containerType instance.ContainerType) {
	defer s.patch(c).Finish()

	s.expectContainerManagerConfig(containerType)
	s.initialiser.EXPECT().Initialise().Return(nil)

	s.PatchValue(
		&factory.NewContainerManager,
		func(forType instance.ContainerType, conf container.ManagerConfig) (container.Manager, error) {
			return s.manager, nil
		})

	cs := s.setUpContainerSetup(c, containerType)
	abort := make(chan struct{})
	close(abort)
	err := cs.initialiseContainers(abort)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *containerSetupSuite) TestInitialiseContainerProvisionerKVM(c *gc.C) {
	s.testInitialiseContainers(c, instance.KVM)
}

func (s *containerSetupSuite) TestInitialiseContainerProvisonerLXD(c *gc.C) {
	s.testInitialiseContainers(c, instance.LXD)
}

func (s *containerSetupSuite) TestContainerManagerConfigError(c *gc.C) {
	defer s.patch(c).Finish()

	s.facadeCaller.EXPECT().FacadeCall(
		"ContainerManagerConfig", params.ContainerManagerConfigParams{Type: "lxd"}, gomock.Any()).Return(
		errors.New("boom"))

	cs := s.setUpContainerSetup(c, instance.LXD)
	abort := make(chan struct{})
	close(abort)
	err := cs.initialiseContainers(abort)
	c.Assert(err, gc.ErrorMatches, ".*generating container manager config: boom")
}

func (s *containerSetupSuite) setUpContainerSetup(c *gc.C, containerType instance.ContainerType) *ContainerSetup {
	pState := apiprovisioner.NewStateFromFacade(s.facadeCaller)

	cfg, err := agent.NewAgentConfig(
		agent.AgentConfigParams{
			Paths:             agent.DefaultPaths,
			Tag:               s.machine.MachineTag(),
			UpgradedToVersion: jujuversion.Current,
			Password:          "password",
			Nonce:             "nonce",
			APIAddresses:      []string{"0.0.0.0:12345"},
			CACert:            coretesting.CACert,
			Controller:        names.NewControllerTag(s.controllerUUID.String()),
			Model:             names.NewModelTag(s.modelUUID.String()),
		})
	c.Assert(err, jc.ErrorIsNil)

	args := ContainerSetupParams{
		Logger:        &noOpLogger{},
		ContainerType: containerType,
		MachineZone:   s.machine,
		MTag:          s.machine.MachineTag(),
		Provisioner:   pState,
		Config:        cfg,
		MachineLock:   s.machineLock,
		CredentialAPI: &credentialAPIForTest{},
		GetNetConfig: func(_ network.ConfigSource) ([]params.NetworkConfig, error) {
			return nil, nil
		},
	}

	return NewContainerSetup(args)
}

func (s *containerSetupSuite) patch(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	s.initialiser = testing.NewMockInitialiser(ctrl)
	s.facadeCaller = apimocks.NewMockFacadeCaller(ctrl)
	s.machine = provisionermocks.NewMockMachineProvisioner(ctrl)
	s.manager = testing.NewMockManager(ctrl)

	s.machine.EXPECT().MachineTag().Return(names.NewMachineTag("0")).AnyTimes()

	s.PatchValue(GetContainerInitialiser, func(instance.ContainerType, map[string]string, string) container.Initialiser {
		return s.initialiser
	})

	return ctrl
}

// expectContainerManagerConfig sets up expectations associated with
// acquisition and decoration of container manager configuration.
func (s *containerSetupSuite) expectContainerManagerConfig(cType instance.ContainerType) {
	resultSource := params.ContainerManagerConfig{
		ManagerConfig: map[string]string{"model-uuid": s.modelUUID.String()},
	}
	s.facadeCaller.EXPECT().FacadeCall(
		"ContainerManagerConfig", params.ContainerManagerConfigParams{Type: cType}, gomock.Any(),
	).SetArg(2, resultSource).MinTimes(1)
}

type credentialAPIForTest struct{}

func (*credentialAPIForTest) InvalidateModelCredential(reason string) error {
	return nil
}

type fakeMachineLock struct {
	mu sync.Mutex
}

func (f *fakeMachineLock) Acquire(spec machinelock.Spec) (func(), error) {
	f.mu.Lock()
	return func() {
		f.mu.Unlock()
	}, nil
}

func (f *fakeMachineLock) Report(opts ...machinelock.ReportOption) (string, error) {
	return "", nil
}

type noOpLogger struct {
}

func (noOpLogger) Errorf(format string, values ...interface{})   {}
func (noOpLogger) Warningf(format string, values ...interface{}) {}
func (noOpLogger) Infof(format string, values ...interface{})    {}
func (noOpLogger) Debugf(format string, values ...interface{})   {}
func (noOpLogger) Tracef(format string, values ...interface{})   {}
func (noOpLogger) IsTraceEnabled() bool                          { return false }
