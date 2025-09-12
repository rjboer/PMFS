package PMFS

import "testing"

func TestDeleteAndRestoreRequirementByID(t *testing.T) {
	SetBaseDir(t.TempDir())
	prj := &ProjectType{
		ID:        1,
		ProductID: 1,
		D:         ProjectData{Requirements: []Requirement{{ID: 1, Condition: ConditionType{Active: true}}}},
	}
	prj.DeleteRequirementByID(1)
	if !prj.D.Requirements[0].Condition.Deleted {
		t.Fatalf("requirement not deleted")
	}
	prj.RestoreRequirementByID(1)
	if prj.D.Requirements[0].Condition.Deleted {
		t.Fatalf("requirement not restored")
	}
}
