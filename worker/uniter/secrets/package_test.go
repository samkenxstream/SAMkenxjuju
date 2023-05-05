// Copyright 2021 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package secrets_test

import (
	"testing"

	gc "gopkg.in/check.v1"
)

//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/tracker_mock.go github.com/juju/juju/worker/uniter/secrets SecretStateTracker
//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/client_mock.go github.com/juju/juju/worker/uniter/secrets SecretsClient

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}
