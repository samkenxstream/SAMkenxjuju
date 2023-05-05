// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package secretdrainworker

import (
	"github.com/juju/errors"
	"github.com/juju/worker/v3"
	"github.com/juju/worker/v3/dependency"

	"github.com/juju/juju/api/agent/secretsdrain"
	"github.com/juju/juju/api/agent/secretsmanager"
	"github.com/juju/juju/api/base"
	jujusecrets "github.com/juju/juju/secrets"
)

// ManifoldConfig describes the resources used by the secretdrainworker worker.
type ManifoldConfig struct {
	APICallerName string
	Logger        Logger

	NewSecretsDrainFacade func(base.APICaller) SecretsDrainFacade
	NewWorker             func(Config) (worker.Worker, error)
	NewBackendsClient     func(jujusecrets.JujuAPIClient) (jujusecrets.BackendsClient, error)
}

// NewSecretsDrainFacade returns a new SecretsDrainFacade.
func NewSecretsDrainFacade(caller base.APICaller) SecretsDrainFacade {
	return secretsdrain.NewClient(caller)
}

// NewBackendsClient returns a new secret backends client.
func NewBackendsClient(facade jujusecrets.JujuAPIClient) (jujusecrets.BackendsClient, error) {
	return jujusecrets.NewClient(facade)
}

// Manifold returns a Manifold that encapsulates the secretdrainworker worker.
func Manifold(config ManifoldConfig) dependency.Manifold {
	return dependency.Manifold{
		Inputs: []string{
			config.APICallerName,
		},
		Start: config.start,
	}
}

// Validate is called by start to check for bad configuration.
func (cfg ManifoldConfig) Validate() error {
	if cfg.APICallerName == "" {
		return errors.NotValidf("empty APICallerName")
	}
	if cfg.Logger == nil {
		return errors.NotValidf("nil Logger")
	}
	if cfg.NewSecretsDrainFacade == nil {
		return errors.NotValidf("nil NewSecretsDrainFacade")
	}
	if cfg.NewWorker == nil {
		return errors.NotValidf("nil NewWorker")
	}
	if cfg.NewBackendsClient == nil {
		return errors.NotValidf("nil NewBackendsClient")
	}
	return nil
}

// start is a StartFunc for a Worker manifold.
func (cfg ManifoldConfig) start(context dependency.Context) (worker.Worker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Trace(err)
	}

	var apiCaller base.APICaller
	if err := context.Get(cfg.APICallerName, &apiCaller); err != nil {
		return nil, errors.Trace(err)
	}

	worker, err := cfg.NewWorker(Config{
		SecretsDrainFacade: cfg.NewSecretsDrainFacade(apiCaller),
		Logger:             cfg.Logger,
		SecretsBackendGetter: func() (jujusecrets.BackendsClient, error) {
			return cfg.NewBackendsClient(secretsmanager.NewClient(apiCaller))
		},
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	return worker, nil
}
