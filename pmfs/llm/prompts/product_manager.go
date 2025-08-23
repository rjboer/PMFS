package prompts

// productManagerPrompts guide product managers in defining vision and roadmap.
var productManagerPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What problem does this product solve for the customer?",
		FollowUp: "How did you validate this problem?",
	},
	{
		ID:       "2",
		Question: "What are the key features for the first release?",
		FollowUp: "How did you prioritize them?",
	},
	{
		ID:       "3",
		Question: "How will feedback be integrated into the roadmap?",
		FollowUp: "Which channels will you use to gather feedback?",
	},
}
