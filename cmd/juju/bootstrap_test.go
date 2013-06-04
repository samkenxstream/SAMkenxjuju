// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"bytes"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/dummy"
	envtesting "launchpad.net/juju-core/environs/testing"
	"launchpad.net/juju-core/environs/tools"
	"launchpad.net/juju-core/testing"
	"launchpad.net/juju-core/version"
	"strings"
)

type BootstrapSuite struct {
	testing.LoggingSuite
	testing.MgoSuite
}

var _ = Suite(&BootstrapSuite{})

func (s *BootstrapSuite) SetUpSuite(c *C) {
	s.LoggingSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *BootstrapSuite) SetUpTest(c *C) {
	s.LoggingSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
}

func (s *BootstrapSuite) TearDownSuite(c *C) {
	s.MgoSuite.TearDownSuite(c)
	s.LoggingSuite.TearDownSuite(c)
}

func (s *BootstrapSuite) TearDownTest(c *C) {
	s.MgoSuite.TearDownTest(c)
	s.LoggingSuite.TearDownTest(c)
	dummy.Reset()
}

func (*BootstrapSuite) TestMissingEnvironment(c *C) {
	defer testing.MakeFakeHomeNoEnvironments(c, "empty").Restore()
	ctx := testing.Context(c)
	code := cmd.Main(&BootstrapCommand{}, ctx, nil)
	c.Check(code, Equals, 1)
	errStr := ctx.Stderr.(*bytes.Buffer).String()
	strippedErr := strings.Replace(errStr, "\n", "", -1)
	c.Assert(strippedErr, Matches, ".*No juju environment configuration file exists.*")
}

func (s *BootstrapSuite) TestTests(c *C) {
	uploadTools = mockUploadTools
	defer func() { uploadTools = tools.Upload }()
	for i, test := range bootstrapTests {
		c.Logf("\ntest %d: %s", i, test.info)
		test.run(c)
	}
}

type bootstrapTest struct {
	info string
	// binary version string used to set version.Current
	version string
	args    []string
	err     string
	// binary version strings for expected tools; if set, no default tools
	// will be uploaded before running the test.
	uploads     []string
	constraints constraints.Value
}

func (test bootstrapTest) run(c *C) {
	defer testing.MakeFakeHome(c, envConfig).Restore()
	dummy.Reset()
	env, err := environs.NewFromName("peckham")
	c.Assert(err, IsNil)
	envtesting.RemoveAllTools(c, env)

	if test.version != "" {
		origVersion := version.Current
		version.Current = version.MustParseBinary(test.version)
		defer func() { version.Current = origVersion }()
	}
	uploadCount := len(test.uploads)
	if uploadCount == 0 {
		usefulVersion := version.Current
		usefulVersion.Series = env.Config().DefaultSeries()
		envtesting.UploadFakeToolsVersion(c, env.Storage(), usefulVersion)
	}

	// Run command and check for uploads.
	opc, errc := runCommand(new(BootstrapCommand), test.args...)
	if uploadCount > 0 {
		for i := 0; i < uploadCount; i++ {
			c.Check((<-opc).(dummy.OpPutFile).Env, Equals, "peckham")
		}
		list, err := environs.FindAvailableTools(env, version.Current.Major)
		c.Check(err, IsNil)
		c.Logf("found: " + list.String())
		urls := list.URLs()
		c.Check(urls, HasLen, len(test.uploads))
		for _, v := range test.uploads {
			c.Logf("seeking: " + v)
			vers := version.MustParseBinary(v)
			_, found := urls[vers]
			c.Check(found, Equals, true)
		}
	}

	// Check for remaining operations/errors.
	if test.err != "" {
		c.Check(<-errc, ErrorMatches, test.err)
		return
	}
	if !c.Check(<-errc, IsNil) {
		return
	}
	opPutBootstrapVerifyFile := (<-opc).(dummy.OpPutFile)
	c.Check(opPutBootstrapVerifyFile.Env, Equals, "peckham")

	opBootstrap := (<-opc).(dummy.OpBootstrap)
	c.Check(opBootstrap.Env, Equals, "peckham")
	c.Check(opBootstrap.Constraints, DeepEquals, test.constraints)

	// Check a CA cert/key was generated by reloading the environment.
	env, err = environs.NewFromName("peckham")
	c.Assert(err, IsNil)
	_, hasCert := env.Config().CACert()
	c.Check(hasCert, Equals, true)
	_, hasKey := env.Config().CAPrivateKey()
	c.Check(hasKey, Equals, true)
}

var bootstrapTests = []bootstrapTest{{
	info: "no args, no error, no uploads, no constraints",
}, {
	info: "bad arg",
	args: []string{"twiddle"},
	err:  `unrecognized args: \["twiddle"\]`,
}, {
	info: "bad --constraints",
	args: []string{"--constraints", "bad=wrong"},
	err:  `invalid value "bad=wrong" for flag --constraints: unknown constraint "bad"`,
}, {
	info: "bad --series",
	args: []string{"--series", "bad1"},
	err:  `invalid value "bad1" for flag --series: invalid series name "bad1"`,
}, {
	info: "lonely --series",
	args: []string{"--series", "fine"},
	err:  `--series requires --upload-tools`,
}, {
	info: "bad environment",
	args: []string{"-e", "brokenenv"},
	err:  `environment configuration has no admin-secret`,
}, {
	info:        "constraints",
	args:        []string{"--constraints", "mem=4G cpu-cores=4"},
	constraints: constraints.MustParse("mem=4G cpu-cores=4"),
}, {
	info:    "--upload-tools picks all reasonable series",
	version: "1.2.3-hostseries-hostarch",
	args:    []string{"--upload-tools"},
	uploads: []string{
		"1.2.3.1-hostseries-hostarch",    // from version.Current
		"1.2.3.1-defaultseries-hostarch", // from env.Config().DefaultSeries()
		"1.2.3.1-precise-hostarch",       // from environs/config.DefaultSeries
	},
}, {
	info:    "--upload-tools only uploads each file once",
	version: "1.2.3-precise-hostarch",
	args:    []string{"--upload-tools"},
	uploads: []string{
		"1.2.3.1-defaultseries-hostarch",
		"1.2.3.1-precise-hostarch",
	},
}, {
	info:    "--upload-tools accepts specific series even if they're crazy",
	version: "1.2.3-hostseries-hostarch",
	args:    []string{"--upload-tools", "--series", "ping,ping,pong"},
	uploads: []string{
		"1.2.3.1-hostseries-hostarch",
		"1.2.3.1-ping-hostarch",
		"1.2.3.1-pong-hostarch",
	},
	err: "no matching tools available",
}, {
	info:    "--upload-tools always bumps build number",
	version: "1.2.3.4-defaultseries-hostarch",
	args:    []string{"--upload-tools"},
	uploads: []string{
		"1.2.3.5-defaultseries-hostarch",
		"1.2.3.5-precise-hostarch",
	},
}}
