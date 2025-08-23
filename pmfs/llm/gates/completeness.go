package gates

func init() {
	register(Gate{
		ID:       "completeness-1",
		Question: "Does the requirement include all necessary conditions and context?",
		FollowUp: "Add any missing conditions or context needed to understand the requirement.",
	})
	register(Gate{
		ID:       "completeness-2",
		Question: "Are all actors and data elements referenced in the requirement defined elsewhere?",
		FollowUp: "Define each actor or data element referenced in the requirement.",
	})
}
