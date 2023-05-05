// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cloud_test

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	"github.com/kr/pretty"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	cloudapi "github.com/juju/juju/api/client/cloud"
	"github.com/juju/juju/apiserver/common"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/cloud"
	"github.com/juju/juju/rpc/params"
	coretesting "github.com/juju/juju/testing"
)

type cloudSuite struct {
}

var _ = gc.Suite(&cloudSuite{})

func (s *cloudSuite) SetUpTest(c *gc.C) {
}

func (s *cloudSuite) TestCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{Tag: "cloud-foo"}},
	}
	res := new(params.CloudResults)
	results := params.CloudResults{
		Results: []params.CloudResult{{
			Cloud: &params.Cloud{
				Type:      "dummy",
				AuthTypes: []string{"empty", "userpass"},
				Regions:   []params.CloudRegion{{Name: "nether", Endpoint: "endpoint"}},
			}},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Cloud", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.Cloud(names.NewCloudTag("foo"))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, cloud.Cloud{
		Name:      "foo",
		Type:      "dummy",
		AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
		Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
	})
}

func (s *cloudSuite) TestCloudInfo(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{
			{Tag: "cloud-foo"}, {Tag: "cloud-bar"},
		},
	}
	res := new(params.CloudInfoResults)
	results := params.CloudInfoResults{
		Results: []params.CloudInfoResult{{
			Result: &params.CloudInfo{
				CloudDetails: params.CloudDetails{
					Type:      "dummy",
					AuthTypes: []string{"empty", "userpass"},
					Regions:   []params.CloudRegion{{Name: "nether", Endpoint: "endpoint"}},
				},
				Users: []params.CloudUserInfo{{
					UserName:    "fred",
					DisplayName: "Fred",
					Access:      "admin",
				}, {
					UserName:    "bob",
					DisplayName: "Bob",
					Access:      "add-model",
				}},
			},
		}, {
			Result: &params.CloudInfo{
				CloudDetails: params.CloudDetails{
					Type:      "dummy",
					AuthTypes: []string{"empty", "userpass"},
					Regions:   []params.CloudRegion{{Name: "nether", Endpoint: "endpoint"}},
				},
				Users: []params.CloudUserInfo{{
					UserName:    "mary",
					DisplayName: "Mary",
					Access:      "admin",
				}},
			},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CloudInfo", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.CloudInfo([]names.CloudTag{
		names.NewCloudTag("foo"),
		names.NewCloudTag("bar"),
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, []cloudapi.CloudInfo{{
		Cloud: cloud.Cloud{
			Name:      "foo",
			Type:      "dummy",
			AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
			Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
		},
		Users: map[string]cloudapi.CloudUserInfo{
			"bob": {
				DisplayName: "Bob",
				Access:      "add-model",
			},
			"fred": {
				DisplayName: "Fred",
				Access:      "admin",
			},
		},
	}, {
		Cloud: cloud.Cloud{
			Name:      "bar",
			Type:      "dummy",
			AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
			Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
		},
		Users: map[string]cloudapi.CloudUserInfo{
			"mary": {
				DisplayName: "Mary",
				Access:      "admin",
			},
		},
	}})
}

func (s *cloudSuite) TestClouds(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	res := new(params.CloudsResult)
	results := params.CloudsResult{
		Clouds: map[string]params.Cloud{
			"cloud-foo": {
				Type: "bar",
			},
			"cloud-baz": {
				Type:      "qux",
				AuthTypes: []string{"empty", "userpass"},
				Regions:   []params.CloudRegion{{Name: "nether", Endpoint: "endpoint"}},
			},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Clouds", nil, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	clouds, err := client.Clouds()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(clouds, jc.DeepEquals, map[names.CloudTag]cloud.Cloud{
		names.NewCloudTag("foo"): {
			Name: "foo",
			Type: "bar",
		},
		names.NewCloudTag("baz"): {
			Name:      "baz",
			Type:      "qux",
			AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
			Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
		},
	})
}

func (s *cloudSuite) TestUserCredentials(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UserClouds{UserClouds: []params.UserCloud{{
		UserTag:  "user-bob",
		CloudTag: "cloud-foo",
	}}}
	res := new(params.StringsResults)
	results := params.StringsResults{
		Results: []params.StringsResult{{
			Result: []string{
				"cloudcred-foo_bob_one",
				"cloudcred-foo_bob_two",
			},
		}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UserCredentials", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UserCredentials(names.NewUserTag("bob"), names.NewCloudTag("foo"))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.SameContents, []names.CloudCredentialTag{
		names.NewCloudCredentialTag("foo/bob/one"),
		names.NewCloudCredentialTag("foo/bob/two"),
	})
}

func (s *cloudSuite) TestUpdateCredential(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{{}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UpdateCredentialsCheckModels(testCredentialTag, testCredential)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, gc.IsNil)
}

func (s *cloudSuite) TestUpdateCredentialError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{
				CredentialTag: "cloudcred-foo_bob_bar",
				Error:         apiservererrors.ServerError(errors.New("validation failure")),
			},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	errs, err := client.UpdateCredentialsCheckModels(testCredentialTag, testCredential)
	c.Assert(err, gc.ErrorMatches, "validation failure")
	c.Assert(errs, gc.IsNil)
}

func (s *cloudSuite) TestUpdateCredentialManyResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{},
			{},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UpdateCredentialsCheckModels(testCredentialTag, testCredential)
	c.Assert(err, gc.ErrorMatches, `expected 1 result got 2 when updating credentials`)
	c.Assert(result, gc.IsNil)
}

func (s *cloudSuite) TestUpdateCredentialModelErrors(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{
				CredentialTag: testCredentialTag.String(),
				Models: []params.UpdateCredentialModelResult{
					{
						ModelUUID: coretesting.ModelTag.Id(),
						ModelName: "test-model",
						Errors: []params.ErrorResult{
							{apiservererrors.ServerError(errors.New("validation failure one"))},
							{apiservererrors.ServerError(errors.New("validation failure two"))},
						},
					},
				},
			},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	errs, err := client.UpdateCredentialsCheckModels(testCredentialTag, testCredential)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(errs, gc.DeepEquals, []params.UpdateCredentialModelResult{
		{
			ModelUUID: "deadbeef-0bad-400d-8000-4b1d0d06f00d",
			ModelName: "test-model",
			Errors: []params.ErrorResult{
				{Error: &params.Error{Message: "validation failure one", Code: ""}},
				{Error: &params.Error{Message: "validation failure two", Code: ""}},
			},
		},
	})
}

var (
	testCredentialTag = names.NewCloudCredentialTag("foo/bob/bar")
	testCredential    = cloud.NewCredential(cloud.UserPassAuthType, map[string]string{
		"username": "admin",
		"password": "adm1n",
	})
)

func (s *cloudSuite) TestRevokeCredential(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.RevokeCredentialArgs{
		Credentials: []params.RevokeCredentialArg{
			{Tag: "cloudcred-foo_bob_bar", Force: true},
		},
	}
	res := new(params.ErrorResults)
	results := params.ErrorResults{
		Results: []params.ErrorResult{{}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("RevokeCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	tag := names.NewCloudCredentialTag("foo/bob/bar")
	err := client.RevokeCredential(tag, true)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *cloudSuite) TestCredentials(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{Entities: []params.Entity{{
		Tag: "cloudcred-foo_bob_bar",
	}}}
	res := new(params.CloudCredentialResults)
	results := params.CloudCredentialResults{
		Results: []params.CloudCredentialResult{
			{
				Result: &params.CloudCredential{
					AuthType:   "userpass",
					Attributes: map[string]string{"username": "fred"},
					Redacted:   []string{"password"},
				},
			}, {
				Result: &params.CloudCredential{
					AuthType:   "userpass",
					Attributes: map[string]string{"username": "mary"},
					Redacted:   []string{"password"},
				},
			},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Credential", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	tag := names.NewCloudCredentialTag("foo/bob/bar")
	result, err := client.Credentials(tag)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, []params.CloudCredentialResult{
		{
			Result: &params.CloudCredential{
				AuthType:   "userpass",
				Attributes: map[string]string{"username": "fred"},
				Redacted:   []string{"password"},
			},
		}, {
			Result: &params.CloudCredential{
				AuthType:   "userpass",
				Attributes: map[string]string{"username": "mary"},
				Redacted:   []string{"password"},
			},
		},
	})
}

var testCloud = cloud.Cloud{
	Name:      "foo",
	Type:      "dummy",
	AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
	Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
}

func (s *cloudSuite) TestAddCloudForce(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	force := true
	args := params.AddCloudArgs{
		Name:  "foo",
		Cloud: common.CloudToParams(testCloud),
		Force: &force,
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("AddCloud", args, nil).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.AddCloud(testCloud, force)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *cloudSuite) TestCredentialContentsArgumentCheck(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	// Check supplying cloud name without credential name is invalid.
	result, err := client.CredentialContents("cloud", "", true)
	c.Assert(result, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "credential name must be supplied")

	// Check supplying credential name without cloud name is invalid.
	result, err = client.CredentialContents("", "credential", true)
	c.Assert(result, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "cloud name must be supplied")
}

func (s *cloudSuite) TestCredentialContentsAll(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	expectedResults := []params.CredentialContentResult{
		{
			Result: &params.ControllerCredentialInfo{
				Content: params.CredentialContent{
					Cloud:    "cloud-name",
					Name:     "credential-name",
					AuthType: "userpass",
					Attributes: map[string]string{
						"username": "fred",
						"password": "sekret"},
				},
				Models: []params.ModelAccess{
					{Model: "abcmodel", Access: "admin"},
					{Model: "xyzmodel", Access: "read"},
					{Model: "no-access-model"}, // no access
				},
			},
		}, {
			Error: apiservererrors.ServerError(errors.New("boom")),
		},
	}

	args := params.CloudCredentialArgs{
		IncludeSecrets: true,
	}
	res := new(params.CredentialContentResults)
	results := params.CredentialContentResults{
		Results: expectedResults,
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CredentialContents", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	ress, err := client.CredentialContents("", "", true)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ress, jc.DeepEquals, expectedResults)
}

func (s *cloudSuite) TestCredentialContentsOne(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.CloudCredentialArgs{
		IncludeSecrets: true,
		Credentials: []params.CloudCredentialArg{
			{CloudName: "cloud-name", CredentialName: "credential-name"},
		},
	}
	res := new(params.CredentialContentResults)
	ress := params.CredentialContentResults{
		Results: []params.CredentialContentResult{
			{},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CredentialContents", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	results, err := client.CredentialContents("cloud-name", "credential-name", true)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.HasLen, 1)
}

func (s *cloudSuite) TestCredentialContentsGotMoreThanBargainedFor(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.CloudCredentialArgs{
		IncludeSecrets: true,
		Credentials: []params.CloudCredentialArg{
			{CloudName: "cloud-name", CredentialName: "credential-name"},
		},
	}
	res := new(params.CredentialContentResults)
	ress := params.CredentialContentResults{
		Results: []params.CredentialContentResult{
			{},
			{},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CredentialContents", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	results, err := client.CredentialContents("cloud-name", "credential-name", true)
	c.Assert(results, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "expected 1 result for credential \"cloud-name\" on cloud \"credential-name\", got 2")
}

func (s *cloudSuite) TestCredentialContentsServerError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.CloudCredentialArgs{
		IncludeSecrets: true,
	}
	res := new(params.CredentialContentResults)
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CredentialContents", args, res).Return(errors.New("boom"))
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	results, err := client.CredentialContents("", "", true)
	c.Assert(results, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "boom")
}

func (s *cloudSuite) TestRemoveCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{Tag: "cloud-foo"}},
	}
	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{
			Error: &params.Error{Message: "FAIL"},
		}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("RemoveClouds", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.RemoveCloud("foo")
	c.Assert(err, gc.ErrorMatches, "FAIL")
}

func (s *cloudSuite) TestRemoveCloudErrorMapping(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{Tag: "cloud-foo"}},
	}
	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{
			Error: &params.Error{
				Code:    params.CodeNotFound,
				Message: `cloud "cloud-foo" not found`,
			}},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("RemoveClouds", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.RemoveCloud("foo")
	c.Assert(err, jc.Satisfies, errors.IsNotFound, gc.Commentf("expected client to be map server error into a NotFound error"))
}

func (s *cloudSuite) TestGrantCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyCloudAccessRequest{
		Changes: []params.ModifyCloudAccess{
			{UserTag: "user-fred", CloudTag: "cloud-fluffy", Action: "grant", Access: "admin"},
		},
	}
	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{
			Error: &params.Error{Message: "FAIL"}},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyCloudAccess", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.GrantCloud("fred", "admin", "fluffy")
	c.Assert(err, gc.ErrorMatches, "FAIL")
}

func (s *cloudSuite) TestRevokeCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModifyCloudAccessRequest{
		Changes: []params.ModifyCloudAccess{
			{UserTag: "user-fred", CloudTag: "cloud-fluffy", Action: "revoke", Access: "admin"},
		},
	}
	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{
			Error: &params.Error{Message: "FAIL"}},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModifyCloudAccess", args, res).SetArg(2, ress).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.RevokeCloud("fred", "admin", "fluffy")
	c.Assert(err, gc.ErrorMatches, "FAIL")
}

func createCredentials(n int) map[string]cloud.Credential {
	result := map[string]cloud.Credential{}
	for i := 0; i < n; i++ {
		result[names.NewCloudCredentialTag(fmt.Sprintf("foo/bob/bar%d", i)).String()] = testCredential
	}
	return result
}

func (s *cloudSuite) TestUpdateCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	updatedCloud := cloud.Cloud{
		Name:      "foo",
		Type:      "dummy",
		AuthTypes: []cloud.AuthType{cloud.EmptyAuthType, cloud.UserPassAuthType},
		Regions:   []cloud.Region{{Name: "nether", Endpoint: "endpoint"}},
	}

	args := params.UpdateCloudArgs{Clouds: []params.AddCloudArgs{{
		Name:  "foo",
		Cloud: common.CloudToParams(updatedCloud),
	}}}
	res := new(params.ErrorResults)
	results := params.ErrorResults{
		Results: []params.ErrorResult{{}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCloud", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	err := client.UpdateCloud(updatedCloud)

	c.Assert(err, jc.ErrorIsNil)
}

func (s *cloudSuite) TestUpdateCloudsCredentials(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Force: true,
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar0",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{{}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UpdateCloudsCredentials(createCredentials(1), true)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, gc.DeepEquals, []params.UpdateCredentialResult{{}})
}

func (s *cloudSuite) TestUpdateCloudsCredentialsError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Force: false,
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar0",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{CredentialTag: "cloudcred-foo_bob_bar0",
				Error: apiservererrors.ServerError(errors.New("validation failure")),
			},
		},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	errs, err := client.UpdateCloudsCredentials(createCredentials(1), false)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(errs, gc.DeepEquals, []params.UpdateCredentialResult{
		{CredentialTag: "cloudcred-foo_bob_bar0", Error: apiservererrors.ServerError(errors.New("validation failure"))},
	})
}

func (s *cloudSuite) TestUpdateCloudsCredentialsManyResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Force: false,
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar0",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{},
			{},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UpdateCloudsCredentials(createCredentials(1), false)
	c.Assert(err, gc.ErrorMatches, `expected 1 result got 2 when updating credentials`)
	c.Assert(result, gc.IsNil)
}

func (s *cloudSuite) TestUpdateCloudsCredentialsManyMatchingResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Force: false,
	}
	count := 2
	for tag, credential := range createCredentials(count) {
		args.Credentials = append(args.Credentials, params.TaggedCredential{
			Tag: tag,
			Credential: params.CloudCredential{
				AuthType:   string(credential.AuthType()),
				Attributes: credential.Attributes(),
			},
		})
	}

	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{},
			{},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", cloudCredentialMatcher{args}, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.UpdateCloudsCredentials(createCredentials(count), false)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, gc.HasLen, count)
}

func (s *cloudSuite) TestUpdateCloudsCredentialsModelErrors(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Force: false,
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar0",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{
			{
				CredentialTag: testCredentialTag.String(),
				Models: []params.UpdateCredentialModelResult{
					{
						ModelUUID: coretesting.ModelTag.Id(),
						ModelName: "test-model",
						Errors: []params.ErrorResult{
							{apiservererrors.ServerError(errors.New("validation failure one"))},
							{apiservererrors.ServerError(errors.New("validation failure two"))},
						},
					},
				},
			},
		}}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	errs, err := client.UpdateCloudsCredentials(createCredentials(1), false)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(errs, gc.DeepEquals, []params.UpdateCredentialResult{
		{CredentialTag: "cloudcred-foo_bob_bar",
			Models: []params.UpdateCredentialModelResult{
				{ModelUUID: "deadbeef-0bad-400d-8000-4b1d0d06f00d",
					ModelName: "test-model",
					Errors: []params.ErrorResult{
						{apiservererrors.ServerError(errors.New("validation failure one"))},
						{apiservererrors.ServerError(errors.New("validation failure two"))},
					},
				},
			},
		},
	})
}

func (s *cloudSuite) TestAddCloudsCredentials(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UpdateCredentialArgs{
		Credentials: []params.TaggedCredential{{
			Tag: "cloudcred-foo_bob_bar0",
			Credential: params.CloudCredential{
				AuthType: "userpass",
				Attributes: map[string]string{
					"username": "admin",
					"password": "adm1n",
				},
			},
		}}}
	res := new(params.UpdateCredentialResults)
	results := params.UpdateCredentialResults{
		Results: []params.UpdateCredentialResult{{}},
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UpdateCredentialsCheckModels", args, res).SetArg(2, results).Return(nil)
	client := cloudapi.NewClientFromCaller(mockFacadeCaller)

	result, err := client.AddCloudsCredentials(createCredentials(1))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, gc.DeepEquals, []params.UpdateCredentialResult{{}})
}

type cloudCredentialMatcher struct {
	arg params.UpdateCredentialArgs
}

func (c cloudCredentialMatcher) Matches(x interface{}) bool {
	cred, ok := x.(params.UpdateCredentialArgs)
	if !ok {
		return false
	}
	if len(cred.Credentials) != len(c.arg.Credentials) {
		return false
	}
	// sort both input and expected slices the same way to avoid ordering discrepancies when ranging.
	sort.Slice(cred.Credentials, func(i, j int) bool { return cred.Credentials[i].Tag < cred.Credentials[j].Tag })
	sort.Slice(c.arg.Credentials, func(i, j int) bool { return c.arg.Credentials[i].Tag < c.arg.Credentials[j].Tag })
	for idx, taggedCred := range cred.Credentials {
		if taggedCred.Tag != c.arg.Credentials[idx].Tag {
			return false
		}
		if !reflect.DeepEqual(taggedCred.Credential, c.arg.Credentials[idx].Credential) {
			return false
		}
	}

	if cred.Force != c.arg.Force {
		return false
	}
	return true
}

func (c cloudCredentialMatcher) String() string {
	return pretty.Sprint(c.arg)
}
