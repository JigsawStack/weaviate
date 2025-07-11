package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type EmbeddingRequest struct {
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type"`
}

type EmbeddingResponse struct {
	Success    bool        `json:"success"`
	Embeddings [][]float32 `json:"embeddings"`
}

type TestResults struct {
	Timestamp  string         `json:"timestamp"`
	Endpoint   string         `json:"endpoint"`
	TextTest   EndpointTest   `json:"textTest"`
	ImageTest  EndpointTest   `json:"imageTest"`
}

type EndpointTest struct {
	Request      interface{}   `json:"request"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
	SampleVector []float32     `json:"sampleVector,omitempty"`
	VectorSize   int           `json:"vectorSize,omitempty"`
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("JIGSAWSTACK_APIKEY")
	if apiKey == "" {
		fmt.Println("Error: JIGSAWSTACK_APIKEY environment variable is not set")
		fmt.Println("Please set the environment variable with: export JIGSAWSTACK_APIKEY=your_api_key")
		os.Exit(1)
	}
	
	baseURL := "https://api.jigsawstack.com"
	endpoint := "/v1/embedding"
	
	results := TestResults{
		Timestamp: time.Now().Format(time.RFC3339),
		Endpoint:  baseURL + endpoint,
	}
	
	// Test text embedding
	textRequest := EmbeddingRequest{
		Text: "Artificial intelligence is transforming the world today",
		Type: "text",
	}
	
	fmt.Println("Testing text embedding...")
	textEndpointTest := testEndpoint(baseURL+endpoint, apiKey, textRequest)
	results.TextTest = textEndpointTest
	
	// Test image embedding
	imageRequest := EmbeddingRequest{
		URL:  "https://images.pexels.com/photos/267569/pexels-photo-267569.jpeg",
		Type: "image",
	}
	
	fmt.Println("Testing image embedding...")
	imageEndpointTest := testEndpoint(baseURL+endpoint, apiKey, imageRequest)
	results.ImageTest = imageEndpointTest
	
	// Write results to file
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding results: %v\n", err)
		return
	}
	
	err = os.WriteFile("api_test_results.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing results to file: %v\n", err)
		return
	}
	
	fmt.Println("Test results written to api_test_results.json")
}

func testEndpoint(url string, apiKey string, request EmbeddingRequest) EndpointTest {
	test := EndpointTest{
		Request: request,
	}
	
	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Failed to encode request: %v", err)
		return test
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		return test
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	
	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Request failed: %v", err)
		return test
	}
	defer resp.Body.Close()
	
	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Failed to read response: %v", err)
		return test
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("API returned non-OK status: %d - %s", resp.StatusCode, string(bodyBytes))
		return test
	}
	
	// Parse response
	var response EmbeddingResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Failed to parse response: %v", err)
		return test
	}
	
	// Check if successful
	if !response.Success {
		test.Success = false
		test.ErrorMessage = "API returned unsuccessful response"
		return test
	}
	
	// Extract embeddings
	if len(response.Embeddings) == 0 || len(response.Embeddings[0]) == 0 {
		test.Success = false
		test.ErrorMessage = "API returned empty embeddings"
		return test
	}
	
	// Set results
	test.Success = true
	
	// Include sample of the vector
	sampleSize := 5
	if len(response.Embeddings[0]) < sampleSize {
		sampleSize = len(response.Embeddings[0])
	}
	test.SampleVector = response.Embeddings[0][:sampleSize]
	test.VectorSize = len(response.Embeddings[0])
	
	fmt.Printf("Successfully obtained %d-dimensional vector\n", test.VectorSize)
	fmt.Printf("Sample: %v\n", test.SampleVector)
	
	return test
} 