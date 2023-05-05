// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package lxd

import (
	"encoding/pem"

	lxdtesting "github.com/juju/juju/container/lxd/testing"
	gc "gopkg.in/check.v1"
)

type sharedSuite struct {
	lxdtesting.BaseSuite
}

var _ = gc.Suite(&sharedSuite{})

func (sharedSuite) TestGenerateMemCert(c *gc.C) {

	cert, key, err := GenerateMemCert(false, true)
	if err != nil {
		c.Error(err)
		return
	}

	if cert == nil {
		c.Error("GenerateMemCert returned a nil cert")
		return
	}

	if key == nil {
		c.Error("GenerateMemCert returned a nil key")
		return
	}

	block, rest := pem.Decode(cert)
	if len(rest) != 0 {
		c.Errorf("GenerateMemCert returned a cert with trailing content: %q", string(rest))
	}

	if block.Type != "CERTIFICATE" {
		c.Errorf("GenerateMemCert returned a cert with Type %q not \"CERTIFICATE\"", block.Type)
	}

	block, rest = pem.Decode(key)
	if len(rest) != 0 {
		c.Errorf("GenerateMemCert returned a key with trailing content: %q", string(rest))
	}

	if block.Type != "EC PRIVATE KEY" {
		c.Errorf("GenerateMemCert returned a cert with Type %q not \"EC PRIVATE KEY\"", block.Type)
	}
}
