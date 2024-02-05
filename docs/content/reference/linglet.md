---
title: "LinGoose Linglets"
description:
linkTitle: "Linglets"
menu: { main: { parent: 'reference', weight: -89 } }
---

Linglets are pre-built LinGoose Assistants with a specific purpose. They are designed to be used as a starting point for building your own AI app. You can use them as a reference to understand how to build your own assistant.

There are two Linglets available:

- `sql` - A Linglet that can understand and respond to SQL queries.
- `summarize` - A Linglet that can summarize text.

## Using SQL Linglet

The sql Linglet helps to build SQL queries. It can be used to automate tasks, defining a comlex SQL query, and provide information. 

```go
db, err := sql.Open("sqlite3", "Chinook_Sqlite.sqlite")
if err != nil {
    panic(err)
}

lingletSQL := lingletsql.New(
    openai.New().WithMaxTokens(2000).WithTemperature(0).WithModel(openai.GPT3Dot5Turbo16K0613),
    db,
)

result, err := lingletSQL.Run(
    context.Background(),
    "list the top 3 albums that are most frequently present in playlists.",
)
if err != nil {
    panic(err)
}

fmt.Printf("SQL Query\n-----\n%s\n\n", result.SQLQuery)
fmt.Printf("Answer\n-------\n%s\n", result.Answer)
```

## Using Summarize Linglet

The summarize Linglet helps to summarize text. 

```go

textLoader := loader.NewTextLoader("state_of_the_union.txt", nil).
    WithTextSplitter(textsplitter.NewRecursiveCharacterTextSplitter(4000, 0))

summarize := summarize.New(
    openai.New().WithMaxTokens(2000).WithTemperature(0).WithModel(openai.GPT3Dot5Turbo16K0613),
    textLoader,
).WithCallback(
    func(t *thread.Thread, i, n int) {
        fmt.Printf("Progress : %.0f%%\n", float64(i)/float64(n)*100)
    },
)

summary, err := summarize.Run(context.Background())
if err != nil {
    panic(err)
}

fmt.Println(*summary)
```

The summarize linglet chunks the input text into smaller pieces and then iterate over each chunk to summarize the result. It also provides a callback function to track the progress of the summarization process.