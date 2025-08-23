package prompts

// salesPrompts support sales teams in planning strategies and tools.
var salesPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What is the sales strategy for this product?",
		FollowUp: "Which channels will be prioritized?",
	},
	{
		ID:       "2",
		Question: "How will you handle customer objections?",
		FollowUp: "What resources do you need to address them?",
	},
	{
		ID:       "3",
		Question: "What tools will support the sales team?",
		FollowUp: "How will you measure their effectiveness?",
	},
}
