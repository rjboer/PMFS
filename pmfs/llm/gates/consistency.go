package gates

func init() {
	register(Gate{
		ID:       "consistency-1",
		Question: "Does the requirement avoid contradicting other requirements?",
		FollowUp: "Resolve any conflicts so this requirement aligns with related requirements.",
	})
	register(Gate{
		ID:       "consistency-2",
		Question: "Is terminology used consistently with other requirements and project documents?",
		FollowUp: "Standardize terminology to match other requirements and project documents.",
	})
}
