package prompts

// qaLeadPrompts support the Quality Assurance Lead in planning testing
// strategies, automation, and coverage.
var qaLeadPrompts = []Prompt{
	{
		ID:       "1",
		Template: "Given the requirement %s, what testing strategies will you employ for this project?",
		FollowUp: "How will these strategies cover edge cases?",
	},
	{
		ID:       "2",
		Template: "Given the requirement %s, how will automation be integrated into the QA process?",
		FollowUp: "Which tools will you use for automation?",
	},
	{
		ID:       "3",
		Template: "Given the requirement %s, what is the plan for regression testing?",
		FollowUp: "How will you maintain test cases over time?",
	},
}

func init() { RegisterRole("qa_lead", qaLeadPrompts) }
