package gates

func init() {
	register(Gate{
		ID:       "testability-1",
		Question: "Can the requirement be verified through inspection, demonstration, or test?",
		FollowUp: "Rewrite the requirement so that it can be objectively verified.",
	})
	register(Gate{
		ID:       "testability-2",
		Question: "Does the requirement specify measurable criteria for success?",
		FollowUp: "Add quantifiable success criteria to the requirement.",
	})
}
