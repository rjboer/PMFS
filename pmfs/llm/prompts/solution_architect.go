package prompts

// solutionArchitectPrompts address the needs of Solution Architects who design
// system architectures that meet requirements for scalability and security.
var solutionArchitectPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what architecture patterns are most suitable for this solution?",
		FollowUp: "Why do these patterns fit the requirements?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, how will you ensure scalability in the design?",
		FollowUp: "Which components are critical for scaling?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how are security concerns integrated into the architecture?",
		FollowUp: "What standards will be applied to ensure security compliance?",
	},
}
