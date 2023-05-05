// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package subnets_test

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/subnets"
	"github.com/juju/juju/rpc/params"
)

// SubnetsSuite tests the client side subnets API
type SubnetsSuite struct {
}

var _ = gc.Suite(&SubnetsSuite{})

// TestNewAPISuccess checks that a new subnets API is created when passed a non-nil caller
func (s *SubnetsSuite) TestNewAPISuccess(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	apiCaller := basemocks.NewMockAPICallCloser(ctrl)
	apiCaller.EXPECT().BestFacadeVersion("Subnets").Return(4)

	api := subnets.NewAPI(apiCaller)
	c.Check(api, gc.NotNil)
}

// TestNewAPIWithNilCaller checks that a new subnets API is not created when passed a nil caller
func (s *SubnetsSuite) TestNewAPIWithNilCaller(c *gc.C) {
	panicFunc := func() { subnets.NewAPI(nil) }
	c.Assert(panicFunc, gc.PanicMatches, "caller is nil")
}

func makeListSubnetsArgs(space *names.SpaceTag, zone string) (params.SubnetsFilters, params.ListSubnetsResults) {
	expectArgs := params.SubnetsFilters{
		SpaceTag: space.String(),
		Zone:     zone,
	}
	return expectArgs, params.ListSubnetsResults{}
}

func (s *SubnetsSuite) TestListSubnetsNoResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	space := names.NewSpaceTag("foo")
	zone := "bar"
	args, results := makeListSubnetsArgs(&space, zone)
	result := new(params.ListSubnetsResults)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListSubnets", args, result).SetArg(2, results).Return(nil)
	client := subnets.NewAPIFromCaller(mockFacadeCaller)

	obtainedResults, err := client.ListSubnets(&space, zone)

	c.Assert(err, jc.ErrorIsNil)

	var expectedResults []params.Subnet
	c.Assert(obtainedResults, jc.DeepEquals, expectedResults)
}

func (s *SubnetsSuite) TestListSubnetsFails(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	space := names.NewSpaceTag("foo")
	zone := "bar"
	args, results := makeListSubnetsArgs(&space, zone)
	result := new(params.ListSubnetsResults)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListSubnets", args, result).SetArg(2, results).Return(errors.New("bang"))
	client := subnets.NewAPIFromCaller(mockFacadeCaller)

	obtainedResults, err := client.ListSubnets(&space, zone)
	c.Assert(err, gc.ErrorMatches, "bang")

	var expectedResults []params.Subnet
	c.Assert(obtainedResults, jc.DeepEquals, expectedResults)
}

func (s *SubnetsSuite) testSubnetsByCIDR(c *gc.C,
	ctrl *gomock.Controller,
	cidrs []string,
	results []params.SubnetsResult,
	err error, expectErr string,
) {
	var expectedResults params.SubnetsResults
	if results != nil {
		expectedResults.Results = results
	}
	args := params.CIDRParams{CIDRS: cidrs}

	result := new(params.SubnetsResults)
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("SubnetsByCIDR", args, result).SetArg(2, expectedResults).Return(err)
	client := subnets.NewAPIFromCaller(mockFacadeCaller)

	gotResult, gotErr := client.SubnetsByCIDR(cidrs)
	c.Assert(gotResult, jc.DeepEquals, results)

	if expectErr != "" {
		c.Assert(gotErr, gc.ErrorMatches, expectErr)
		return
	}

	if err != nil {
		c.Assert(gotErr, jc.DeepEquals, err)
	} else {
		c.Assert(gotErr, jc.ErrorIsNil)
	}
}

func (s *SubnetsSuite) TestSubnetsByCIDRWithNoCIDRs(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	var cidrs []string

	s.testSubnetsByCIDR(c, ctrl, cidrs, []params.SubnetsResult{}, nil, "")
}

func (s *SubnetsSuite) TestSubnetsByCIDRWithNoResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	cidrs := []string{"10.0.1.10/24"}

	s.testSubnetsByCIDR(c, ctrl, cidrs, []params.SubnetsResult{}, nil, "")
}

func (s *SubnetsSuite) TestSubnetsByCIDRWithResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	cidrs := []string{"10.0.1.10/24"}

	s.testSubnetsByCIDR(c, ctrl, cidrs, []params.SubnetsResult{{
		Subnets: []params.SubnetV2{{
			ID: "aaabbb",
			Subnet: params.Subnet{
				CIDR: "10.0.1.10/24",
			},
		}},
	}}, nil, "")
}
