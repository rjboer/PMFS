package gates

func init() {
	register(Gate{
		ID:       "duplicate-1",
		Question: "Does the requirement avoid duplicating existing requirements?",
		FollowUp: "Consolidate duplicate requirements or remove redundant statements.",
	})
	register(Gate{
		ID:       "duplicate-2",
		Question: "Is this requirement uniquely distinguishable from others?",
		FollowUp: "Merge overlapping requirements into a single, distinct statement.",
	})
}
