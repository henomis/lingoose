package agent

var promptTemplate string = `Answer the following questions as best you can. Don't resolve calculus. You have access to the following tools:
{{range .tools}}
{{.Name}}: {{.Description}}
{{end}}
Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [{{ range $i, $tool := .tools }}{{ if $i }},{{ end }}{{ $tool.Name }}{{ end }}]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question

Begin!

Question: {{.question}}
Thought:`
