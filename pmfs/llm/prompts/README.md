# Prompts

This package contains role-specific prompts used by the PMFS library.

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

