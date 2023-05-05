// Copyright 2022 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package modelupgrader_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/mock/gomock"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/version/v2"
	gc "gopkg.in/check.v1"
	"gopkg.in/httprequest.v1"

	"github.com/juju/juju/api/client/modelupgrader"
	"github.com/juju/juju/api/client/modelupgrader/mocks"
	"github.com/juju/juju/rpc/params"
	coretesting "github.com/juju/juju/testing"
	coretools "github.com/juju/juju/tools"
)

type UpgradeModelSuite struct {
	coretesting.BaseSuite
}

var _ = gc.Suite(&UpgradeModelSuite{})

func (s *UpgradeModelSuite) TestAbortModelUpgrade(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	apiCaller := mocks.NewMockAPICallCloser(ctrl)

	gomock.InOrder(
		apiCaller.EXPECT().BestFacadeVersion("ModelUpgrader").Return(1),
		apiCaller.EXPECT().APICall(
			"ModelUpgrader", 1, "", "AbortModelUpgrade",
			params.ModelParam{
				ModelTag: coretesting.ModelTag.String(),
			}, nil,
		).Return(nil),
	)

	client := modelupgrader.NewClient(apiCaller)
	err := client.AbortModelUpgrade(coretesting.ModelTag.Id())
	c.Assert(err, jc.ErrorIsNil)
}

func (s *UpgradeModelSuite) TestUpgradeModel(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	apiCaller := mocks.NewMockAPICallCloser(ctrl)

	gomock.InOrder(
		apiCaller.EXPECT().BestFacadeVersion("ModelUpgrader").Return(1),
		apiCaller.EXPECT().APICall(
			"ModelUpgrader", 1, "", "UpgradeModel",
			params.UpgradeModelParams{
				ModelTag:            coretesting.ModelTag.String(),
				TargetVersion:       version.MustParse("2.9.1"),
				IgnoreAgentVersions: true,
				DryRun:              true,
			}, &params.UpgradeModelResult{},
		).DoAndReturn(func(objType string, facadeVersion int, id, request string, args, result interface{}) error {
			out := result.(*params.UpgradeModelResult)
			out.ChosenVersion = version.MustParse("2.9.99")
			return nil
		}),
	)

	client := modelupgrader.NewClient(apiCaller)
	chosenVersion, err := client.UpgradeModel(
		coretesting.ModelTag.Id(),
		version.MustParse("2.9.1"),
		"", true, true,
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(chosenVersion, gc.DeepEquals, version.MustParse("2.9.99"))
}

func (s *UpgradeModelSuite) TestUploadTools(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	apiCaller := mocks.NewMockAPICallCloser(ctrl)
	doer := mocks.NewMockDoer(ctrl)
	ctx := mocks.NewMockContext(ctrl)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"/tools?binaryVersion=%s",
			version.MustParseBinary("2.9.100-ubuntu-amd64"),
		), nil,
	)
	c.Assert(err, jc.ErrorIsNil)
	req.Header.Set("Content-Type", "application/x-tar-gz")
	req = req.WithContext(ctx)

	resp := &http.Response{
		Request:    req,
		StatusCode: http.StatusCreated,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"tools": [{"version": "2.9.100-ubuntu-amd64"}]}`)),
	}
	resp.Header.Set("Content-Type", "application/json")

	gomock.InOrder(
		apiCaller.EXPECT().BestFacadeVersion("ModelUpgrader").Return(1),
		apiCaller.EXPECT().HTTPClient().Return(&httprequest.Client{Doer: doer}, nil),
		apiCaller.EXPECT().Context().Return(ctx),
		doer.EXPECT().Do(req).Return(resp, nil),
	)

	client := modelupgrader.NewClient(apiCaller)

	result, err := client.UploadTools(
		nil, version.MustParseBinary("2.9.100-ubuntu-amd64"),
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, gc.DeepEquals, coretools.List{
		{Version: version.MustParseBinary("2.9.100-ubuntu-amd64")},
	})
}
