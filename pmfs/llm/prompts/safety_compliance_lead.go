package prompts

// safetyComplianceLeadPrompts help ensure regulatory and safety requirements are met.
var safetyComplianceLeadPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, which regulations apply to this project?",
		FollowUp: "How will you ensure adherence to these regulations?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, what safety risks have been identified?",
		FollowUp: "What mitigation strategies will be implemented?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how will compliance be monitored over time?",
		FollowUp: "Who is responsible for ongoing audits?",
	},
}
