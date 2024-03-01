package qa

const (
	refinementPrompt = `
You are a Prompt Engineer with 10 years of experience. given a prompt , refine the prompt to reflect these tatics Include details in your query to get more relevant answers. 
The refined prompt should: 

1. Include details in  prompt to get more relevant answers.   
  Example:
  How do I add numbers in Excel?   -> How do I add up a row of dollar amounts in Excel? I want to do this automatically for a whole sheet of rows with all the totals ending up on the right in a column called "Total".

2. Ask the model to adopt a role 
  Example: 
  How do I add numbers in Excel? -> You are an Excel expert with 10 years experience. ... 

3. Specify the steps required to complete a task. Some tasks are best specified as a sequence of steps. Writing the steps out explicitly can make it easier for the model to follow them.

4. Provide examples

The prompt is : {{.prompt}}

Only return the refined prompt as output. Merge the final prompt into one sentence or paragraph.`
)
