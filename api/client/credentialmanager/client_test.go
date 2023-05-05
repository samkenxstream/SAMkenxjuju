// Copyright 2018 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package credentialmanager_test

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/credentialmanager"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/rpc/params"
)

var _ = gc.Suite(&CredentialManagerSuite{})

type CredentialManagerSuite struct {
}

func (s *CredentialManagerSuite) TestInvalidateModelCredential(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	args := params.InvalidateCredentialArg{Reason: "auth fail"}
	result := new(params.ErrorResult)
	results := params.ErrorResult{}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("InvalidateModelCredential", args, result).SetArg(2, results).Return(nil)
	client := credentialmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.InvalidateModelCredential("auth fail")
	c.Assert(err, jc.ErrorIsNil)
}

func (s *CredentialManagerSuite) TestInvalidateModelCredentialBackendFailure(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	args := params.InvalidateCredentialArg{}
	result := new(params.ErrorResult)
	results := params.ErrorResult{Error: apiservererrors.ServerError(errors.New("boom"))}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("InvalidateModelCredential", args, result).SetArg(2, results).Return(nil)
	client := credentialmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.InvalidateModelCredential("")
	c.Assert(err, gc.ErrorMatches, "boom")
}

func (s *CredentialManagerSuite) TestInvalidateModelCredentialError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	args := params.InvalidateCredentialArg{}
	result := new(params.ErrorResult)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("InvalidateModelCredential", args, result).Return(errors.New("foo"))
	client := credentialmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.InvalidateModelCredential("")
	c.Assert(err, gc.ErrorMatches, "foo")
}
