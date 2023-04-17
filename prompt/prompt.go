package prompt

type Prompt interface {
	Format(inputs Inputs) (string, error)
	SetPartials(partials *Inputs)
	Save(path string) error
	Load(path string) error
}
