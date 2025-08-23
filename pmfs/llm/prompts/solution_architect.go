package prompts

var solutionArchitectPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What architecture patterns are most suitable for this solution?",
		FollowUp: "Why do these patterns fit the requirements?",
	},
	{
		ID:       "2",
		Question: "How will you ensure scalability in the design?",
		FollowUp: "Which components are critical for scaling?",
	},
}
