package PMFS

import "testing"

func TestActivateRequirementByID(t *testing.T) {
	SetBaseDir(t.TempDir())
	prj := &ProjectType{
		ID:        1,
		ProductID: 1,
		D: ProjectData{
			Requirements: []Requirement{
				{ID: 1, Condition: ConditionType{Proposed: true}},
				{ID: 2, Condition: ConditionType{Proposed: true}},
			},
		},
	}
	prj.ActivateRequirementByID(2)
	if prj.D.Requirements[1].Condition.Proposed || !prj.D.Requirements[1].Condition.Active {
		t.Fatalf("requirement not activated: %+v", prj.D.Requirements[1].Condition)
	}
	if !prj.D.Requirements[0].Condition.Proposed || prj.D.Requirements[0].Condition.Active {
		t.Fatalf("unexpected change to requirement 1: %+v", prj.D.Requirements[0].Condition)
	}
}

func TestActivateRequirementsWhere(t *testing.T) {
	SetBaseDir(t.TempDir())
	prj := &ProjectType{
		ID:        1,
		ProductID: 1,
		D: ProjectData{
			Requirements: []Requirement{
				{ID: 1, Priority: 1, Condition: ConditionType{Proposed: true}},
				{ID: 2, Priority: 2, Condition: ConditionType{Proposed: true}},
				{ID: 3, Priority: 1, Condition: ConditionType{Proposed: true}},
			},
		},
	}
	prj.ActivateRequirementsWhere(func(r Requirement) bool {
		return r.Priority == 1
	})
	if prj.D.Requirements[0].Condition.Proposed || !prj.D.Requirements[0].Condition.Active {
		t.Fatalf("requirement 1 not activated: %+v", prj.D.Requirements[0].Condition)
	}
	if prj.D.Requirements[2].Condition.Proposed || !prj.D.Requirements[2].Condition.Active {
		t.Fatalf("requirement 3 not activated: %+v", prj.D.Requirements[2].Condition)
	}
	if !prj.D.Requirements[1].Condition.Proposed || prj.D.Requirements[1].Condition.Active {
		t.Fatalf("requirement 2 should remain proposed: %+v", prj.D.Requirements[1].Condition)
	}
}
