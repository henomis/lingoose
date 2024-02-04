---
title: "Caching LLM responses"
description:
linkTitle: "Cache"
menu: { main: { parent: 'reference', weight: -92 } }
---
 
Caching LLM responses can be a good way to improve the performance of your application. This is especially true when you are using an LLM to generate responses to user messages in real-time. By caching the responses, you can avoid making repeated calls to the LLM, which can be slow and expensive.

LinGoose provides a built-in caching mechanism that you can use to cache LLM responses. The cache needs an Index