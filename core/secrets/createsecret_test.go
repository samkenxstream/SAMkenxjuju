// Copyright 2021 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package secrets_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/core/secrets"
)

type CreateSecretSuite struct{}

var _ = gc.Suite(&CreateSecretSuite{})

func (s *CreateSecretSuite) TestBadKey(c *gc.C) {
	_, err := secrets.CreateSecretData([]string{"fo=bar"})
	c.Assert(err, gc.ErrorMatches, `key "fo" not valid`)
}

func (s *CreateSecretSuite) TestKeyValues(c *gc.C) {
	data, err := secrets.CreateSecretData([]string{"foo=bar", "hello=world", "goodbye#base64=world"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(data, jc.DeepEquals, secrets.SecretData{
		"foo":     "YmFy",
		"hello":   "d29ybGQ=",
		"goodbye": "world",
	})
}

func (s *CreateSecretSuite) TestKeyContentTooLarge(c *gc.C) {
	content := strings.Repeat("a", 8*1024)
	_, err := secrets.CreateSecretData([]string{"foo=" + content})
	c.Assert(err, gc.ErrorMatches, `secret content for key "foo" too large: 8192 bytes`)
}

func (s *CreateSecretSuite) TestTotalContentTooLarge(c *gc.C) {
	content := strings.Repeat("a", 4*1024)
	var args []string
	for i := 1; i <= 20; i++ {
		args = append(args, fmt.Sprintf("key%d=%s", i, content))
	}
	_, err := secrets.CreateSecretData(args)
	c.Assert(err, gc.ErrorMatches, `secret content too large: 81920 bytes`)
}

func (s *CreateSecretSuite) TestSecretKeyFromFile(c *gc.C) {
	content := `
      -----BEGIN CERTIFICATE-----
      MIIFYjCCA0qgAwIBAgIQKaPND9YggIG6+jOcgmpk3DANBgkqhkiG9w0BAQsFADAz
      MRwwGgYDVQQKExNsaW51eGNvbnRhaW5lcnMub3JnMRMwEQYDVQQDDAp0aW1AZWx3
      -----END CERTIFICATE-----`[1:]

	dir := c.MkDir()
	fileName := filepath.Join(dir, "secret-data.bin")
	err := os.WriteFile(fileName, []byte(content), os.FileMode(0644))
	c.Assert(err, jc.ErrorIsNil)

	data, err := secrets.CreateSecretData([]string{"key1=value1", "key2#file=" + fileName})
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(data, jc.DeepEquals, secrets.SecretData{
		"key1": "dmFsdWUx",
		"key2": `ICAgICAgLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCiAgICAgIE1JSUZZakNDQTBxZ0F3SUJBZ0lRS2FQTkQ5WWdnSUc2K2pPY2dtcGszREFOQmdrcWhraUc5dzBCQVFzRkFEQXoKICAgICAgTVJ3d0dnWURWUVFLRXhOc2FXNTFlR052Ym5SaGFXNWxjbk11YjNKbk1STXdFUVlEVlFRRERBcDBhVzFBWld4MwogICAgICAtLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t`,
	})
}

func (s *CreateSecretSuite) TestYAMLFile(c *gc.C) {
	data := `
    hello: world
    goodbye#base64: world
    another-key: !!binary |
      R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5
      OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/+
      +f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLC
      AgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=`

	dir := c.MkDir()
	fileName := filepath.Join(dir, "secret.yaml")
	err := os.WriteFile(fileName, []byte(data), os.FileMode(0644))
	c.Assert(err, jc.ErrorIsNil)

	attrs, err := secrets.ReadSecretData(fileName)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attrs, jc.DeepEquals, secrets.SecretData{
		"hello":       "d29ybGQ=",
		"goodbye":     "world",
		"another-key": `R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/++f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLCAgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=`,
	})
}

func (s *CreateSecretSuite) TestJSONFile(c *gc.C) {
	data := `{
    "hello": "world",
    "goodbye#base64": "world",
}`

	dir := c.MkDir()
	fileName := filepath.Join(dir, "secret.json")
	err := os.WriteFile(fileName, []byte(data), os.FileMode(0644))
	c.Assert(err, jc.ErrorIsNil)

	attrs, err := secrets.ReadSecretData(fileName)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attrs, jc.DeepEquals, secrets.SecretData{
		"hello":   "d29ybGQ=",
		"goodbye": "world",
	})
}
