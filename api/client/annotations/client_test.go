// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package annotations_test

import (
	"github.com/golang/mock/gomock"
	jc "github.com/juju/testing/checkers"
	"github.com/kr/pretty"
	gc "gopkg.in/check.v1"

	basemocks "github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/client/annotations"
	"github.com/juju/juju/rpc/params"
)

type annotationsMockSuite struct{}

var _ = gc.Suite(&annotationsMockSuite{})

func (s *annotationsMockSuite) TestSetEntitiesAnnotation(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	annts := map[string]string{"annotation1": "test"}
	annts2 := map[string]string{"annotation2": "test"}
	setParams := map[string]map[string]string{
		"charmA":       annts,
		"applicationB": annts2,
	}

	args := params.AnnotationsSet{
		Annotations: []params.EntityAnnotations{
			{
				EntityTag:   "charmA",
				Annotations: annts,
			},
			{
				EntityTag:   "applicationB",
				Annotations: annts2,
			},
		},
	}

	result := new(params.ErrorResults)
	results := params.ErrorResults{
		Results: nil,
	}
	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Set", annotationsSetMatcher{c, args}, result).SetArg(2, results).DoAndReturn(
		func(arg0 string, args params.AnnotationsSet, results *params.ErrorResults) []error {
			for _, aParam := range args.Annotations {
				// Since sometimes arrays returned on some
				// architectures vary the order within params.AnnotationsSet,
				// simply assert that each entity has its own annotations.
				// Bug 1409141
				c.Assert(aParam.Annotations, gc.DeepEquals, setParams[aParam.EntityTag])
			}
			return nil
		})

	annotationsClient := annotations.NewClientFromCaller(mockFacadeCaller)
	callErrs, err := annotationsClient.Set(setParams)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(callErrs, gc.HasLen, 0)
}

func (s *annotationsMockSuite) TestGetEntitiesAnnotations(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	args := params.Entities{
		Entities: []params.Entity{{"charm"}},
	}
	facadeAnnts := map[string]string{
		"annotations": "test",
	}
	entitiesAnnts := params.AnnotationsGetResult{
		EntityTag:   "charm",
		Annotations: facadeAnnts,
	}
	result := new(params.AnnotationsGetResults)
	results := params.AnnotationsGetResults{
		Results: []params.AnnotationsGetResult{entitiesAnnts},
	}

	mockFacadeCaller := basemocks.NewMockFacadeCaller(ctrl)
	mockFacadeCaller.EXPECT().FacadeCall("Get", args, result).SetArg(2, results).Return(nil)

	annotationsClient := annotations.NewClientFromCaller(mockFacadeCaller)
	found, err := annotationsClient.Get([]string{"charm"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(found, gc.HasLen, 1)
}

type annotationsSetMatcher struct {
	m            *gc.C
	expectedArgs params.AnnotationsSet
}

func (c annotationsSetMatcher) Matches(x interface{}) bool {
	obtainedArgs, ok := x.(params.AnnotationsSet)
	if !ok {
		return false
	}
	c.m.Assert(obtainedArgs.Annotations, gc.HasLen, len(c.expectedArgs.Annotations))

	for _, obt := range obtainedArgs.Annotations {
		var found bool
		for _, exp := range c.expectedArgs.Annotations {
			if obt.EntityTag == exp.EntityTag {
				c.m.Assert(obt, jc.DeepEquals, exp)
				found = true
				break
			}
		}
		c.m.Assert(found, jc.IsTrue, gc.Commentf("unexpected annotation entity tag %s"))
	}
	return true
}

func (c annotationsSetMatcher) String() string {
	return pretty.Sprintf("Match the contents of %v", c.expectedArgs)
}
