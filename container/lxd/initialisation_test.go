// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

//go:build linux

package lxd

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/packaging/v2/commands"
	"github.com/juju/packaging/v2/manager"
	"github.com/juju/proxy"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/container/lxd/mocks"
	lxdtesting "github.com/juju/juju/container/lxd/testing"
	coretesting "github.com/juju/juju/testing"
)

type initialiserTestSuite struct {
	coretesting.BaseSuite
	testing.PatchExecHelper
}

// patchDF100GB ensures that df always returns 100GB.
func (s *initialiserTestSuite) patchDF100GB() {
	df100 := func(path string) (uint64, error) {
		return 100 * 1024 * 1024 * 1024, nil
	}
	s.PatchValue(&df, df100)
}

type InitialiserSuite struct {
	initialiserTestSuite
	calledCmds []string
}

var _ = gc.Suite(&InitialiserSuite{})

const lxdSnapChannel = "latest/stable"

func (s *InitialiserSuite) SetUpTest(c *gc.C) {
	coretesting.SkipLXDNotSupported(c)
	s.initialiserTestSuite.SetUpTest(c)
	s.calledCmds = []string{}
	s.PatchValue(&manager.RunCommandWithRetry, getMockRunCommandWithRetry(&s.calledCmds))

	nonRandomizedOctetRange := func() []int {
		// chosen by fair dice roll
		// guaranteed to be random :)
		// intentionally not random to allow for deterministic tests
		return []int{4, 5, 6, 7, 8}
	}
	s.PatchValue(&randomizedOctetRange, nonRandomizedOctetRange)
	// Fake the lxc executable for all the tests.
	testing.PatchExecutableAsEchoArgs(c, s, "lxc")
	testing.PatchExecutableAsEchoArgs(c, s, "lxd")
}

// getMockRunCommandWithRetry is a helper function which returns a function
// with an identical signature to manager.RunCommandWithRetry which saves each
// command it receives in a slice and always returns no output, error code 0
// and a nil error.
func getMockRunCommandWithRetry(calledCmds *[]string) func(string, manager.Retryable, manager.RetryPolicy) (string, int, error) {
	return func(cmd string, _ manager.Retryable, _ manager.RetryPolicy) (string, int, error) {
		*calledCmds = append(*calledCmds, cmd)
		return "", 0, nil
	}
}

func (s *initialiserTestSuite) containerInitialiser(svr lxd.InstanceServer, lxdIsRunning bool, containerNetworkingMethod string) *containerInitialiser {
	result := NewContainerInitialiser(lxdSnapChannel, containerNetworkingMethod).(*containerInitialiser)
	result.configureLxdProxies = func(proxy.Settings, func() (bool, error), func() (*Server, error)) error { return nil }
	result.newLocalServer = func() (*Server, error) { return NewServer(svr) }
	result.isRunningLocally = func() (bool, error) {
		return lxdIsRunning, nil
	}
	return result
}

func (s *InitialiserSuite) TestSnapInstalledNoAptInstall(c *gc.C) {
	PatchLXDViaSnap(s, true)
	PatchHostSeries(s, "cosmic")

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mgr := mocks.NewMockSnapManager(ctrl)
	mgr.EXPECT().InstalledChannel("lxd").Return("latest/stable")
	PatchGetSnapManager(s, mgr)

	err := s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(s.calledCmds, gc.DeepEquals, []string{})
}

func (s *InitialiserSuite) TestSnapChannelMismatch(c *gc.C) {
	PatchLXDViaSnap(s, true)
	PatchHostSeries(s, "focal")

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mgr := mocks.NewMockSnapManager(ctrl)
	gomock.InOrder(
		mgr.EXPECT().InstalledChannel("lxd").Return("3.2/stable"),
		mgr.EXPECT().ChangeChannel("lxd", lxdSnapChannel),
	)
	PatchGetSnapManager(s, mgr)

	err := s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *InitialiserSuite) TestSnapChannelPrefixMatch(c *gc.C) {
	PatchLXDViaSnap(s, true)
	PatchHostSeries(s, "focal")

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mgr := mocks.NewMockSnapManager(ctrl)
	gomock.InOrder(
		// The channel for the installed lxd snap also includes the
		// branch for the focal release. The "track/risk" prefix is
		// the same however so the container manager should not attempt
		// to change the channel.
		mgr.EXPECT().InstalledChannel("lxd").Return("latest/stable/ubuntu-20.04"),
	)
	PatchGetSnapManager(s, mgr)

	err := s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *InitialiserSuite) TestNoSeriesPackages(c *gc.C) {
	PatchLXDViaSnap(s, false)

	// Here we want to test for any other series whilst avoiding the
	// possibility of hitting a cloud archive-requiring release.
	// As such, we simply pass an empty series.
	PatchHostSeries(s, "")

	paccmder, err := commands.NewPackageCommander("xenial")
	c.Assert(err, jc.ErrorIsNil)

	err = s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(s.calledCmds, gc.DeepEquals, []string{
		paccmder.InstallCmd("lxd"),
	})
}

func (s *InitialiserSuite) TestInstallViaSnap(c *gc.C) {
	PatchLXDViaSnap(s, false)

	PatchHostSeries(s, "disco")

	paccmder := commands.NewSnapPackageCommander()

	err := s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(s.calledCmds, gc.DeepEquals, []string{
		paccmder.InstallCmd("--classic --channel latest/stable lxd"),
	})
}

func (s *InitialiserSuite) TestLXDInitBionicLocalNetworkingMethod(c *gc.C) {
	s.patchDF100GB()
	PatchHostSeries(s, "bionic")

	err := s.containerInitialiser(nil, true, "local").Initialise()
	c.Assert(err, jc.ErrorIsNil)

	testing.AssertEchoArgs(c, "lxd", "init", "--auto")
}

func (s *InitialiserSuite) TestLXDInitBionicFanNetworkingMethod(c *gc.C) {
	s.patchDF100GB()
	PatchHostSeries(s, "bionic")

	err := s.containerInitialiser(nil, true, "fan").Initialise()
	c.Assert(err, jc.ErrorIsNil)

	testing.AssertEchoArgs(c, "lxd", "init", "--preseed")
}

func (s *InitialiserSuite) TestLXDAlreadyInitialized(c *gc.C) {
	s.patchDF100GB()
	PatchHostSeries(s, "bionic")

	ci := s.containerInitialiser(nil, true, "local")
	ci.getExecCommand = s.PatchExecHelper.GetExecCommand(testing.PatchExecConfig{
		Stderr:   `error: You have existing containers or images. lxd init requires an empty LXD.`,
		ExitCode: 1,
	})

	// the above error should be ignored by the code that calls lxd init.
	err := ci.Initialise()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *InitialiserSuite) TestInitializeSetsProxies(c *gc.C) {
	PatchHostSeries(s, "")

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	cSvr := lxdtesting.NewMockInstanceServer(ctrl)

	s.PatchEnvironment("http_proxy", "http://test.local/http/proxy")
	s.PatchEnvironment("https_proxy", "http://test.local/https/proxy")
	s.PatchEnvironment("no_proxy", "test.local,localhost")

	updateReq := api.ServerPut{Config: map[string]interface{}{
		"core.proxy_http":         "http://test.local/http/proxy",
		"core.proxy_https":        "http://test.local/https/proxy",
		"core.proxy_ignore_hosts": "test.local,localhost",
	}}
	gomock.InOrder(
		cSvr.EXPECT().GetServer().Return(&api.Server{}, lxdtesting.ETag, nil).Times(2),
		cSvr.EXPECT().UpdateServer(updateReq, lxdtesting.ETag).Return(nil),
	)

	ci := s.containerInitialiser(cSvr, true, "local")
	ci.configureLxdProxies = internalConfigureLXDProxies
	err := ci.Initialise()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *InitialiserSuite) TestConfigureProxiesLXDNotRunning(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	cSvr := lxdtesting.NewMockInstanceServer(ctrl)

	s.PatchEnvironment("http_proxy", "http://test.local/http/proxy")
	s.PatchEnvironment("https_proxy", "http://test.local/https/proxy")
	s.PatchEnvironment("no_proxy", "test.local,localhost")

	// No expected calls.
	ci := s.containerInitialiser(cSvr, false, "local")
	err := ci.Initialise()
	c.Assert(err, jc.ErrorIsNil)
}

type ConfigureInitialiserSuite struct {
	initialiserTestSuite
	testing.PatchExecHelper
}

var _ = gc.Suite(&ConfigureInitialiserSuite{})

func (s *ConfigureInitialiserSuite) SetUpTest(c *gc.C) {
	s.initialiserTestSuite.SetUpTest(c)
	// Fake the lxc executable for all the tests.
	testing.PatchExecutableAsEchoArgs(c, s, "lxc")
	testing.PatchExecutableAsEchoArgs(c, s, "lxd")
}
