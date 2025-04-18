# Multi2Vec-JigsawStack

This module integrates [JigsawStack's Embedding API](https://jigsawstack.com/docs/api-reference/ai/embedding) with Weaviate, enabling the vectorization of text and images.

## Features

- Text embedding generation
- Image embedding generation
- Support for nearText and nearImage search functionality

## How to use

### Environment variables

The module uses the following environment variable:

- `JIGSAWSTACK_APIKEY`: Your API key for JigsawStack

### Configuration Options

| Option    | Type   | Description                  | Default                       |
| --------- | ------ | ---------------------------- | ----------------------------- |
| `baseURL` | string | JigsawStack API endpoint URL | `https://api.jigsawstack.com` |

### Schema Configuration

When creating a schema, you need to specify which fields should be vectorized using the JigsawStack embeddings. You can do this by adding the `textFields` or `imageFields` options to your schema.

```json
{
  "class": "Article",
  "vectorizer": "multi2vec-jigsawstack",
  "moduleConfig": {
    "multi2vec-jigsawstack": {
      "textFields": ["title", "content"],
      "imageFields": ["image"]
    }
  },
  "properties": [
    {
      "name": "title",
      "dataType": ["text"]
    },
    {
      "name": "content",
      "dataType": ["text"]
    },
    {
      "name": "image",
      "dataType": ["text"]
    }
  ]
}
```

### Search by text similarity

```graphql
{
  Get {
    Article(nearText: { concepts: ["artificial intelligence"] }) {
      title
      content
    }
  }
}
```

### Search by image similarity

```graphql
{
  Get {
    Article(nearImage: { image: "https://example.com/image.jpg" }) {
      title
      image
    }
  }
}
```

## Contributing

If you'd like to contribute to the development of this module, please see the [contributing guidelines](https://github.com/weaviate/weaviate/blob/master/CONTRIBUTE.md).
