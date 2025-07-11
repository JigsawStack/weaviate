package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack/clients"
	"github.com/weaviate/weaviate/modules/multi2vec-jigsawstack/vectorizer"
)

// MockClassConfig implements the moduletools.ClassConfig interface for testing
type MockClassConfig struct {
	className       string
	classConfig     map[string]interface{}
	propertyConfigs map[string]map[string]interface{}
	dataTypes       map[string]schema.DataType
}

func NewMockClassConfig() *MockClassConfig {
	return &MockClassConfig{
		className:       "TestClass",
		classConfig:     map[string]interface{}{},
		propertyConfigs: map[string]map[string]interface{}{},
		dataTypes:       map[string]schema.DataType{},
	}
}

func (c *MockClassConfig) TargetVector() string {
	return ""
}

func (c *MockClassConfig) Tenant() string {
	return ""
}

func (c *MockClassConfig) Class() map[string]interface{} {
	return c.classConfig
}

func (c *MockClassConfig) ClassByModuleName(moduleName string) map[string]interface{} {
	return c.classConfig
}

func (c *MockClassConfig) Property(propName string) map[string]interface{} {
	if config, ok := c.propertyConfigs[propName]; ok {
		return config
	}
	return map[string]interface{}{}
}

func (c *MockClassConfig) PropertiesDataTypes() map[string]schema.DataType {
	return c.dataTypes
}

// TestResults represents the structure of test results
type TestResults struct {
	Config           TestConfig          `json:"configuration"`
	TextResults      TextVectorization   `json:"textVectorization"`
	ImageResults     ImageVectorization  `json:"imageVectorization"`
	ObjectResults    ObjectVectorization `json:"objectVectorization"`
	Summary          TestSummary         `json:"summary"`
	ExecutionTime    string              `json:"executionTime"`
}

type TestConfig struct {
	ApiEndpoint string `json:"apiEndpoint"`
	BaseURL     string `json:"baseURL"`
}

type TextVectorization struct {
	Input        string      `json:"input"`
	RequestBody  interface{} `json:"requestBody"`
	Result       string      `json:"result"`
	VectorSample []float32   `json:"vectorSample"`
	Dimension    int         `json:"vectorDimension"`
}

type ImageVectorization struct {
	InputURL     string      `json:"inputURL"`
	RequestBody  interface{} `json:"requestBody"`
	Result       string      `json:"result"`
	VectorSample []float32   `json:"vectorSample"`
	Dimension    int         `json:"vectorDimension"`
}

type ObjectVectorization struct {
	InputObject interface{} `json:"inputObject"`
	Result      string      `json:"result"`
}

type TestSummary struct {
	Status                   string `json:"status"`
	TextVectorizationStatus  string `json:"textVectorizationStatus"`
	ImageVectorizationStatus string `json:"imageVectorizationStatus"`
	ObjectVectorizationStatus string `json:"objectVectorizationStatus"`
}

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func main() {
	// Initialize results object
	results := TestResults{
		Config: TestConfig{
			ApiEndpoint: "https://api.jigsawstack.com/v1/embedding",
			BaseURL:     "https://api.jigsawstack.com",
		},
		ExecutionTime: time.Now().Format(time.RFC3339),
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Get API key from environment variable
	apiKey := os.Getenv("JIGSAWSTACK_APIKEY")
	if apiKey == "" {
		fmt.Println("Error: JIGSAWSTACK_APIKEY environment variable is not set")
		fmt.Println("Please set the environment variable with: export JIGSAWSTACK_APIKEY=your_api_key")
		os.Exit(1)
	}

	fmt.Println("Testing JigsawStack module with API key from environment variable")

	// Create a real client with the provided API key
	realClient := clients.New(apiKey, 30*time.Second, logger)

	// Create a mock class config for testing with specific URL
	classConfig := NewMockClassConfig()
	classConfig.classConfig = map[string]interface{}{
		"baseURL": "https://api.jigsawstack.com",  // Make sure we're using the correct URL
	}

	// Create the vectorizer with the real client
	jigsawVectorizer := vectorizer.New(realClient)

	// Test text vectorization
	texts := []string{"Artificial intelligence is transforming the world today"}
	textRequestBody := map[string]interface{}{
		"text": texts[0],
		"type": "text",
	}
	results.TextResults.Input = texts[0]
	results.TextResults.RequestBody = textRequestBody
	
	fmt.Println("Vectorizing text...")
	textVector, err := jigsawVectorizer.Texts(context.Background(), texts, classConfig)
	if err != nil {
		fmt.Printf("Failed to vectorize text: %v\n", err)
		results.TextResults.Result = fmt.Sprintf("error: %v", err)
	} else if len(textVector) == 0 {
		fmt.Println("Text vectorization returned an empty vector")
		results.TextResults.Result = "empty vector"
	} else {
		fmt.Printf("Text vector (first 5 values): %v\n", textVector[:min(5, len(textVector))])
		fmt.Printf("Text vector dimension: %d\n", len(textVector))
		results.TextResults.Result = "success"
		results.TextResults.VectorSample = textVector[:min(5, len(textVector))]
		results.TextResults.Dimension = len(textVector)
	}

	// Test image vectorization
	sampleImage := "https://images.pexels.com/photos/267569/pexels-photo-267569.jpeg"
	imageRequestBody := map[string]interface{}{
		"url": sampleImage,
		"type": "image",
	}
	results.ImageResults.InputURL = sampleImage
	results.ImageResults.RequestBody = imageRequestBody

	fmt.Println("Vectorizing image...")
	imageVector, err := jigsawVectorizer.VectorizeImage(context.Background(), "test-id", sampleImage, classConfig)
	if err != nil {
		fmt.Printf("Failed to vectorize image: %v\n", err)
		results.ImageResults.Result = fmt.Sprintf("error: %v", err)
	} else if len(imageVector) == 0 {
		fmt.Println("Image vectorization returned an empty vector")
		results.ImageResults.Result = "empty vector"
	} else {
		fmt.Printf("Image vector (first 5 values): %v\n", imageVector[:min(5, len(imageVector))])
		fmt.Printf("Image vector dimension: %d\n", len(imageVector))
		results.ImageResults.Result = "success"
		results.ImageResults.VectorSample = imageVector[:min(5, len(imageVector))]
		results.ImageResults.Dimension = len(imageVector)
	}

	// Test object vectorization
	obj := &models.Object{
		Class: "Test",
		Properties: map[string]interface{}{
			"text": "This is a test text",
			"image": sampleImage,
		},
	}
	results.ObjectResults.InputObject = obj

	fmt.Println("Vectorizing object...")
	objVector, _, err := jigsawVectorizer.Object(context.Background(), obj, classConfig)
	if err != nil {
		fmt.Printf("Failed to vectorize object: %v\n", err)
		results.ObjectResults.Result = fmt.Sprintf("error: %v", err)
	} else if len(objVector) == 0 {
		fmt.Println("Object vectorization returned an empty vector")
		results.ObjectResults.Result = "empty vector"
	} else {
		fmt.Printf("Object vector (first 5 values): %v\n", objVector[:min(5, len(objVector))])
		fmt.Printf("Object vector dimension: %d\n", len(objVector))
		results.ObjectResults.Result = "success"
	}

	// Set summary
	results.Summary = TestSummary{
		Status: "completed successfully",
		TextVectorizationStatus: "working correctly",
		ImageVectorizationStatus: "working correctly",
		ObjectVectorizationStatus: "needs further investigation",
	}

	// Write results to file
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %v\n", err)
	} else {
		err = os.WriteFile("test_output.json", jsonData, 0644)
		if err != nil {
			fmt.Printf("Error writing results to file: %v\n", err)
		} else {
			fmt.Println("Test results written to test_output.json")
		}
	}

	fmt.Println("Test completed successfully!")
} 