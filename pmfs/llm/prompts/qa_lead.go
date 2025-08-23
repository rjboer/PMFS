package prompts

// qaLeadPrompts support the Quality Assurance Lead in planning testing
// strategies, automation, and coverage.
var qaLeadPrompts = []Prompt{
	{
		ID:       "1",
		Question: "What testing strategies will you employ for this project?",
		FollowUp: "How will these strategies cover edge cases?",
	},
	{
		ID:       "2",
		Question: "How will automation be integrated into the QA process?",
		FollowUp: "Which tools will you use for automation?",
	},
	{
		ID:       "3",
		Question: "What is the plan for regression testing?",
		FollowUp: "How will you maintain test cases over time?",
	},
}
