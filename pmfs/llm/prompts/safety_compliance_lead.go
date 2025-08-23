package prompts

// safetyComplianceLeadPrompts help ensure regulatory and safety requirements are met.
var safetyComplianceLeadPrompts = []Prompt{
	{
		ID:       "1",
		Question: "Which regulations apply to this project?",
		FollowUp: "How will you ensure adherence to these regulations?",
	},
	{
		ID:       "2",
		Question: "What safety risks have been identified?",
		FollowUp: "What mitigation strategies will be implemented?",
	},
	{
		ID:       "3",
		Question: "How will compliance be monitored over time?",
		FollowUp: "Who is responsible for ongoing audits?",
	},
}
