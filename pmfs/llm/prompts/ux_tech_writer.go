package prompts

// uxTechWriterPrompts guide creation of clear, user-focused documentation.
var uxTechWriterPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what user documentation is required?",
		FollowUp: "Who is the target audience for this documentation?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, how will complex technical concepts be communicated clearly?",
		FollowUp: "What examples will you provide?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how will documentation be maintained over time?",
		FollowUp: "What process will capture updates?",
	},
}
