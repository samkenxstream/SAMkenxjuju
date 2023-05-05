// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package applicationoffers_test

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/applicationoffers"
	"github.com/juju/juju/rpc/params"
)

type accessSuite struct {
}

var _ = gc.Suite(&accessSuite{})

const (
	someOffer = "user/prod.hosted-mysql"
)

func accessCall(client *applicationoffers.Client, action params.OfferAction, user, access string, offerURLs ...string) error {
	switch action {
	case params.GrantOfferAccess:
		return client.GrantOffer(user, access, offerURLs...)
	case params.RevokeOfferAccess:
		return client.RevokeOffer(user, access, offerURLs...)
	default:
		panic(action)
	}
}

func (s *accessSuite) TestGrantOfferReadOnlyUser(c *gc.C) {
	s.readOnlyUser(c, params.GrantOfferAccess)
}

func (s *accessSuite) TestRevokeOfferReadOnlyUser(c *gc.C) {
	s.readOnlyUser(c, params.RevokeOfferAccess)
}

func (s *accessSuite) readOnlyUser(c *gc.C, action params.OfferAction) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyOfferAccessRequest{
		Changes: []params.ModifyOfferAccess{
			{
				UserTag:  names.NewUserTag("bob").String(),
				Action:   action,
				Access:   params.OfferReadAccess,
				OfferURL: someOffer,
			},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{Results: []params.ErrorResult{{Error: nil}}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyOfferAccess", args, res).SetArg(2, ress).Return(nil)

	client := applicationoffers.NewClientFromCaller(mockFacadeCaller)

	err := accessCall(client, action, "bob", "read", someOffer)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *accessSuite) TestGrantOfferAdminUser(c *gc.C) {
	s.adminUser(c, params.GrantOfferAccess)
}

func (s *accessSuite) TestRevokeOfferAdminUser(c *gc.C) {
	s.adminUser(c, params.RevokeOfferAccess)
}

func (s *accessSuite) adminUser(c *gc.C, action params.OfferAction) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyOfferAccessRequest{
		Changes: []params.ModifyOfferAccess{
			{
				UserTag:  names.NewUserTag("bob").String(),
				Action:   action,
				Access:   params.OfferConsumeAccess,
				OfferURL: someOffer,
			},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{Results: []params.ErrorResult{{Error: nil}}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyOfferAccess", args, res).SetArg(2, ress).Return(nil)

	client := applicationoffers.NewClientFromCaller(mockFacadeCaller)
	err := accessCall(client, action, "bob", "consume", someOffer)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *accessSuite) TestGrantThreeOffers(c *gc.C) {
	s.threeOffers(c, params.GrantOfferAccess)
}

func (s *accessSuite) TestRevokeThreeOffers(c *gc.C) {
	s.threeOffers(c, params.RevokeOfferAccess)
}

func (s *accessSuite) threeOffers(c *gc.C, action params.OfferAction) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyOfferAccessRequest{
		Changes: []params.ModifyOfferAccess{
			{
				UserTag:  names.NewUserTag("carol").String(),
				Action:   action,
				Access:   params.OfferReadAccess,
				OfferURL: someOffer,
			},
			{
				UserTag:  names.NewUserTag("carol").String(),
				Action:   action,
				Access:   params.OfferReadAccess,
				OfferURL: someOffer,
			},
			{
				UserTag:  names.NewUserTag("carol").String(),
				Action:   action,
				Access:   params.OfferReadAccess,
				OfferURL: someOffer,
			},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{Results: []params.ErrorResult{{Error: nil}, {Error: nil}, {Error: nil}}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyOfferAccess", args, res).SetArg(2, ress).Return(nil)

	client := applicationoffers.NewClientFromCaller(mockFacadeCaller)
	err := accessCall(client, action, "carol", "read", someOffer, someOffer, someOffer)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *accessSuite) TestGrantErrorResult(c *gc.C) {
	s.errorResult(c, params.GrantOfferAccess)
}

func (s *accessSuite) TestRevokeErrorResult(c *gc.C) {
	s.errorResult(c, params.RevokeOfferAccess)
}

func (s *accessSuite) errorResult(c *gc.C, action params.OfferAction) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyOfferAccessRequest{
		Changes: []params.ModifyOfferAccess{
			{
				UserTag:  names.NewUserTag("aaa").String(),
				Action:   action,
				Access:   params.OfferConsumeAccess,
				OfferURL: someOffer,
			},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{Results: []params.ErrorResult{{Error: &params.Error{Message: "unfortunate mishap"}}}}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyOfferAccess", args, res).SetArg(2, ress).Return(nil)
	client := applicationoffers.NewClientFromCaller(mockFacadeCaller)

	err := accessCall(client, action, "aaa", "consume", someOffer)
	c.Assert(err, gc.ErrorMatches, "unfortunate mishap")
}

func (s *accessSuite) TestInvalidResultCount(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyOfferAccessRequest{
		Changes: []params.ModifyOfferAccess{
			{
				UserTag:  names.NewUserTag("bob").String(),
				Action:   params.GrantOfferAccess,
				Access:   params.OfferConsumeAccess,
				OfferURL: someOffer,
			},
			{
				UserTag:  names.NewUserTag("bob").String(),
				Action:   params.GrantOfferAccess,
				Access:   params.OfferConsumeAccess,
				OfferURL: someOffer,
			},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{Results: nil}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyOfferAccess", args, res).SetArg(2, ress).Return(nil)
	client := applicationoffers.NewClientFromCaller(mockFacadeCaller)

	err := client.GrantOffer("bob", "consume", someOffer, someOffer)
	c.Assert(err, gc.ErrorMatches, "expected 2 results, got 0")
}
