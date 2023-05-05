// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package common

import (
	jujuos "github.com/juju/juju/core/os"
)

// MinRootDiskSizeGiB is the minimum size for the root disk of an
// instance, in Gigabytes. This value accommodates the anticipated
// size of the initial image, any updates, and future application
// data.
func MinRootDiskSizeGiB(os jujuos.OSType) uint64 {
	switch os {
	case jujuos.Ubuntu, jujuos.CentOS, jujuos.OpenSUSE:
		return 8
	// By default we just return a "sane" default, since the error will just
	// be returned by the api and seen in juju status
	default:
		return 8
	}
}

// MiBToGiB converts the provided megabytes (base-2) into the nearest
// gigabytes (base-2), rounding up. This is useful for providers that
// deal in gigabytes (while juju deals in megabytes).
func MiBToGiB(m uint64) uint64 {
	return (m + 1023) / 1024
}
