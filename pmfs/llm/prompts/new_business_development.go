package prompts

// newBusinessDevelopmentPrompts focus on identifying growth opportunities and partnerships.
var newBusinessDevelopmentPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what new market opportunities does this project target?",
		FollowUp: "How will you validate demand in these markets?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, which partnerships could accelerate business expansion?",
		FollowUp: "What criteria will you use to evaluate potential partners?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how does this initiative support revenue growth?",
		FollowUp: "What metrics will indicate success?",
	},
}

func init() { RegisterRole("new_business_development", newBusinessDevelopmentPrompts) }
