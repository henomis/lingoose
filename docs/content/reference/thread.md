---
title: "Your conversation history"
description:
linkTitle: "Thread"
menu: { main: { parent: 'reference', weight: -100 } }
---

# Thread

The `thread` package manages communication or interaction in a threaded conversation model. This could be used in a variety of applications, such as chat applications, or any system that requires structured, multi-role communication.

## Create a Thread

To create a new thread, you need to create an instance of `thread.Thread` and then add messages to it. You can then run the thread to get the AI assistant's response.

```go
myThread := thread.New()
```

## Add Messages to a Thread

You can add messages to a thread using the `AddMessages` method. Each message should have a role and a content.

```go
myThread.AddMessages(
    thread.NewSystemMessage().AddContent(
        thread.NewTextContent("You are a powerful AI assistant."),
    ),
    thread.NewUserMessage().AddContent(
        thread.NewTextContent("what is the purpose of NATO?"),
    ),
)
```
A Message can have different types of roles such as `System`, `Assistant` or `User`. A Message can have different types of content, such as text, image, or when available tool calls.

## Your Thread, your history

Your thread will keep track of all the messages and responses. You can access the thread's history using the `Messages` field. To print the thread's history, you can use the `String` method.

```go
fmt.Println(myThread)
```
