---
title: "Observer"
description:
linkTitle: "Observer"
menu: { main: { parent: 'reference', weight: -92 } }
---

## Observer

The `Observer` interface helps to observe, debug, and analyze LLM applications. This component tracks metrics (e.g. LLM cost, latency, quality) and gains insights from external dashboards and data exports. To enable tracing an LLM application, create an observer and pass it to the LLM instance.

### Supported platform

* [Langfuse](https://langfuse.com/)

### Usage

```go

o := langfuse.New(context.Background())
trace, err := o.Trace(&observer.Trace{Name: "Who are you"})
if err != nil {
    panic(err)
}

openaillm := openai.New().WithObserver(o, trace.ID)

t := thread.New().AddMessage(
    thread.NewUserMessage().AddContent(
        thread.NewTextContent("Hello, who are you?"),
    ),
)

err = openaillm.Generate(context.Background(), t)
if err != nil {
    panic(err)
}

o.Flush(context.Background())
```