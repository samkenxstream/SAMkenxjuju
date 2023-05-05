// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package modelmanager_test

import (
	"regexp"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api/base"
	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/modelmanager"
	"github.com/juju/juju/api/common"
	apiservererrors "github.com/juju/juju/apiserver/errors"
	"github.com/juju/juju/core/life"
	"github.com/juju/juju/core/model"
	"github.com/juju/juju/core/status"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/rpc/params"
	coretesting "github.com/juju/juju/testing"
)

type modelmanagerSuite struct {
}

var _ = gc.Suite(&modelmanagerSuite{})

func (s *modelmanagerSuite) TestCreateModelBadUser(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	_, err := client.CreateModel("mymodel", "not a user", "", "", names.CloudCredentialTag{}, nil)
	c.Assert(err, gc.ErrorMatches, `invalid owner name "not a user"`)
}

func (s *modelmanagerSuite) TestCreateModelBadCloud(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	_, err := client.CreateModel("mymodel", "bob", "123!", "", names.CloudCredentialTag{}, nil)
	c.Assert(err, gc.ErrorMatches, `invalid cloud name "123!"`)
}

func (s *modelmanagerSuite) TestCreateModel(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModelCreateArgs{
		Name:        "new-model",
		OwnerTag:    "user-bob",
		Config:      map[string]interface{}{"abc": 123},
		CloudTag:    "cloud-nimbus",
		CloudRegion: "catbus",
	}

	result := new(params.ModelInfo)
	ress := params.ModelInfo{}
	ress.Name = "dowhatimean"
	ress.Type = "iaas"
	ress.UUID = "youyoueyedee"
	ress.ControllerUUID = "youyoueyedeetoo"
	ress.ProviderType = "C-123"
	ress.DefaultSeries = "M*A*S*H"
	ress.CloudTag = "cloud-nimbus"
	ress.CloudRegion = "catbus"
	ress.OwnerTag = "user-fnord"
	ress.Life = "alive"

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("CreateModel", args, result).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	newModel, err := client.CreateModel(
		"new-model",
		"bob",
		"nimbus",
		"catbus",
		names.CloudCredentialTag{},
		map[string]interface{}{"abc": 123},
	)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(newModel, jc.DeepEquals, base.ModelInfo{
		Name:           "dowhatimean",
		Type:           model.IAAS,
		UUID:           "youyoueyedee",
		ControllerUUID: "youyoueyedeetoo",
		ProviderType:   "C-123",
		DefaultSeries:  "M*A*S*H",
		Cloud:          "nimbus",
		CloudRegion:    "catbus",
		Owner:          "fnord",
		Life:           "alive",
		Status: base.Status{
			Data: make(map[string]interface{}),
		},
		Users:    []base.UserInfo{},
		Machines: []base.Machine{},
	})
}

func (s *modelmanagerSuite) TestListModelsBadUser(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	_, err := client.ListModels("not a user")
	c.Assert(err, gc.ErrorMatches, `invalid user name "not a user"`)
}

func (s *modelmanagerSuite) TestListModels(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	lastConnection := time.Now()
	args := params.Entity{"user-user@remote"}

	result := new(params.UserModelList)
	ress := params.UserModelList{
		UserModels: []params.UserModel{{
			Model: params.Model{
				Name:     "yo",
				UUID:     "wei",
				Type:     "caas",
				OwnerTag: "user-user@remote",
			},
			LastConnection: &lastConnection,
		}, {
			Model: params.Model{
				Name:     "sup",
				UUID:     "hazzagarn",
				Type:     "iaas",
				OwnerTag: "user-phyllis@thrace",
			},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListModels", args, result).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	models, err := client.ListModels("user@remote")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(models, jc.DeepEquals, []base.UserModel{{
		Name:           "yo",
		UUID:           "wei",
		Type:           model.CAAS,
		Owner:          "user@remote",
		LastConnection: &lastConnection,
	}, {
		Name:  "sup",
		UUID:  "hazzagarn",
		Type:  model.IAAS,
		Owner: "phyllis@thrace",
	}})
}

func (s *modelmanagerSuite) testDestroyModel(c *gc.C, destroyStorage, force *bool, maxWait *time.Duration, timeout time.Duration) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.DestroyModelsParams{
		Models: []params.DestroyModelParams{{
			ModelTag:       coretesting.ModelTag.String(),
			DestroyStorage: destroyStorage,
			Force:          force,
			MaxWait:        maxWait,
			Timeout:        &timeout,
		}},
	}

	result := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("DestroyModels", args, result).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.DestroyModel(coretesting.ModelTag, destroyStorage, force, maxWait, timeout)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *modelmanagerSuite) TestDestroyModel(c *gc.C) {
	true_ := true
	false_ := false
	defaultMin := 1 * time.Minute
	s.testDestroyModel(c, nil, nil, nil, time.Minute)
	s.testDestroyModel(c, nil, &true_, nil, time.Minute)
	s.testDestroyModel(c, nil, &true_, &defaultMin, time.Minute)
	s.testDestroyModel(c, nil, &false_, nil, time.Minute)
	s.testDestroyModel(c, &true_, nil, nil, time.Minute)
	s.testDestroyModel(c, &true_, &false_, nil, time.Minute)
	s.testDestroyModel(c, &true_, &true_, &defaultMin, time.Minute)
	s.testDestroyModel(c, &false_, nil, nil, time.Minute)
	s.testDestroyModel(c, &false_, &false_, nil, time.Minute)
	s.testDestroyModel(c, &false_, &true_, &defaultMin, time.Minute)
}

func (s *modelmanagerSuite) TestModelDefaults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{Tag: names.NewCloudTag("aws").String()}},
	}

	res := new(params.ModelDefaultsResults)
	ress := params.ModelDefaultsResults{
		Results: []params.ModelDefaultsResult{{Config: map[string]params.ModelDefaults{
			"foo": {"bar", "model", []params.RegionDefaults{{
				"dummy-region",
				"dummy-value"}}},
		}}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModelDefaultsForClouds", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	result, err := client.ModelDefaults("aws")
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(result, jc.DeepEquals, config.ModelDefaultAttributes{
		"foo": {"bar", "model", []config.RegionDefaultValue{{
			"dummy-region",
			"dummy-value"}}},
	})
}

func (s *modelmanagerSuite) TestSetModelDefaults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.SetModelDefaults{
		Config: []params.ModelDefaultValues{{
			CloudTag:    "cloud-mycloud",
			CloudRegion: "region",
			Config: map[string]interface{}{
				"some-name":  "value",
				"other-name": true,
			},
		}}}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("SetModelDefaults", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.SetModelDefaults("mycloud", "region", map[string]interface{}{
		"some-name":  "value",
		"other-name": true,
	})
	c.Assert(err, jc.ErrorIsNil)
}

func (s *modelmanagerSuite) TestUnsetModelDefaults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.UnsetModelDefaults{
		Keys: []params.ModelUnsetKeys{{
			CloudTag:    "cloud-mycloud",
			CloudRegion: "region",
			Keys:        []string{"foo", "bar"},
		}}}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("UnsetModelDefaults", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.UnsetModelDefaults("mycloud", "region", "foo", "bar")
	c.Assert(err, jc.ErrorIsNil)
}

func (s *modelmanagerSuite) TestModelStatus(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{
			{Tag: coretesting.ModelTag.String()},
			{Tag: coretesting.ModelTag.String()},
		},
	}

	res := new(params.ModelStatusResults)
	ress := params.ModelStatusResults{
		Results: []params.ModelStatus{
			{
				ModelTag:           coretesting.ModelTag.String(),
				OwnerTag:           "user-glenda",
				ApplicationCount:   3,
				HostedMachineCount: 2,
				Life:               "alive",
				Machines: []params.ModelMachineInfo{{
					Id:         "0",
					InstanceId: "inst-ance",
					Status:     "pending",
				}},
			},
			{
				Error: apiservererrors.ServerError(errors.New("model error")),
			},
		},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModelStatus", args, res).SetArg(2, ress).Return(nil)
	client := common.NewModelStatusAPI(mockFacadeCaller)

	results, err := client.ModelStatus(coretesting.ModelTag, coretesting.ModelTag)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results[0], jc.DeepEquals, base.ModelStatus{
		UUID:               coretesting.ModelTag.Id(),
		TotalMachineCount:  1,
		HostedMachineCount: 2,
		ApplicationCount:   3,
		Owner:              "glenda",
		Life:               life.Alive,
		Machines:           []base.Machine{{Id: "0", InstanceId: "inst-ance", Status: "pending"}},
	})
	c.Assert(results[1].Error, gc.ErrorMatches, "model error")
}

func (s *modelmanagerSuite) TestModelStatusEmpty(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{},
	}

	res := new(params.ModelStatusResults)
	ress := params.ModelStatusResults{}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModelStatus", args, res).SetArg(2, ress).Return(nil)
	client := common.NewModelStatusAPI(mockFacadeCaller)

	results, err := client.ModelStatus()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, []base.ModelStatus{})
}

func (s *modelmanagerSuite) TestModelStatusError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{
			{Tag: coretesting.ModelTag.String()},
			{Tag: coretesting.ModelTag.String()},
		},
	}

	res := new(params.ModelStatusResults)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ModelStatus", args, res).Return(errors.New("model error"))
	client := common.NewModelStatusAPI(mockFacadeCaller)
	out, err := client.ModelStatus(coretesting.ModelTag, coretesting.ModelTag)
	c.Assert(err, gc.ErrorMatches, "model error")
	c.Assert(out, gc.IsNil)
}

func createModelSummary() *params.ModelSummary {
	return &params.ModelSummary{
		Name:               "name",
		UUID:               "uuid",
		Type:               "iaas",
		ControllerUUID:     "controllerUUID",
		ProviderType:       "aws",
		DefaultSeries:      "xenial",
		CloudTag:           "cloud-aws",
		CloudRegion:        "us-east-1",
		CloudCredentialTag: "cloudcred-foo_bob_one",
		OwnerTag:           "user-admin",
		Life:               life.Alive,
		Status:             params.EntityStatus{Status: status.Status("active")},
		UserAccess:         params.ModelAdminAccess,
		Counts:             []params.ModelEntityCount{},
	}
}

func (s *modelmanagerSuite) TestListModelSummaries(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	userTag := names.NewUserTag("commander")
	testModelInfo := createModelSummary()

	args := params.ModelSummariesRequest{
		UserTag: userTag.String(),
		All:     true,
	}

	res := new(params.ModelSummaryResults)
	ress := params.ModelSummaryResults{
		Results: []params.ModelSummaryResult{
			{Result: testModelInfo},
			{Error: apiservererrors.ServerError(errors.New("model error"))},
		},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListModelSummaries", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	results, err := client.ListModelSummaries(userTag.Id(), true)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(results, gc.HasLen, 2)
	c.Assert(results[0], jc.DeepEquals, base.UserModelSummary{Name: testModelInfo.Name,
		UUID:            testModelInfo.UUID,
		Type:            model.IAAS,
		ControllerUUID:  testModelInfo.ControllerUUID,
		ProviderType:    testModelInfo.ProviderType,
		DefaultSeries:   testModelInfo.DefaultSeries,
		Cloud:           "aws",
		CloudRegion:     "us-east-1",
		CloudCredential: "foo/bob/one",
		Owner:           "admin",
		Life:            "alive",
		Status: base.Status{
			Status: status.Active,
			Data:   map[string]interface{}{},
		},
		ModelUserAccess: "admin",
		Counts:          []base.EntityCount{},
	})
	c.Assert(errors.Cause(results[1].Error), gc.ErrorMatches, "model error")
}

func (s *modelmanagerSuite) TestListModelSummariesParsingErrors(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	badOwnerInfo := createModelSummary()
	badOwnerInfo.OwnerTag = "owner-user"

	badCloudInfo := createModelSummary()
	badCloudInfo.CloudTag = "not-cloud"

	badCredentialsInfo := createModelSummary()
	badCredentialsInfo.CloudCredentialTag = "not-credential"

	args := params.ModelSummariesRequest{
		UserTag: "user-commander",
		All:     true,
	}

	res := new(params.ModelSummaryResults)
	ress := params.ModelSummaryResults{
		Results: []params.ModelSummaryResult{
			{Result: badOwnerInfo},
			{Result: badCloudInfo},
			{Result: badCredentialsInfo},
		},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListModelSummaries", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	results, err := client.ListModelSummaries("commander", true)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.HasLen, 3)
	c.Assert(results[0].Error, gc.ErrorMatches, `while parsing model owner tag: "owner-user" is not a valid tag`)
	c.Assert(results[1].Error, gc.ErrorMatches, `while parsing model cloud tag: "not-cloud" is not a valid tag`)
	c.Assert(results[2].Error, gc.ErrorMatches, `while parsing model cloud credential tag: "not-credential" is not a valid tag`)
}

func (s *modelmanagerSuite) TestListModelSummariesInvalidUserIn(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	out, err := client.ListModelSummaries("++)captain", false)
	c.Assert(err, gc.ErrorMatches, regexp.QuoteMeta(`invalid user name "++)captain"`))
	c.Assert(out, gc.IsNil)
}

func (s *modelmanagerSuite) TestListModelSummariesServerError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.ModelSummariesRequest{
		UserTag: "user-captain",
		All:     false,
	}

	res := new(params.ModelSummaryResults)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListModelSummaries", args, res).Return(errors.New("captain, error"))
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	out, err := client.ListModelSummaries("captain", false)
	c.Assert(err, gc.ErrorMatches, "captain, error")
	c.Assert(out, gc.IsNil)
}

func (s *modelmanagerSuite) TestChangeModelCredential(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	credentialTag := names.NewCloudCredentialTag("foo/bob/bar")
	args := params.ChangeModelCredentialsParams{
		Models: []params.ChangeModelCredentialParams{
			{ModelTag: coretesting.ModelTag.String(), CloudCredentialTag: credentialTag.String()},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ChangeModelCredential", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.ChangeModelCredential(coretesting.ModelTag, credentialTag)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *modelmanagerSuite) TestChangeModelCredentialManyResults(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	credentialTag := names.NewCloudCredentialTag("foo/bob/bar")

	args := params.ChangeModelCredentialsParams{
		Models: []params.ChangeModelCredentialParams{
			{ModelTag: coretesting.ModelTag.String(), CloudCredentialTag: credentialTag.String()},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{}, {}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ChangeModelCredential", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.ChangeModelCredential(coretesting.ModelTag, credentialTag)
	c.Assert(err, gc.ErrorMatches, `expected 1 result, got 2`)
}

func (s *modelmanagerSuite) TestChangeModelCredentialCallFailed(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	credentialTag := names.NewCloudCredentialTag("foo/bob/bar")
	args := params.ChangeModelCredentialsParams{
		Models: []params.ChangeModelCredentialParams{
			{ModelTag: coretesting.ModelTag.String(), CloudCredentialTag: credentialTag.String()},
		},
	}

	res := new(params.ErrorResults)

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ChangeModelCredential", args, res).Return(errors.New("failed call"))
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)
	err := client.ChangeModelCredential(coretesting.ModelTag, credentialTag)
	c.Assert(err, gc.ErrorMatches, `failed call`)
}

func (s *modelmanagerSuite) TestChangeModelCredentialUpdateFailed(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	credentialTag := names.NewCloudCredentialTag("foo/bob/bar")
	args := params.ChangeModelCredentialsParams{
		Models: []params.ChangeModelCredentialParams{
			{ModelTag: coretesting.ModelTag.String(), CloudCredentialTag: credentialTag.String()},
		},
	}

	res := new(params.ErrorResults)
	ress := params.ErrorResults{
		Results: []params.ErrorResult{{Error: apiservererrors.ServerError(errors.New("update error"))}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ChangeModelCredential", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	err := client.ChangeModelCredential(coretesting.ModelTag, credentialTag)
	c.Assert(err, gc.ErrorMatches, `update error`)
}

type dumpModelSuite struct {
	coretesting.BaseSuite
}

var _ = gc.Suite(&dumpModelSuite{})

func (s *dumpModelSuite) TestDumpModelDB(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	expected := map[string]interface{}{
		"models": []map[string]interface{}{{
			"name": "admin",
			"uuid": "some-uuid",
		}},
		"machines": []map[string]interface{}{{
			"id":   "0",
			"life": 0,
		}},
	}
	args := params.Entities{[]params.Entity{{coretesting.ModelTag.String()}}}

	res := new(params.MapResults)
	ress := params.MapResults{Results: []params.MapResult{{
		Result: expected,
	}}}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("DumpModelsDB", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	out, err := client.DumpModelDB(coretesting.ModelTag)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(out, jc.DeepEquals, expected)
}

func (s *dumpModelSuite) TestDumpModelDBError(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{[]params.Entity{{coretesting.ModelTag.String()}}}

	res := new(params.MapResults)
	ress := params.MapResults{Results: []params.MapResult{{
		Error: &params.Error{Message: "fake error"},
	}}}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("DumpModelsDB", args, res).SetArg(2, ress).Return(nil)
	client := modelmanager.NewClientFromCaller(mockFacadeCaller)

	out, err := client.DumpModelDB(coretesting.ModelTag)
	c.Assert(err, gc.ErrorMatches, "fake error")
	c.Assert(out, gc.IsNil)
}
