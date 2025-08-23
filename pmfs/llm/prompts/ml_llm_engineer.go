package prompts

// mlLlmEngineerPrompts focus on model selection, data, and evaluation for ML/LLM projects.
var mlLlmEngineerPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what machine learning models are planned for use?",
		FollowUp: "Why were these models chosen?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, what data is required for training?",
		FollowUp: "How will data quality be ensured?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, how will model performance be evaluated?",
		FollowUp: "What metrics define success?",
	},
}
