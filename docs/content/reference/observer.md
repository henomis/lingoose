---
title: "Observe and Analyze LLM Applications"
description:
linkTitle: "Observer"
menu: { main: { parent: 'reference', weight: -92 } }
---

## Observer

The `Observer` interface helps to observe, debug, and analyze LLM applications. This component tracks metrics (e.g. LLM cost, latency, quality) and gains insights from external dashboards and data exports. To enable tracing an LLM application, create an observer and attach it to the context. The observer will then track the application's execution and provide insights.

### Supported platform

* [Langfuse](https://langfuse.com/)

### Usage

```go
ctx := context.Background()

o := langfuse.New(ctx)
trace, err := o.Trace(&observer.Trace{Name: "Who are you"})
if err != nil {
    panic(err)
}

ctx = observer.ContextWithObserverInstance(ctx, o)
ctx = observer.ContextWithTraceID(ctx, trace.ID)

openaillm := openai.New()

t := thread.New().AddMessage(
    thread.NewUserMessage().AddContent(
        thread.NewTextContent("Hello, who are you?"),
    ),
)

err = openaillm.Generate(ctx, t)
if err != nil {
    panic(err)
}

o.Flush(ctx)
```
