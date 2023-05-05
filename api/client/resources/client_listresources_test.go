// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package resources_test

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/errors"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/resources"
	coreresources "github.com/juju/juju/core/resources"
	"github.com/juju/juju/rpc/params"
)

var _ = gc.Suite(&ListResourcesSuite{})

type ListResourcesSuite struct{}

func (s *ListResourcesSuite) TestListResources(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := &params.ListResourcesArgs{[]params.Entity{{
		Tag: "application-a-application",
	}, {
		Tag: "application-other-application",
	}}}
	expected1, apiResult1 := newResourceResult(c, "spam")
	expected2, apiResult2 := newResourceResult(c, "eggs", "ham")
	result := new(params.ResourcesResults)
	results := params.ResourcesResults{
		Results: []params.ResourcesResult{apiResult1, apiResult2},
	}

	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListResources", args, result).SetArg(2, results).Return(nil)
	client := resources.NewClientFromCaller(mockFacadeCaller)

	res, err := client.ListResources([]string{"a-application", "other-application"})
	c.Assert(err, jc.ErrorIsNil)
	c.Check(res, jc.DeepEquals, []coreresources.ApplicationResources{
		{Resources: expected1},
		{Resources: expected2},
	})
}

func (s *ListResourcesSuite) TestBadApplication(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	client := resources.NewClientFromCaller(mockFacadeCaller)
	_, err := client.ListResources([]string{"???"})
	c.Check(err, gc.ErrorMatches, `.*invalid application.*`)
}

func (s *ListResourcesSuite) TestEmptyResources(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := &params.ListResourcesArgs{[]params.Entity{{
		Tag: "application-a-application",
	}, {
		Tag: "application-other-application",
	}}}
	result := new(params.ResourcesResults)
	results := params.ResourcesResults{
		Results: []params.ResourcesResult{{}, {}},
	}
	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListResources", args, result).SetArg(2, results).Return(nil)
	client := resources.NewClientFromCaller(mockFacadeCaller)

	res, err := client.ListResources([]string{"a-application", "other-application"})
	c.Assert(err, jc.ErrorIsNil)
	c.Check(res, jc.DeepEquals, []coreresources.ApplicationResources{{}, {}})
}

func (s *ListResourcesSuite) TestServerError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := &params.ListResourcesArgs{[]params.Entity{{
		Tag: "application-a-application",
	}}}
	result := new(params.ResourcesResults)
	results := params.ResourcesResults{
		Results: []params.ResourcesResult{{}},
	}
	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListResources", args, result).SetArg(2, results).Return(errors.New("boom"))
	client := resources.NewClientFromCaller(mockFacadeCaller)

	_, err := client.ListResources([]string{"a-application"})
	c.Assert(err, gc.ErrorMatches, "boom")
}

func (s *ListResourcesSuite) TestArity(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := &params.ListResourcesArgs{[]params.Entity{{
		Tag: "application-a-application",
	}, {
		Tag: "application-other-application",
	}}}
	result := new(params.ResourcesResults)
	results := params.ResourcesResults{
		Results: []params.ResourcesResult{{}},
	}
	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListResources", args, result).SetArg(2, results).Return(nil)
	client := resources.NewClientFromCaller(mockFacadeCaller)

	_, err := client.ListResources([]string{"a-application", "other-application"})
	c.Assert(err, gc.ErrorMatches, "expected 2 results, got 1")
}

func (s *ListResourcesSuite) TestConversionFailed(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := &params.ListResourcesArgs{[]params.Entity{{
		Tag: "application-a-application",
	}}}
	result := new(params.ResourcesResults)
	results := params.ResourcesResults{
		Results: []params.ResourcesResult{{
			ErrorResult: params.ErrorResult{Error: &params.Error{Message: "boom"}},
		}},
	}
	mockFacadeCaller := mocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListResources", args, result).SetArg(2, results).Return(nil)
	client := resources.NewClientFromCaller(mockFacadeCaller)

	_, err := client.ListResources([]string{"a-application"})
	c.Assert(err, gc.ErrorMatches, "boom")
}
