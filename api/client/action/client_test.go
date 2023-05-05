// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package action_test

import (
	"github.com/golang/mock/gomock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/action"
	"github.com/juju/juju/rpc/params"
)

type actionSuite struct {
}

var _ = gc.Suite(&actionSuite{})

func (s *actionSuite) TestApplicationCharmActions(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	tests := []struct {
		description    string
		patchResults   []params.ApplicationCharmActionsResult
		patchErr       string
		expectedErr    string
		expectedResult map[string]action.ActionSpec
	}{{
		description: "result from wrong application",
		patchResults: []params.ApplicationCharmActionsResult{
			{
				ApplicationTag: names.NewApplicationTag("bar").String(),
			},
		},
		expectedErr: `action results received for wrong application "application-bar"`,
	}, {
		description: "some other error",
		patchResults: []params.ApplicationCharmActionsResult{
			{
				ApplicationTag: names.NewApplicationTag("foo").String(),
				Error: &params.Error{
					Message: "something bad",
				},
			},
		},
		expectedErr: `something bad`,
	}, {
		description: "more than one result",
		patchResults: []params.ApplicationCharmActionsResult{
			{},
			{},
		},
		expectedErr: "2 results, expected 1",
	}, {
		description:  "no results",
		patchResults: []params.ApplicationCharmActionsResult{},
		expectedErr:  "0 results, expected 1",
	}, {
		description: "error on facade call",
		patchErr:    "something went wrong",
		expectedErr: "something went wrong",
	}, {
		description: "normal result",
		patchResults: []params.ApplicationCharmActionsResult{
			{
				ApplicationTag: names.NewApplicationTag("foo").String(),
				Actions: map[string]params.ActionSpec{
					"action": {
						Description: "description",
						Params: map[string]interface{}{
							"foo": "bar",
						},
					},
				},
			},
		},
		expectedResult: map[string]action.ActionSpec{
			"action": {
				Description: "description",
				Params: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}}

	for i, t := range tests {
		c.Logf("test %d: %s", i, t.description)
		args := params.Entities{
			Entities: []params.Entity{{
				Tag: names.NewApplicationTag("foo").String(),
			}},
		}
		res := new(params.ApplicationsCharmActionsResults)
		ress := params.ApplicationsCharmActionsResults{
			Results: t.patchResults,
		}
		var facadeReturn error
		if t.patchErr != "" {
			facadeReturn = errors.New(t.patchErr)
		}
		mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
		mockFacadeCaller.EXPECT().FacadeCall("ApplicationsCharmsActions", args, res).SetArg(2, ress).Return(facadeReturn)
		client := action.NewClientFromCaller(mockFacadeCaller)

		result, err := client.ApplicationCharmActions("foo")
		if t.expectedErr != "" {
			c.Check(err, gc.ErrorMatches, t.expectedErr)
		} else {
			c.Check(err, jc.ErrorIsNil)
			c.Check(result, jc.DeepEquals, t.expectedResult)
		}
	}
}

func (s *actionSuite) TestWatchActionProgress(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{
			Tag: "action-666",
		}},
	}
	res := new(params.StringsWatchResults)
	ress := params.StringsWatchResults{
		Results: []params.StringsWatchResult{{
			Error: &params.Error{Message: "FAIL"},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("WatchActionsProgress", args, res).SetArg(2, ress).Return(nil)
	client := action.NewClientFromCaller(mockFacadeCaller)

	w, err := client.WatchActionProgress("666")
	c.Assert(w, gc.IsNil)
	c.Assert(err, gc.ErrorMatches, "FAIL")
}

func (s *actionSuite) TestWatchActionProgressArity(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{
			Tag: "action-666",
		}},
	}
	res := new(params.StringsWatchResults)
	ress := params.StringsWatchResults{
		Results: []params.StringsWatchResult{{
			Error: &params.Error{Message: "FAIL"},
		}, {
			Error: &params.Error{Message: "ANOTHER"},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("WatchActionsProgress", args, res).SetArg(2, ress).Return(nil)
	client := action.NewClientFromCaller(mockFacadeCaller)

	_, err := client.WatchActionProgress("666")
	c.Assert(err, gc.ErrorMatches, "expected 1 result, got 2")
}

func (s *actionSuite) TestListOperations(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	offset := 100
	limit := 200
	args := params.OperationQueryArgs{
		Applications: []string{"app"},
		Units:        []string{"unit/0"},
		Machines:     []string{"0"},
		ActionNames:  []string{"backup"},
		Status:       []string{"running"},
		Offset:       &offset,
		Limit:        &limit,
	}
	res := new(params.OperationResults)
	ress := params.OperationResults{
		Results: []params.OperationResult{{
			OperationTag: "operation-1",
			Summary:      "hello",
			Fail:         "fail",
			Status:       "error",
			Actions: []params.ActionResult{{
				Action: &params.Action{Tag: "action-666", Name: "test", Receiver: "unit-mysql-0"},
			}},
		}},
		Truncated: true,
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("ListOperations", args, res).SetArg(2, ress).Return(nil)
	client := action.NewClientFromCaller(mockFacadeCaller)

	result, err := client.ListOperations(action.OperationQueryArgs{
		Applications: []string{"app"},
		Units:        []string{"unit/0"},
		Machines:     []string{"0"},
		ActionNames:  []string{"backup"},
		Status:       []string{"running"},
		Offset:       &offset,
		Limit:        &limit,
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, action.Operations{
		Operations: []action.Operation{{
			ID:      "1",
			Summary: "hello",
			Status:  "error",
			Fail:    "fail",
			Actions: []action.ActionResult{{
				Action: &action.Action{ID: "666", Name: "test", Receiver: "unit-mysql-0"},
			}},
		}},
		Truncated: true,
	})
}

func (s *actionSuite) TestOperation(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{Entities: []params.Entity{{Tag: "operation-666"}}}
	res := new(params.OperationResults)
	ress := params.OperationResults{
		Results: []params.OperationResult{{
			OperationTag: "operation-1",
			Summary:      "hello",
			Fail:         "fail",
			Actions: []params.ActionResult{{
				Action: &params.Action{Tag: "action-666", Name: "test", Receiver: "unit-mysql-0"},
			}},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Operations", args, res).SetArg(2, ress).Return(nil)
	client := action.NewClientFromCaller(mockFacadeCaller)

	result, err := client.Operation("666")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, action.Operation{
		ID:      "1",
		Summary: "hello",
		Fail:    "fail",
		Actions: []action.ActionResult{{
			Action: &action.Action{ID: "666", Name: "test", Receiver: "unit-mysql-0"},
		}},
	})
}

func (s *actionSuite) TestEnqueueOperation(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := []action.Action{{
		Receiver: "unit/0",
		Name:     "test",
		Parameters: map[string]interface{}{
			"foo": "bar",
		}},
	}
	fArgs := params.Actions{
		Actions: []params.Action{{
			Receiver: "unit/0",
			Name:     "test",
			Parameters: map[string]interface{}{
				"foo": "bar",
			},
		}},
	}
	res := new(params.EnqueuedActions)
	ress := params.EnqueuedActions{
		OperationTag: "operation-1",
		Actions: []params.ActionResult{{
			Error: &params.Error{Message: "FAIL"},
		}},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("EnqueueOperation", fArgs, res).SetArg(2, ress).Return(nil)
	client := action.NewClientFromCaller(mockFacadeCaller)

	result, err := client.EnqueueOperation(args)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, action.EnqueuedActions{
		Actions: []action.ActionResult{{
			Error: &params.Error{Message: "FAIL"},
		}},
		OperationID: "1",
	})
}
