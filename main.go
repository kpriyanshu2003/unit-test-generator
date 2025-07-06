package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting unit test generator")

	// Load rules from YAML file, or use defaults
	rules, err := LoadRules("rules.yaml")
	if err != nil {
		log.Printf("Failed to load rules.yaml, using defaults: %v", err)
		rules = GetDefaultRules()
	}
	log.Printf("Using rules: Language=%s, Framework=%s, Model=%s",
		rules.Language, rules.Framework, rules.ModelConfig.PrimaryModel)

	// TODO:  Load extra prompt if available
	_, err = LoadExtraPrompt("extra_prompt.txt")
	if err != nil {
		log.Printf("Failed to load extra_prompt.txt: %v", err)
	}

	// Initialize Ollama client
	client, err := initializeOllamaClient()
	if err != nil {
		log.Fatalf("Failed to initialize Ollama client: %v", err)
	}

	// Check Ollama server status
	resp, err := client.List(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to Ollama server: %v", err)
	}
	log.Printf("Ollama server running, available models: %v", resp.Models)

	// Read codebase
	files, err := ReadCodebase(rules.Paths.CodebaseDir, rules.Paths.FoldersToScan)
	if err != nil {
		log.Fatalf("Failed to read codebase: %v", err)
	}

	// Create tests directory if it doesn't exist
	if err := os.MkdirAll(rules.Paths.TestsDir, 0755); err != nil {
		log.Fatalf("Failed to create tests directory: %v", err)
	}
	log.Printf("Tests directory ready: %s", rules.Paths.TestsDir)

	generator := NewTestGenerator(client, rules)
	err = generator.ProcessFiles(files)
	if err != nil {
		log.Fatalf("Failed to process files: %v", err)
	}
	log.Println("Unit test generation completed successfully")

	// test the generated tests
}

func initializeOllamaClient() (*api.Client, error) {
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
		log.Println("OLLAMA_HOST not set, using default:", ollamaURL)
	} else {
		log.Println("Using OLLAMA_HOST:", ollamaURL)
	}

	url, err := url.Parse(ollamaURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama URL %s: %v", ollamaURL, err)
	}

	client := api.NewClient(url, http.DefaultClient)
	log.Println("Ollama client initialized")
	return client, nil
}
