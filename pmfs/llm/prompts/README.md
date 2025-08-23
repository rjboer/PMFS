# Prompts

This package contains role-specific prompts used by the PMFS library.

## Role Descriptions

- **CTO** – Oversees technology strategy and ensures alignment with business objectives.
- **Solution Architect** – Designs system architecture that satisfies scalability and security requirements.
- **QA Lead** – Leads quality assurance efforts, defining test strategies and automation practices.
- **New Business Development** – Identifies new markets and partnerships to drive growth.
- **Sales** – Drives revenue through customer acquisition and retention strategies.
- **Product Manager** – Defines product vision and feature prioritization based on user needs.
- **Safety & Compliance Lead** – Ensures projects adhere to regulatory and safety standards.
- **Security/Privacy Officer** – Manages data protection and privacy compliance measures.
- **ML/LLM Engineer** – Develops and evaluates machine learning and language models.
- **DevOps/Platform** – Maintains deployment pipelines and reliable infrastructure operations.
- **UX/Tech Writer** – Produces clear, user-focused documentation and technical content.

## Adding a New Role

1. Create a new file in this directory named after the role, for example `dev_ops.go`.
2. Define a slice of `Prompt` values for the role:

```go
package prompts

var devOpsPrompts = []Prompt{
    {ID: "1", Question: "...", FollowUp: "..."},
}
```

3. Update the `GetPrompts` function in `prompts.go` to return the slice for the new role.
4. Document the role in the "Role Descriptions" section above.

