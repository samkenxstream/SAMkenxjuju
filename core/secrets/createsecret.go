// Copyright 2021 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package secrets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/utils/v3"
	"gopkg.in/yaml.v2"
)

var keyRegExp = regexp.MustCompile("^([a-z](?:-?[a-z0-9]){2,})$")

// SecretData holds secret key values.
type SecretData map[string]string

const (
	fileSuffix          = "#file"
	maxValueSizeBytes   = 5 * 1024
	maxContentSizeBytes = 64 * 1024
)

// CreateSecretData creates a secret data bag from a list of arguments.
// If a key has the #base64 suffix, then the value is already base64 encoded,
// otherwise the value is base64 encoded as it is added to the data bag.
func CreateSecretData(args []string) (SecretData, error) {
	data := make(SecretData)
	for _, val := range args {
		// Remove any base64 padding ("=") before splitting the key=value.
		stripped := strings.TrimRight(val, string(base64.StdPadding))
		idx := strings.Index(stripped, "=")
		if idx < 1 {
			return nil, errors.NotValidf("key value %q", val)
		}
		keyVal := []string{
			val[0:idx],
			val[idx+1:],
		}
		key := keyVal[0]
		value := keyVal[1]
		if !strings.HasSuffix(key, fileSuffix) {
			data[key] = value
			continue
		}
		key = strings.TrimSuffix(key, fileSuffix)
		path, err := utils.NormalizePath(value)
		if err != nil {
			return nil, errors.Trace(err)
		}
		fs, err := os.Stat(path)
		if err == nil && fs.Size() > maxValueSizeBytes {
			return nil, errors.Errorf("secret content in file %q too large: %d bytes", path, fs.Size())
		}
		content, err := os.ReadFile(value)
		if err != nil {
			return nil, errors.Annotatef(err, "reading content for secret key %q", key)
		}
		data[key] = string(content)
	}
	return encodeBase64(data)
}

// ReadSecretData reads secret data from a YAML or JSON file as key value pairs.
func ReadSecretData(f string) (SecretData, error) {
	attrs := make(SecretData)
	path, err := utils.NormalizePath(f)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fs, err := os.Stat(path)
	if err == nil && fs.Size() > maxContentSizeBytes {
		return nil, errors.Errorf("secret content in file %q too large: %d bytes", path, fs.Size())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err := json.Unmarshal(data, &attrs); err != nil {
		err = yaml.Unmarshal(data, &attrs)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return encodeBase64(attrs)
}

const base64Suffix = "#base64"

func encodeBase64(in SecretData) (SecretData, error) {
	out := make(SecretData, len(in))
	var contentSize int
	for k, v := range in {
		if len(v) > maxValueSizeBytes {
			return nil, errors.Errorf("secret content for key %q too large: %d bytes", k, len(v))
		}
		contentSize += len(v)
		if strings.HasSuffix(k, base64Suffix) {
			k = strings.TrimSuffix(k, base64Suffix)
			if !keyRegExp.MatchString(k) {
				return nil, errors.NotValidf("key %q", k)
			}
			out[k] = v
			continue
		}
		if !keyRegExp.MatchString(k) {
			return nil, errors.NotValidf("key %q", k)
		}
		out[k] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", v)))
	}
	if contentSize > maxContentSizeBytes {
		return nil, errors.Errorf("secret content too large: %d bytes", contentSize)
	}
	return out, nil
}
