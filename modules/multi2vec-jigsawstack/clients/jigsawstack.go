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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack/ent"
	"github.com/weaviate/weaviate/usecases/modulecomponents"
)

const (
	DefaultBaseURL = "https://api.jigsawstack.com"
	DefaultTimeout = 60 * time.Second
)

type vectorizer struct {
	apiKey  string
	httpClient *http.Client
	logger  logrus.FieldLogger
}

func New(apiKey string, timeout time.Duration, logger logrus.FieldLogger) *vectorizer {
	return &vectorizer{
		apiKey:  apiKey,
		httpClient: &http.Client{Timeout: timeout},
		logger:  logger,
	}
}

type InputType string

const (
	Text  InputType = "text"
	Image InputType = "image"
)

type Settings struct {
	Type    InputType
	BaseURL string
}

// JigsawStack API request and response structures
type embeddingRequest struct {
	Text string    `json:"text,omitempty"`
	URL  string    `json:"url,omitempty"`
	Type InputType `json:"type"`
}

type embeddingResponse struct {
	Success    bool          `json:"success"`
	Embeddings [][]float32   `json:"embeddings"`
	UsageInfo  usageInfo     `json:"_usage"`
}

type usageInfo struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	InferenceTokens int `json:"inference_time_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

func (v *vectorizer) Vectorize(ctx context.Context,
	texts, images []string, cfg moduletools.ClassConfig,
) (*modulecomponents.VectorizationCLIPResult[[]float32], error) {
	return v.vectorize(ctx, texts, images, cfg)
}

func (v *vectorizer) VectorizeQuery(ctx context.Context,
	input []string, cfg moduletools.ClassConfig,
) (*modulecomponents.VectorizationCLIPResult[[]float32], error) {
	return v.vectorize(ctx, input, nil, cfg)
}

func (v *vectorizer) vectorize(ctx context.Context,
	texts, images []string, cfg moduletools.ClassConfig,
) (*modulecomponents.VectorizationCLIPResult[[]float32], error) {
	var textVectors [][]float32
	var imageVectors [][]float32
	settings := ent.NewClassSettings(cfg)

	// Process text inputs
	if len(texts) > 0 {
		for _, text := range texts {
			vectors, err := v.getEmbeddings(ctx, text, Settings{
				Type:    Text,
				BaseURL: settings.BaseURL(),
			})
			if err != nil {
				return nil, err
			}
			textVectors = append(textVectors, vectors...)
		}
	}

	// Process image inputs (URLs or base64)
	if len(images) > 0 {
		for _, image := range images {
			vectors, err := v.getEmbeddings(ctx, image, Settings{
				Type:    Image,
				BaseURL: settings.BaseURL(),
			})
			if err != nil {
				return nil, err
			}
			imageVectors = append(imageVectors, vectors...)
		}
	}

	return &modulecomponents.VectorizationCLIPResult[[]float32]{
		TextVectors:  textVectors,
		ImageVectors: imageVectors,
	}, nil
}

func (v *vectorizer) getEmbeddings(ctx context.Context, content string, settings Settings) ([][]float32, error) {
	// Create request body based on content type
	reqBody := embeddingRequest{
		Type: settings.Type,
	}

	if settings.Type == Text {
		reqBody.Text = content
	} else if settings.Type == Image {
		reqBody.URL = content
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	// Create request
	baseURL := settings.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	url := fmt.Sprintf("%s/v1/embedding", baseURL)
	v.logger.Infof("Making API request to: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", v.apiKey)
	v.logger.Infof("Request headers: Content-Type: application/json, x-api-key: %s (truncated)", v.apiKey[:10]+"...")

	// Log request body
	v.logger.Infof("Request body: %s", string(jsonData))

	// Make the request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request error: %d %s - %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	// Parse response
	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	if !result.Success {
		return nil, errors.New("API returned unsuccessful response")
	}

	return result.Embeddings, nil
}

func (v *vectorizer) MetaInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":    "JigsawStack",
		"version": "0.1.0",
	}, nil
} 