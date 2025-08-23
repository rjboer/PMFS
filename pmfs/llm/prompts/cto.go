package prompts

// ctoPrompts contain questions for the Chief Technology Officer, who oversees
// technology strategy and alignment with business goals.
var ctoPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what is the main technical challenge you foresee with this project?",
		FollowUp: "How do you plan to address this challenge?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, how will this project align with the overall company strategy?",
		FollowUp: "What metrics will you track to ensure alignment?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, what resources are required for successful execution?",
		FollowUp: "Where do you anticipate the most resource risk?",
	},
}

func init() { RegisterRole("cto", ctoPrompts) }
