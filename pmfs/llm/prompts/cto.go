package prompts

// ctoPrompts contain questions for the Chief Technology Officer, who oversees
// technology strategy and alignment with business goals.
var ctoPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What is the main technical challenge you foresee with this project?",
		FollowUp: "How do you plan to address this challenge?",
	},
	{
		ID:       "2",
		Question: "How will this project align with the overall company strategy?",
		FollowUp: "What metrics will you track to ensure alignment?",
	},
	{
		ID:       "3",
		Question: "What resources are required for successful execution?",
		FollowUp: "Where do you anticipate the most resource risk?",
	},
}
