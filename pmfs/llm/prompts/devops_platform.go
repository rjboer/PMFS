package prompts

// devOpsPlatformPrompts assist in planning deployment and infrastructure reliability.
var devOpsPlatformPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what deployment pipeline will be used?",
		FollowUp: "How will you ensure pipeline reliability?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, how will infrastructure be provisioned?",
		FollowUp: "What automation tools will manage it?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, what monitoring will be in place?",
		FollowUp: "Which alerts are considered critical?",
	},
}
