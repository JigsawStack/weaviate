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

package modulesjigsawstack

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack/clients"
)

func TestModule(t *testing.T) {
	logger, _ := test.NewNullLogger()
	
	t.Run("module initialization", func(t *testing.T) {
		mod := New()
		assert.NotNil(t, mod)
	})
	
	t.Run("module name", func(t *testing.T) {
		mod := New()
		assert.Equal(t, "multi2vec-jigsawstack", mod.Name())
	})
	
	t.Run("init without API key", func(t *testing.T) {
		mod := New()
		cfg := &models.ModuleConfig{
			Class: "Test",
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JIGSAWSTACK_APIKEY")
	})
	
	t.Run("init with API key", func(t *testing.T) {
		// Save original env value and set it back after test
		origAPIKey := os.Getenv("JIGSAWSTACK_APIKEY")
		defer os.Setenv("JIGSAWSTACK_APIKEY", origAPIKey)
		
		os.Setenv("JIGSAWSTACK_APIKEY", "test-api-key")
		
		mod := New()
		cfg := &models.ModuleConfig{
			Class:               "Test",
			ClassConfig:         map[string]interface{}{},
			ModuleConfig:        map[string]interface{}{},
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.NoError(t, err)
	})
	
	t.Run("default configurations", func(t *testing.T) {
		// Save original env value and set it back after test
		origAPIKey := os.Getenv("JIGSAWSTACK_APIKEY")
		defer os.Setenv("JIGSAWSTACK_APIKEY", origAPIKey)
		
		os.Setenv("JIGSAWSTACK_APIKEY", "test-api-key")
		
		mod := New()
		cfg := &models.ModuleConfig{
			Class:               "Test",
			ClassConfig:         map[string]interface{}{},
			ModuleConfig:        map[string]interface{}{},
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.NoError(t, err)
		
		// Use the default value and timeout for initialization
		assert.NotNil(t, mod.(*Module).vectorizer)
	})
	
	t.Run("override configurations", func(t *testing.T) {
		// Save original env value and set it back after test
		origAPIKey := os.Getenv("JIGSAWSTACK_APIKEY")
		defer os.Setenv("JIGSAWSTACK_APIKEY", origAPIKey)
		
		os.Setenv("JIGSAWSTACK_APIKEY", "test-api-key")
		
		mod := New()
		cfg := &models.ModuleConfig{
			Class:               "Test",
			ClassConfig:         map[string]interface{}{},
			ModuleConfig:        map[string]interface{}{
				"baseURL": "https://custom.jigsawstack.com/api",
				"timeout": float64(60),
			},
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.NoError(t, err)
		
		// Custom values should be used
		assert.NotNil(t, mod.(*Module).vectorizer)
	})
}

func TestVectorize(t *testing.T) {
	logger := logrus.New()
	
	t.Run("with mock client", func(t *testing.T) {
		mod := New()
		mockModule := mod.(*Module)
		mockClient := clients.NewMockClient("test-api-key", 30*time.Second, logger)
		mockModule.vectorizer = mockClient
		
		texts := []string{"hello world"}
		images := []string{"data:image/jpeg;base64,/9j/4AAQSkZJRg=="}
		
		classConfig := moduletools.NewClassBasedModuleConfig(
			nil, "Test", "jigsawstack", nil, "jigsawstack",
		)
		
		result, err := mockModule.vectorizer.Vectorize(context.Background(), texts, images, classConfig)
		require.NoError(t, err)
		require.NotNil(t, result)
		
		// Check vector dimensions based on our mock
		assert.Len(t, result.TextVectors, 1)
		assert.Len(t, result.ImageVectors, 1)
		assert.Len(t, result.TextVectors[0], 5)
		assert.Len(t, result.ImageVectors[0], 5)
	})
}

func TestNearText(t *testing.T) {
	logger, _ := test.NewNullLogger()
	
	t.Run("with mock client", func(t *testing.T) {
		// Save original env value and set it back after test
		origAPIKey := os.Getenv("JIGSAWSTACK_APIKEY")
		defer os.Setenv("JIGSAWSTACK_APIKEY", origAPIKey)
		
		os.Setenv("JIGSAWSTACK_APIKEY", "test-api-key")
		
		mod := New()
		cfg := &models.ModuleConfig{
			Class:        "Test",
			ClassConfig:  map[string]interface{}{},
			ModuleConfig: map[string]interface{}{},
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.NoError(t, err)
		
		// Replace real client with mock
		mockModule := mod.(*Module)
		mockClient := clients.NewMockClient("test-api-key", 30*time.Second, logger)
		mockModule.vectorizer = mockClient
		
		// Test NearTextArguments
		nearTextProvider := mod.(NearTextParamsProvider)
		args := nearTextProvider.GetNearTextArguments()
		require.NotNil(t, args)
		
		// Create a nearText search with parameters
		searchParams := map[string]interface{}{
			"concepts": []string{"hello world"},
		}
		
		params := moduletools.BaseParams{
			Class:      "Test",
			Properties: []string{"text"},
		}
		
		nearText, err := nearTextProvider.NearTextTransformer(context.Background(), searchParams, params)
		require.NoError(t, err)
		
		// Verify output structure
		assert.Len(t, nearText.Vector, 5)
	})
}

func TestNearImage(t *testing.T) {
	logger, _ := test.NewNullLogger()
	
	t.Run("with mock client", func(t *testing.T) {
		// Save original env value and set it back after test
		origAPIKey := os.Getenv("JIGSAWSTACK_APIKEY")
		defer os.Setenv("JIGSAWSTACK_APIKEY", origAPIKey)
		
		os.Setenv("JIGSAWSTACK_APIKEY", "test-api-key")
		
		mod := New()
		cfg := &models.ModuleConfig{
			Class:        "Test",
			ClassConfig:  map[string]interface{}{},
			ModuleConfig: map[string]interface{}{},
		}
		err := mod.Init(context.Background(), cfg, nil, logger)
		require.NoError(t, err)
		
		// Replace real client with mock
		mockModule := mod.(*Module)
		mockClient := clients.NewMockClient("test-api-key", 30*time.Second, logger)
		mockModule.vectorizer = mockClient
		
		// Test NearImageArguments
		nearImageProvider := mod.(NearImageParamsProvider)
		args := nearImageProvider.GetNearImageArguments()
		require.NotNil(t, args)
		
		// Create a nearImage search with parameters
		searchParams := map[string]interface{}{
			"image": "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
		}
		
		params := moduletools.BaseParams{
			Class:      "Test",
			Properties: []string{"image"},
		}
		
		nearImage, err := nearImageProvider.NearImageTransformer(context.Background(), searchParams, params)
		require.NoError(t, err)
		
		// Verify output structure 
		assert.Len(t, nearImage.Vector, 5)
	})
} 