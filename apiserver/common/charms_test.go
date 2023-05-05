// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package common

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/rpc/params"
)

type charmOriginSuite struct{}

var _ = gc.Suite(&charmOriginSuite{})

func (s *charmOriginSuite) TestValidateCharmOriginSuccessCharmHub(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{
		Hash:   "myHash",
		ID:     "myID",
		Source: "charm-hub",
	})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsFalse)
}

func (s *charmOriginSuite) TestValidateCharmOriginSuccessLocal(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{Source: "local"})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsFalse)
}

func (s *charmOriginSuite) TestValidateCharmOriginNil(c *gc.C) {
	err := ValidateCharmOrigin(nil)
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsTrue)
}

func (s *charmOriginSuite) TestValidateCharmOriginNilSource(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{Source: ""})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsTrue)
}

func (s *charmOriginSuite) TestValidateCharmOriginBadSource(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{Source: "charm-store"})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsTrue)
}

func (s *charmOriginSuite) TestValidateCharmOriginCharmHubIDNoHash(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{
		ID:     "myID",
		Source: "charm-hub",
	})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsTrue)
}

func (s *charmOriginSuite) TestValidateCharmOriginCharmHubHashNoID(c *gc.C) {
	err := ValidateCharmOrigin(&params.CharmOrigin{
		Hash:   "myHash",
		Source: "charm-hub",
	})
	c.Assert(errors.Is(err, errors.BadRequest), jc.IsTrue)
}
