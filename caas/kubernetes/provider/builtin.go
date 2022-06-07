// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	jujuclock "github.com/juju/clock"
	"github.com/juju/errors"

	k8s "github.com/juju/juju/caas/kubernetes"
	"github.com/juju/juju/caas/kubernetes/clientconfig"
	k8scloud "github.com/juju/juju/caas/kubernetes/cloud"
	jujucloud "github.com/juju/juju/cloud"
	"github.com/juju/juju/version"
)

func attemptMicroK8sCloud(cmdRunner CommandRunner) (jujucloud.Cloud, error) {
	microk8sConfig, err := getLocalMicroK8sConfig(cmdRunner)
	if err != nil {
		return jujucloud.Cloud{}, err
	}

	k8sCloud, err := k8scloud.CloudFromKubeConfigClusterReader(
		k8s.MicroK8sClusterName,
		bytes.NewReader(microk8sConfig),
		k8scloud.CloudParamaters{
			Description: jujucloud.DefaultCloudDescription(jujucloud.CloudTypeKubernetes),
			Name:        k8s.K8sCloudMicrok8s,
			Regions: []jujucloud.Region{{
				Name: k8s.Microk8sRegion,
			}},
		},
	)

	return k8sCloud, err
}

func attemptMicroK8sCredential(cmdRunner CommandRunner) (jujucloud.Credential, error) {
	microk8sConfig, err := getLocalMicroK8sConfig(cmdRunner)
	if err != nil {
		return jujucloud.Credential{}, err
	}

	k8sConfig, err := k8scloud.ConfigFromReader(bytes.NewReader(microk8sConfig))
	if err != nil {
		return jujucloud.Credential{}, errors.Annotate(err, "processing microk8s config to make juju credentials")
	}

	contextName, err := k8scloud.PickContextByClusterName(k8sConfig, k8s.MicroK8sClusterName)
	if err != nil {
		return jujucloud.Credential{}, errors.Trace(err)
	}

	context := k8sConfig.Contexts[contextName]
	resolver := clientconfig.GetJujuAdminServiceAccountResolver(jujuclock.WallClock)
	conf, err := resolver(k8s.K8sCloudMicrok8s, k8sConfig, contextName)
	if err != nil {
		return jujucloud.Credential{}, errors.Annotate(err, "resolving microk8s credentials")
	}

	return k8scloud.CredentialFromKubeConfig(context.AuthInfo, conf)
}

func getLocalMicroK8sConfig(cmdRunner CommandRunner) ([]byte, error) {
	_, err := cmdRunner.LookPath("microk8s")
	if err != nil {
		return []byte{}, errors.NotFoundf("microk8s")
	}
	notSupportErr := errors.NewNotSupported(nil, fmt.Sprintf("juju %q can only work with strict confined microk8s", version.Current))
	snapDataPath := os.Getenv("SNAP_DATA")
	if snapDataPath == "" {
		return nil, errors.Annotate(notSupportErr, "SNAP_DATA is empty")
	}
	clientConfigPath := filepath.Join(snapDataPath, "credentials", "client.config")
	content, err := ioutil.ReadFile(clientConfigPath)
	if os.IsNotExist(err) {
		return nil, errors.Annotatef(notSupportErr, "%q does not exist", clientConfigPath)
	}
	if err != nil {
		return nil, errors.Annotatef(err, "cannot read %q", clientConfigPath)
	}
	return content, nil
}
