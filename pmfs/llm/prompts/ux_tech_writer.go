package prompts

// uxTechWriterPrompts guide creation of clear, user-focused documentation.
var uxTechWriterPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What user documentation is required?",
		FollowUp: "Who is the target audience for this documentation?",
	},
	{
		ID:       "2",
		Question: "How will complex technical concepts be communicated clearly?",
		FollowUp: "What examples will you provide?",
	},
	{
		ID:       "3",
		Question: "How will documentation be maintained over time?",
		FollowUp: "What process will capture updates?",
	},
}
