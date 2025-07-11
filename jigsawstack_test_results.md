# JigsawStack Module Test Results

## Configuration

- API Endpoint: `https://api.jigsawstack.com/v1/embedding`
- Base URL: `https://api.jigsawstack.com`

## Text Vectorization

- **Input**: "Artificial intelligence is transforming the world today"
- **Request Body**: `{"text":"Artificial intelligence is transforming the world today","type":"text"}`
- **Result**: Success
- **Vector First 5 Values**: [0.024168812, 0.0019911805, -0.0571568, 0.013160101, 0.0023459313]
- **Vector Dimension**: 768

## Image Vectorization

- **Input URL**: "https://images.pexels.com/photos/267569/pexels-photo-267569.jpeg"
- **Request Body**: `{"url":"https://images.pexels.com/photos/267569/pexels-photo-267569.jpeg","type":"image"}`
- **Result**: Success
- **Vector First 5 Values**: [-0.053592477, 0.012322302, 0.012696167, -0.047702186, -0.036593042]
- **Vector Dimension**: 768

## Object Vectorization

- **Input Object**: Object with text and image properties
- **Result**: Returned an empty vector

## Summary

The test completed successfully. Both text and image vectorization are working correctly with the JigsawStack API. Object vectorization needs further investigation.
