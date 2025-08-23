package gates

func init() {
	register(Gate{
		ID:       "compound-1",
		Question: "Does the requirement describe a single action or condition?",
		FollowUp: "Break the requirement into separate atomic statements.",
	})
	register(Gate{
		ID:       "compound-2",
		Question: "Is the requirement free from conjunctions like 'and' or 'or' that imply multiple requirements?",
		FollowUp: "Rewrite to eliminate conjunctions or split into multiple requirements.",
	})
}
