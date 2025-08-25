# Full Example

This example demonstrates a complete PMFS flow using the Gemini client.

- **Requires:** `GEMINI_API_KEY`

1. **Analyze attachment** – `gemini.AnalyzeAttachment` extracts potential requirements from `testdata/spec1.txt`.
2. **Store requirements** – The requirements are stored in a project structure using `PMFS.FromGemini` to convert Gemini output.
3. **Run role questions** – Each requirement's description is posed to several roles (`product_manager`, `qa_lead`, `security_privacy_officer`)
   via `interact.RunQuestion`.
4. **Evaluate gates** – Each requirement is checked against quality gates using `gates.Evaluate`.

5. **Interpret output** – The program prints the results, including role agreement, follow-ups, and gate outcomes.

Run the example with:

```bash
go run ./examples/full
```

