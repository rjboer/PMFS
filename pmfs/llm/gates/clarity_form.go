package gates

func init() {
	register(Gate{
		ID:       "clarity-form-1",
		Question: "Is the requirement written as a complete, grammatically correct sentence?",
		FollowUp: "Rewrite the requirement so it stands alone as a single, complete sentence.",
	})
	register(Gate{
		ID:       "clarity-form-2",
		Question: "Does the requirement avoid subjective or ambiguous terms (e.g., 'fast', 'user-friendly')?",
		FollowUp: "Replace ambiguous terms with specific, measurable language.",
	})
}
