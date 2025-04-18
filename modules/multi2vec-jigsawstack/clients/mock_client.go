//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package clients

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/usecases/modulecomponents"
)

type MockClient struct {
	apiKey     string
	httpClient *MockHttpClient
	logger     logrus.FieldLogger
}

type MockHttpClient struct {
	timeout time.Duration
}

func NewMockClient(apiKey string, timeout time.Duration, logger logrus.FieldLogger) *MockClient {
	return &MockClient{
		apiKey:     apiKey,
		httpClient: &MockHttpClient{timeout: timeout},
		logger:     logger,
	}
}

func (m *MockClient) Vectorize(ctx context.Context, texts, images []string, cfg moduletools.ClassConfig) (*modulecomponents.VectorizationCLIPResult[[]float32], error) {
	// Return fixed test vectors
	textVectors := make([][]float32, len(texts))
	for i := range textVectors {
		textVectors[i] = []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	}

	imageVectors := make([][]float32, len(images))
	for i := range imageVectors {
		imageVectors[i] = []float32{0.6, 0.7, 0.8, 0.9, 1.0}
	}

	return &modulecomponents.VectorizationCLIPResult[[]float32]{
		TextVectors:  textVectors,
		ImageVectors: imageVectors,
	}, nil
}

func (m *MockClient) VectorizeQuery(ctx context.Context, input []string, cfg moduletools.ClassConfig) (*modulecomponents.VectorizationCLIPResult[[]float32], error) {
	return m.Vectorize(ctx, input, nil, cfg)
}

func (m *MockClient) MetaInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":    "MockJigsawStack",
		"version": "test",
	}, nil
} 