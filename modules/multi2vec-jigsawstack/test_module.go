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

// +build integration

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack/clients"
)

// This file is for manual testing purposes only
// It is not run as part of the regular test suite
// To run it manually, use: go run test_module.go

func main() {
	// Set up a log file
	logFile, err := os.Create("jigsawstack_test.log")
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Set your API key for testing, or use the environment variable
	apiKey := os.Getenv("JIGSAWSTACK_APIKEY")
	if apiKey == "" {
		apiKey = "your-test-api-key" // Replace with your API key for testing
	}
	os.Setenv("JIGSAWSTACK_APIKEY", apiKey)

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(logFile)
	logger.Info("Initializing module for testing...")

	// Create the module
	module := modulesjigsawstack.New()

	// Test with mock client
	logger.Info("Testing with mock client...")
	testWithMockClient(module.(*modulesjigsawstack.Module), logger)
	
	// Test with real client if API key is provided
	if apiKey != "your-test-api-key" {
		logger.Info("Testing with real client...")
		testWithRealClient(module.(*modulesjigsawstack.Module), logger)
	} else {
		logger.Info("Skipping real client tests. Set JIGSAWSTACK_APIKEY environment variable to run real tests.")
	}

	fmt.Println("Test completed. Check jigsawstack_test.log for results.")
}

func testWithMockClient(module *modulesjigsawstack.Module, logger *logrus.Logger) {
	// Use mock client
	mockClient := clients.NewMockClient("test-api-key", 30*time.Second, logger)
	module.SetVectorizer(mockClient)

	// Test text vectorization
	texts := []string{"Hello, world!"}
	classConfig := moduletools.NewClassBasedModuleConfig(nil, "TestClass", "", nil, "")
	
	logger.Info("Vectorizing text with mock client...")
	textResult, err := mockClient.Texts(context.Background(), texts, classConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to vectorize text")
	} else {
		logger.Infof("Text vector (first 5 values): %v", textResult[:5])
	}

	// Test image vectorization
	sampleImage := "data:image/jpeg;base64,/9j/4AAQSkZJRg=="
	logger.Info("Vectorizing image with mock client...")
	imageResult, err := mockClient.VectorizeImage(context.Background(), "test-id", sampleImage, classConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to vectorize image")
	} else {
		logger.Infof("Image vector (first 5 values): %v", imageResult[:5])
	}
}

func testWithRealClient(module *modulesjigsawstack.Module, logger *logrus.Logger) {
	// Initialize the module properly
	params := moduletools.NewInitParams(
		func(className string) (bool, error) { return true, nil },
		nil,
		"",
		logger,
		nil,
		moduletools.NewFakeClassLister(),
		&models.ModuleConfig{
			Class:        "TestClass",
			ClassConfig:  map[string]interface{}{},
			ModuleConfig: map[string]interface{}{},
		},
		nil,
	)

	err := module.Init(context.Background(), params)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize module")
		return
	}

	classConfig := moduletools.NewClassBasedModuleConfig(nil, "TestClass", "", nil, "")

	// Test text vectorization
	texts := []string{"Artificial intelligence is transforming the world."}
	logger.Info("Vectorizing text with real client...")
	textResult, err := module.VectorizeInput(context.Background(), texts[0], classConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to vectorize text")
	} else {
		logger.Infof("Text vector (first 5 values): %v", textResult[:5])
		logger.Infof("Vector dimension: %d", len(textResult))
	}

	// Test image vectorization with a sample image URL
	sampleImage := "https://images.pexels.com/photos/1170986/pexels-photo-1170986.jpeg"
	logger.Info("Vectorizing image with real client...")
	obj := &models.Object{
		Properties: map[string]interface{}{
			"image": sampleImage,
		},
	}
	imageResult, _, err := module.VectorizeObject(context.Background(), obj, classConfig)
	if err != nil {
		logger.WithError(err).Error("Failed to vectorize image")
	} else {
		logger.Infof("Image vector (first 5 values): %v", imageResult[:5])
		logger.Infof("Vector dimension: %d", len(imageResult))
	}
} 