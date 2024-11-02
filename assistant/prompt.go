package assistant

const (
	//nolint:lll
	baseRAGPrompt = "Use the following pieces of retrieved context to answer the question.\n\nQuestion: {{.question}}\nContext:\n{{range .results}}{{.}}\n\n{{end}}"
	//nolint:lll
	systemPrompt = "{{if ne .assistantName \"\"}}You name is {{.assistantName}}, {{end}}{{if ne .assistantIdentity \"\"}}you are {{.assistantIdentity}}.{{end}} {{if ne .companyName \"\" }}at {{.companyName}}{{end}}{{if ne .companyDescription \"\" }}, {{.companyDescription}}.{{end}} Your task is to assist humans {{.assistantScope}}."

	defaultAssistantName      = "AI assistant"
	defaultAssistantIdentity  = "a helpful and polite assistant"
	defaultAssistantScope     = "with their questions"
	defaultCompanyName        = ""
	defaultCompanyDescription = ""
)
