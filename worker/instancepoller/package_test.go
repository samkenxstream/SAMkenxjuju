// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package instancepoller

import (
	"testing"

	gc "gopkg.in/check.v1"
)

//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/mocks_watcher.go github.com/juju/juju/core/watcher StringsWatcher
//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/mocks_instances.go github.com/juju/juju/environs/instances Instance
//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/mocks_cred_api.go github.com/juju/juju/worker/common CredentialAPI
//go:generate go run github.com/golang/mock/mockgen -package mocks -destination mocks/mocks_instancepoller.go github.com/juju/juju/worker/instancepoller Environ,Machine

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}
