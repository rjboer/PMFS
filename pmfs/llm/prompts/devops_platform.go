package prompts

// devOpsPlatformPrompts assist in planning deployment and infrastructure reliability.
var devOpsPlatformPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What deployment pipeline will be used?",
		FollowUp: "How will you ensure pipeline reliability?",
	},
	{
		ID:       "2",
		Question: "How will infrastructure be provisioned?",
		FollowUp: "What automation tools will manage it?",
	},
	{
		ID:       "3",
		Question: "What monitoring will be in place?",
		FollowUp: "Which alerts are considered critical?",
	},
}
