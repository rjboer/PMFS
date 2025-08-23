package gates

func init() {
	register(Gate{
		ID:       "over-specification-1",
		Question: "Does the requirement avoid prescribing implementation details?",
		FollowUp: "Remove implementation details to focus on behavior or outcomes.",
	})
	register(Gate{
		ID:       "over-specification-2",
		Question: "Is the requirement stated in terms of what is needed rather than how to achieve it?",
		FollowUp: "Rewrite the requirement to express the need without design decisions.",
	})
}
