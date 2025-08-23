package prompts

// securityPrivacyOfficerPrompts address data protection and security oversight.
var securityPrivacyOfficerPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What data privacy concerns exist for this project?",
		FollowUp: "How will these concerns be addressed?",
	},
	{
		ID:       "2",
		Question: "What security controls are required?",
		FollowUp: "Which standards guide these controls?",
	},
	{
		ID:       "3",
		Question: "How will incident response be handled?",
		FollowUp: "What is the plan for notifying stakeholders?",
	},
}
