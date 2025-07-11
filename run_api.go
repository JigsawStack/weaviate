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

type Request struct {
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type"`
}

type Response struct {
	Success    bool          `json:"success"`
	Embeddings [][]float32   `json:"embeddings"`
	Usage      UsageInfo     `json:"_usage"`
	Chunks     []string      `json:"chunks,omitempty"`
}

type UsageInfo struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	InferenceTokens int `json:"inference_time_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

type TestResult struct {
	Timestamp  string        `json:"timestamp"`
	TextTest   ApiTestResult `json:"textTest"`
	ImageTest  ApiTestResult `json:"imageTest"`
}

type ApiTestResult struct {
	RequestData  Request      `json:"request"`
	ResponseData interface{}  `json:"response"`
	VectorSize   int          `json:"vectorSize,omitempty"`
	SampleVector []float32    `json:"sampleVector,omitempty"`
	Status       string       `json:"status"`
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("JIGSAWSTACK_APIKEY")
	if apiKey == "" {
		fmt.Println("Error: JIGSAWSTACK_APIKEY environment variable not set")
		os.Exit(1)
	}
	
	// Initialize test result
	result := TestResult{
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	// Test text embedding
	textReq := Request{
		Text: "Artificial intelligence is transforming the world today",
		Type: "text",
	}
	
	fmt.Println("Testing text embedding...")
	textResult := callAPI(apiKey, textReq)
	result.TextTest = textResult
	
	// Test image embedding
	imageReq := Request{
		URL: "https://images.pexels.com/photos/267569/pexels-photo-267569.jpeg",
		Type: "image",
	}
	
	fmt.Println("Testing image embedding...")
	imageResult := callAPI(apiKey, imageReq)
	result.ImageTest = imageResult
	
	// Write results to file
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %v\n", err)
		return
	}
	
	err = os.WriteFile("jigsawstack_api_test.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
	
	fmt.Println("Results saved to jigsawstack_api_test.json")
}

func callAPI(apiKey string, req Request) ApiTestResult {
	result := ApiTestResult{
		RequestData: req,
		Status:      "failed",
	}
	
	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return result
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.jigsawstack.com/v1/embedding", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return result
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	
	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		return result
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return result
	}
	
	// Parse response
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return result
	}
	
	// Update result
	result.ResponseData = response
	result.Status = "success"
	
	if len(response.Embeddings) > 0 && len(response.Embeddings[0]) > 0 {
		// Get vector size
		vectorSize := len(response.Embeddings[0])
		result.VectorSize = vectorSize
		
		// Get sample vector (first 5 values or fewer if vector is smaller)
		sampleSize := 5
		if vectorSize < sampleSize {
			sampleSize = vectorSize
		}
		result.SampleVector = response.Embeddings[0][:sampleSize]
		
		fmt.Printf("Successfully obtained %d-dimensional vector\n", vectorSize)
		fmt.Printf("Sample: %v\n", result.SampleVector)
	} else {
		fmt.Println("No embeddings returned")
	}
	
	return result
} 