package prompts

// securityPrivacyOfficerPrompts address data protection and security oversight.
var securityPrivacyOfficerPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what data privacy concerns exist for this project?",
		FollowUp: "How will these concerns be addressed?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, what security controls are required?",
		FollowUp: "Which standards guide these controls?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how will incident response be handled?",
		FollowUp: "What is the plan for notifying stakeholders?",
	},
}

func init() { RegisterRole("security_privacy_officer", securityPrivacyOfficerPrompts) }
