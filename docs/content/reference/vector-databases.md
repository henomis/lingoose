---
linkTitle: Vector Databases
title: Vector Database Integration
description: Guide to using vector databases for embeddings and semantic search with LinGoose
menu: main
weight: -1
---

LinGoose provides seamless integration with multiple vector databases for storing and querying embeddings. This enables powerful semantic search, recommendation systems, and Retrieval Augmented Generation (RAG) applications.

## Overview

Vector databases store high-dimensional vectors (embeddings) that represent the semantic meaning of text, images, or other data. LinGoose supports both cloud-hosted and self-hosted vector database solutions.

### Supported Vector Databases

- **Pinecone** - Managed vector database service
- **Qdrant** - Open-source vector search engine  
- **Redis** - In-memory database with vector search
- **Milvus** - Open-source vector database
- **PostgreSQL** - Traditional database with pgvector extension
- **JSON DB** - Simple file-based storage for development

## Basic Pattern

All vector database integrations follow a similar pattern:

```go
// 1. Create embedder
embedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)

// 2. Create vector database instance
vectorDB := someVectorDB.New(options)

// 3. Create index combining database and embedder
index := index.New(vectorDB, embedder)

// 4. Load documents
index.LoadFromDocuments(ctx, documents)

// 5. Query for similar content
results, err := index.Query(ctx, "query text", indexoption.WithTopK(5))
```

## Pinecone

Pinecone is a fully managed vector database optimized for similarity search at scale.

### Setup

```bash
# Sign up at https://pinecone.io and get your API key
export PINECONE_API_KEY="your-api-key"
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	pineconedb "github.com/henomis/lingoose/index/vectordb/pinecone"
	"github.com/henomis/lingoose/document"
)

func main() {
	// Configure Pinecone index
	replicas := 1
	shards := 1
	
	vectorDB := pineconedb.New(pineconedb.Options{
		IndexName: "my-knowledge-base",
		Namespace: "documents",
		CreateIndexOptions: &pineconedb.CreateIndexOptions{
			Dimension: 1536, // Must match embedding model dimension
			Metric:    "cosine",
			Pod: &pineconedb.Pod{
				Replicas:    &replicas,
				Shards:      &shards,
				PodType:     "s1.x1",
				Environment: "gcp-starter",
			},
		},
	})

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Add documents
	documents := []document.Document{
		{
			Content: "LinGoose is a Go framework for building AI applications",
			Metadata: map[string]interface{}{
				"source": "documentation",
				"type":   "description",
			},
		},
	}

	err := index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		panic(err)
	}

	// Query similar content
	results, err := index.Query(
		context.Background(),
		"What is LinGoose?",
		indexoption.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

### Advanced Pinecone Configuration

```go
// Create index with custom configuration
vectorDB := pineconedb.New(pineconedb.Options{
	IndexName: "advanced-index",
	Namespace: "specialized-docs",
	CreateIndexOptions: &pineconedb.CreateIndexOptions{
		Dimension: 1536,
		Metric:    "dotproduct", // or "euclidean"
		Pod: &pineconedb.Pod{
			Replicas:    &replicas,
			Shards:      &shards,
			PodType:     "p1.x1", // Performance optimized
			Environment: "us-west1-gcp",
		},
	},
}).WithAPIKey("custom-api-key")
```

## Qdrant

Qdrant is an open-source vector search engine with advanced filtering capabilities.

### Setup

```bash
# Run Qdrant locally with Docker
docker run --rm -p 6333:6333 qdrant/qdrant

# Or use Qdrant Cloud
# Sign up at https://cloud.qdrant.io
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	qdrantdb "github.com/henomis/lingoose/index/vectordb/qdrant"
	"github.com/henomis/lingoose/document"
)

func main() {
	vectorDB := qdrantdb.New(qdrantdb.Options{
		CollectionName: "knowledge-base",
		CreateCollection: &qdrantdb.CreateCollectionOptions{
			Dimension: 1536,
			Distance:  qdrantdb.DistanceCosine,
		},
	}).WithAPIKeyAndEndpoint("", "http://localhost:6333")

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Check if collection is empty and add data if needed
	isEmpty, err := index.IsEmpty(context.Background())
	if err != nil {
		panic(err)
	}

	if isEmpty {
		documents := []document.Document{
			{
				Content: "Qdrant supports advanced filtering and hybrid search",
				Metadata: map[string]interface{}{
					"category": "features",
					"level":    "advanced",
				},
			},
		}

		err = index.LoadFromDocuments(context.Background(), documents)
		if err != nil {
			panic(err)
		}
	}

	// Query with filtering
	results, err := index.Query(
		context.Background(),
		"vector database features",
		indexoption.WithTopK(5),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

### Qdrant Cloud Configuration

```go
// For Qdrant Cloud
vectorDB := qdrantdb.New(qdrantdb.Options{
	CollectionName: "my-collection",
	CreateCollection: &qdrantdb.CreateCollectionOptions{
		Dimension: 1536,
		Distance:  qdrantdb.DistanceCosine,
	},
}).WithAPIKeyAndEndpoint("your-api-key", "https://your-cluster.qdrant.io")
```

## Redis

Redis with the RediSearch module provides vector similarity search capabilities.

### Setup

```bash
# Run Redis Stack (includes RediSearch)
docker run -d -p 6379:6379 redis/redis-stack-server:latest

# Or use Redis Cloud with vector search enabled
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	redisdb "github.com/henomis/lingoose/index/vectordb/redis"
	"github.com/henomis/lingoose/document"
)

func main() {
	vectorDB := redisdb.New(redisdb.Options{
		IndexName: "doc-index",
		CreateIndexOptions: &redisdb.CreateIndexOptions{
			Dimension: 1536,
			Distance:  redisdb.DistanceCosine,
		},
	}).WithAddr("localhost:6379").WithPassword("") // No password for local Redis

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Add documents
	documents := []document.Document{
		{
			Content: "Redis provides fast in-memory vector search",
			Metadata: map[string]interface{}{
				"type":      "database",
				"speed":     "fast",
				"storage":   "memory",
			},
		},
	}

	err := index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		panic(err)
	}

	// Search
	results, err := index.Query(
		context.Background(),
		"fast database search",
		indexoption.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

## Milvus

Milvus is an open-source vector database built for scalable similarity search.

### Setup

```bash
# Run Milvus with Docker Compose
# Download docker-compose.yml from Milvus documentation
docker-compose up -d

# Or use Zilliz Cloud (managed Milvus)
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	milvusdb "github.com/henomis/lingoose/index/vectordb/milvus"
	"github.com/henomis/lingoose/document"
)

func main() {
	vectorDB := milvusdb.New(milvusdb.Options{
		CollectionName: "documents",
		CreateCollectionOptions: &milvusdb.CreateCollectionOptions{
			Dimension: 1536,
			Metric:    milvusdb.MetricInnerProduct,
		},
	}).WithConnection("localhost:19530")

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Add documents
	documents := []document.Document{
		{
			Content: "Milvus offers high-performance vector similarity search",
			Metadata: map[string]interface{}{
				"database": "milvus",
				"feature":  "performance",
			},
		},
	}

	err := index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		panic(err)
	}

	// Search
	results, err := index.Query(
		context.Background(),
		"high performance search",
		indexoption.WithTopK(5),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

## PostgreSQL with pgvector

Use PostgreSQL with the pgvector extension for vector operations.

### Setup

```bash
# Install PostgreSQL with pgvector
# Or use a managed service like Supabase, Neon, or AWS RDS

# Enable pgvector extension
psql -c "CREATE EXTENSION vector;"
```

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	postgresdb "github.com/henomis/lingoose/index/vectordb/postgres"
	"github.com/henomis/lingoose/document"
)

func main() {
	vectorDB := postgresdb.New(postgresdb.Options{
		TableName: "documents",
		CreateTableOptions: &postgresdb.CreateTableOptions{
			Dimension: 1536,
		},
	}).WithConnectionString("postgres://user:password@localhost/dbname?sslmode=disable")

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Add documents
	documents := []document.Document{
		{
			Content: "PostgreSQL with pgvector provides SQL-based vector operations",
			Metadata: map[string]interface{}{
				"type":     "relational",
				"language": "sql",
			},
		},
	}

	err := index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		panic(err)
	}

	// Search
	results, err := index.Query(
		context.Background(),
		"SQL vector database",
		indexoption.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

## JSON DB (Development)

Simple file-based vector storage for development and testing.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	jsondb "github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/document"
)

func main() {
	vectorDB := jsondb.New().WithPersist("vector_store.json")

	index := index.New(
		vectorDB,
		openaiembedder.New(openaiembedder.AdaEmbeddingV2),
	).WithIncludeContents(true)

	// Add documents
	documents := []document.Document{
		{
			Content: "JSON DB is perfect for prototyping and small applications",
			Metadata: map[string]interface{}{
				"use_case": "development",
				"storage":  "file",
			},
		},
	}

	err := index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		panic(err)
	}

	// Search
	results, err := index.Query(
		context.Background(),
		"prototyping database",
		indexoption.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Score: %.4f\n", result.Score)
		fmt.Printf("Content: %s\n", result.Content())
		fmt.Printf("Metadata: %v\n\n", result.Metadata)
	}
}
```

## Advanced Usage

### Batch Operations

```go
// Load multiple documents efficiently
documents := []document.Document{
	{Content: "Document 1", Metadata: map[string]interface{}{"id": 1}},
	{Content: "Document 2", Metadata: map[string]interface{}{"id": 2}},
	{Content: "Document 3", Metadata: map[string]interface{}{"id": 3}},
}

err := index.LoadFromDocuments(context.Background(), documents)
if err != nil {
	panic(err)
}
```

### Custom Query Options

```go
// Query with various options
results, err := index.Query(
	context.Background(),
	"search query",
	indexoption.WithTopK(10),           // Return top 10 results
	indexoption.WithScoreThreshold(0.7), // Minimum similarity score
	indexoption.WithMetadataFilter(map[string]interface{}{
		"category": "technical",
	}),
)
```

### Using Different Embedders

```go
// Use different embedding models
import (
	cohereembedder "github.com/henomis/lingoose/embedder/cohere"
	huggingfaceembedder "github.com/henomis/lingoose/embedder/huggingface"
	voyageembedder "github.com/henomis/lingoose/embedder/voyage"
)

// Cohere embeddings
index := index.New(
	vectorDB,
	cohereembedder.New(cohereembedder.EmbedEnglishV2_0),
)

// HuggingFace embeddings
index := index.New(
	vectorDB,
	huggingfaceembedder.New().WithModel("sentence-transformers/all-MiniLM-L6-v2"),
)

// Voyage embeddings
index := index.New(
	vectorDB,
	voyageembedder.New(voyageembedder.VoyageCodeGeneralModel),
)
```

## Best Practices

### Choosing a Vector Database

- **Pinecone**: Best for production applications requiring minimal setup and management
- **Qdrant**: Great balance of features, performance, and self-hosting capabilities  
- **Redis**: Ideal when you need vector search alongside traditional caching/database operations
- **Milvus**: Best for large-scale applications requiring high throughput
- **PostgreSQL**: Good choice when integrating with existing PostgreSQL infrastructure
- **JSON DB**: Perfect for development, testing, and small applications

### Performance Optimization

```go
// Enable content inclusion only when needed
index := index.New(vectorDB, embedder).WithIncludeContents(false)

// Use appropriate batch sizes for loading
const batchSize = 100
for i := 0; i < len(documents); i += batchSize {
	end := i + batchSize
	if end > len(documents) {
		end = len(documents)
	}
	batch := documents[i:end]
	err := index.LoadFromDocuments(context.Background(), batch)
	if err != nil {
		return err
	}
}
```

### Error Handling

```go
func queryWithRetry(index *index.Index, query string) ([]types.Similarity, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		results, err := index.Query(
			context.Background(),
			query,
			indexoption.WithTopK(5),
		)
		if err == nil {
			return results, nil
		}
		
		// Log error and retry
		fmt.Printf("Query attempt %d failed: %v\n", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return nil, fmt.Errorf("query failed after %d retries", maxRetries)
}
```

Vector databases are essential for building semantic search and RAG applications. Choose the right database based on your scalability needs, hosting preferences, and integration requirements.