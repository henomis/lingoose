package prompt

type Prompt interface {
	Format(promptTemplateInputs Inputs) (string, error)
}
