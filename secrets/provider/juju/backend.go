// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package juju

import (
	"context"

	"github.com/juju/errors"

	coresecrets "github.com/juju/juju/core/secrets"
)

// jujuBackend is a dummy backend which returns
// NotFound or NotSupported as needed.
type jujuBackend struct{}

// GetContent implements SecretsBackend.
func (k jujuBackend) GetContent(ctx context.Context, revisionId string) (coresecrets.SecretValue, error) {
	return nil, errors.NotFoundf("secret revision %d", revisionId)
}

// DeleteContent implements SecretsBackend.
func (k jujuBackend) DeleteContent(ctx context.Context, revisionId string) error {
	return errors.NotFoundf("secret revision %d", revisionId)
}

// SaveContent implements SecretsBackend.
func (k jujuBackend) SaveContent(ctx context.Context, uri *coresecrets.URI, revision int, value coresecrets.SecretValue) (string, error) {
	return "", errors.NotSupportedf("saving content to internal backend")
}

// Ping implements SecretsBackend.
func (k jujuBackend) Ping() error {
	return nil
}
