# Knowledge base chatbot 🤖

Place here, in `kb` folder, your knowledge base files. Each file should be a
`.pdf`. Then execute:

```shell
$ go run main.go
```

The app will ingest automatically all the files in the `kb` folder and will
create a local `simpleVectorIndex` to store documents chunks and embeddings.
Then you can start to chat with the bot and ask questions about the knowledge base.