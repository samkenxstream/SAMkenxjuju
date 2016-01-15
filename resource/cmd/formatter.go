// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd

import (
	"fmt"

	charmresource "gopkg.in/juju/charm.v6-unstable/resource"

	"github.com/juju/juju/resource"
)

type charmResourcesFormatter struct {
	resources []charmresource.Resource
}

func newCharmResourcesFormatter(resources []charmresource.Resource) *charmResourcesFormatter {
	// Note that unlike the "juju status" code, we don't worry
	// about "compatVersion".
	crf := charmResourcesFormatter{
		resources: resources,
	}
	return &crf
}

func (crf *charmResourcesFormatter) format() []FormattedCharmResource {
	if crf.resources == nil {
		return nil
	}

	var formatted []FormattedCharmResource
	for _, res := range crf.resources {
		formatted = append(formatted, FormatCharmResource(res))
	}
	return formatted
}

// FormatCharmResource converts the resource info into a FormattedCharmResource.
func FormatCharmResource(res charmresource.Resource) FormattedCharmResource {
	serialized := resource.SerializeCharmResource(res)
	return FormattedCharmResource{
		Name:        serialized.Name,
		Type:        serialized.Type,
		Path:        serialized.Path,
		Comment:     serialized.Comment,
		Revision:    serialized.Revision,
		Origin:      serialized.Origin,
		Fingerprint: res.Fingerprint.String(), // ...the hex string.

		charmRevision: charmRevision(res),
	}
}

// FormatSvcResource converts the resource info into a FormattedServiceResource.
func FormatSvcResource(res resource.Resource) FormattedSvcResource {
	serialized := resource.Serialize(res)
	used := !res.IsPlaceholder()
	return FormattedSvcResource{
		Name:        serialized.Name,
		Type:        serialized.Type,
		Path:        serialized.Path,
		Comment:     serialized.Comment,
		Revision:    serialized.Revision,
		Origin:      serialized.Origin,
		Fingerprint: res.Fingerprint.String(),
		Used:        used,
		Timestamp:   serialized.Timestamp,
		Username:    serialized.Username,

		combinedRevision: combinedRevision(res),
		combinedOrigin:   combinedOrigin(used, res),
		usedYesNo:        usedYesNo(used),
	}
}

func combinedRevision(r resource.Resource) string {
	switch r.Origin {
	case charmresource.OriginStore:
		return fmt.Sprintf("%d", r.Revision)
	case charmresource.OriginUpload:
		if !r.Timestamp.IsZero() {
			return r.Timestamp.Format("2006-02-01")
		}
	}
	return "-"
}

func combinedOrigin(used bool, r resource.Resource) string {
	if r.Origin == charmresource.OriginUpload && used && r.Username != "" {
		return r.Username
	}
	return r.Origin.String()
}

func usedYesNo(used bool) string {
	if used {
		return "yes"
	}
	return "no"
}

func charmRevision(r charmresource.Resource) string {
	if r.Origin == charmresource.OriginStore {
		return fmt.Sprintf("%d", r.Revision)
	}
	return "-"
}
