package gates

func init() {
	register(Gate{
		ID:       "glossary-gaps-1",
		Question: "Are all terms in the requirement defined in the project glossary?",
		FollowUp: "Define any undefined terms in the project glossary.",
	})
	register(Gate{
		ID:       "glossary-gaps-2",
		Question: "Does the requirement avoid undefined acronyms or abbreviations?",
		FollowUp: "Expand or define acronyms and abbreviations used in the requirement.",
	})
}
