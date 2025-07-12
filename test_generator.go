package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

type TestGenerator struct {
	client *api.Client
	rules  *Rules
}

func NewTestGenerator(client *api.Client, rules *Rules) *TestGenerator {
	return &TestGenerator{client: client, rules: rules}
}

// ProcessFiles processes all files and generates test cases for each
func (tg *TestGenerator) ProcessFiles(files map[string]string) error {
	log.Printf("Starting to process %d files", len(files))

	// Group files by their base name (without extension)
	fileGroups := make(map[string]map[string]string)

	for filename, content := range files {
		// Get base name without extension
		baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

		// Initialize the group if it doesn't exist
		if fileGroups[baseName] == nil {
			fileGroups[baseName] = make(map[string]string)
		}

		// Add file to its group
		fileGroups[baseName][filename] = content
	}

	log.Printf("Grouped files into %d base names", len(fileGroups))

	successCount := 0
	failureCount := 0

	// Process each group
	for baseName, group := range fileGroups {
		log.Printf("Processing group: %s", baseName)

		// Find .cpp/.cc file (implementation)
		var implFile, implContent string
		var headerContent string

		for filename, content := range group {
			if strings.HasSuffix(filename, ".cpp") || strings.HasSuffix(filename, ".cc") {
				implFile = filename
				implContent = content
			} else if strings.HasSuffix(filename, ".h") || strings.HasSuffix(filename, ".hpp") {
				// headerFile = filename
				headerContent = content
			}
		}

		// Only process if we have an implementation file
		if implFile == "" {
			log.Printf("Skipping group %s: no implementation file found", baseName)
			continue
		}

		// Combine header and implementation content
		combinedContent := tg.combineHeaderAndImplementation(headerContent, implContent)

		// Use the implementation file name for generating test filename
		if err := tg.processFile(implFile, combinedContent); err != nil {
			log.Printf("Failed to process group %s: %v", baseName, err)
			failureCount++
			continue
		}

		successCount++
		log.Printf("Successfully processed group: %s", baseName)
	}

	log.Printf("Processing complete. Success: %d, Failures: %d", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("failed to process %d out of %d groups", failureCount, len(fileGroups))
	}

	return nil
}

// combineHeaderAndImplementation combines header and implementation content
func (tg *TestGenerator) combineHeaderAndImplementation(headerContent, implContent string) string {
	var combined strings.Builder

	// Add header content first (if exists)
	if headerContent != "" {
		combined.WriteString("// Header file content:\n")
		combined.WriteString(headerContent)
		combined.WriteString("\n\n")
	}

	// Add implementation content
	if implContent != "" {
		combined.WriteString("// Implementation file content:\n")
		combined.WriteString(implContent)
	}

	return combined.String()
}

// processFile processes a single file and generates its test case
func (tg *TestGenerator) processFile(filename, content string) error {
	// Generate unit tests for the file
	testCode, err := tg.GenerateUnitTests(content, "")
	if err != nil {
		return fmt.Errorf("failed to generate unit tests: %v", err)
	}

	// Generate output filename
	outputFilename := tg.generateTestFilename(filename)
	outputPath := filepath.Join(tg.rules.Paths.TestsDir, outputFilename)

	// Save the generated test code
	if err := tg.saveTestFile(outputPath, testCode); err != nil {
		return fmt.Errorf("failed to save test file: %v", err)
	}

	log.Printf("Generated test file: %s (%d bytes)", outputPath, len(testCode))
	return nil
}

// GenerateUnitTests generates unit tests for the given code
func (tg *TestGenerator) GenerateUnitTests(code string, extraPrompt string) (string, error) {
	log.Printf("Generating unit tests with model %s (code length: %d bytes)",
		tg.rules.ModelConfig.PrimaryModel, len(code))

	// Extract imports from the original code
	originalImports := tg.extractImportsFromCode(code)

	fmt.Println("Original imports extracted:", originalImports)

	// Get available models
	resp, err := tg.client.List(context.Background())
	if err != nil {
		log.Printf("Failed to list models: %v", err)
		return "", err
	}

	// Build list of models to try
	modelsToTry := tg.buildModelList(resp)
	log.Printf("Available models from server: %v", tg.getModelNames(resp.Models))
	log.Printf("Models to try in order: %v", modelsToTry)

	// Get methods to test
	methods := tg.getMethodsToTest()
	methodsList := strings.Join(methods, ", ")

	// Generate prompt with original imports
	prompt := tg.generatePrompt(code, methodsList, extraPrompt, originalImports)
	log.Printf("Sending API request with prompt (%d bytes)", len(prompt))

	// Create base request
	req := api.GenerateRequest{
		Model:  tg.rules.ModelConfig.PrimaryModel,
		Prompt: prompt,
		Options: map[string]interface{}{
			"num_ctx":     4096,
			"num_predict": 1024,
			"temperature": 0.7,
		},
	}

	// Try each model with retries
	return tg.tryModelsWithRetries(req, modelsToTry, methods)
}

// buildModelList builds the list of models to try in order
func (tg *TestGenerator) buildModelList(resp *api.ListResponse) []string {
	var modelsToTry []string

	// Add primary model first
	modelsToTry = append(modelsToTry, tg.rules.ModelConfig.PrimaryModel)

	// Add fallback models
	modelsToTry = append(modelsToTry, tg.rules.ModelConfig.FallbackModels...)

	// Filter to only include available models
	availableModels := make(map[string]bool)
	for _, model := range resp.Models {
		availableModels[model.Name] = true
	}

	var validModels []string
	for _, model := range modelsToTry {
		if availableModels[model] {
			validModels = append(validModels, model)
		}
	}

	return validModels
}

// getModelNames extracts model names from the API response
func (tg *TestGenerator) getModelNames(models []api.ListModelResponse) []string {
	var names []string
	for _, model := range models {
		names = append(names, model.Name)
	}
	return names
}

// getMethodsToTest determines which methods to test based on configuration
func (tg *TestGenerator) getMethodsToTest() []string {
	if tg.rules.MethodsToTest.Source == "manual" {
		return tg.rules.MethodsToTest.ManualList
	}

	// Default methods to test for C++
	return []string{
		"all public methods",
		"constructors",
		"destructors",
		"operators",
		"static methods",
	}
}

// generatePrompt creates the prompt for the LLM with stricter output requirements
func (tg *TestGenerator) generatePrompt(code, methodsList, extraPrompt string, originalImports []string) string {
	var prompt strings.Builder

	// Role description
	if tg.rules.LLMPromptGuidance.RoleDescription != "" {
		prompt.WriteString(tg.rules.LLMPromptGuidance.RoleDescription)
		prompt.WriteString("\n\n")
	}

	// Basic instruction with emphasis on output format
	prompt.WriteString("Generate ONLY the C++ unit test code using ")
	prompt.WriteString(tg.rules.TestFramework)
	prompt.WriteString(" framework. Do not include any explanations, comments, or text outside the code.\n\n")

	// Test requirements
	prompt.WriteString("Requirements:\n")
	prompt.WriteString(fmt.Sprintf("- Use C++ standard: %s\n", tg.rules.Standards.CPPStandard))
	prompt.WriteString(fmt.Sprintf("- Include %d test cases per method\n", tg.rules.TestCaseRules.PerMethod))
	prompt.WriteString(fmt.Sprintf("- Maximum total tests: %d\n", tg.rules.TestCaseRules.TotalTests))

	if tg.rules.TestCaseRules.IncludePositive {
		prompt.WriteString("- Include positive test cases\n")
	}
	if tg.rules.TestCaseRules.IncludeNegative {
		prompt.WriteString("- Include negative test cases\n")
	}

	if len(tg.rules.TestCaseRules.AvoidEdgeCases) > 0 {
		prompt.WriteString("- Avoid these edge cases: ")
		prompt.WriteString(strings.Join(tg.rules.TestCaseRules.AvoidEdgeCases, ", "))
		prompt.WriteString("\n")
	}

	// Original imports from source file
	if len(originalImports) > 0 {
		prompt.WriteString("- Include relevant imports such as header files from original file\n")
		prompt.WriteString("- Additionally, include these imports from the original file: ")
		prompt.WriteString(strings.Join(originalImports, ", "))
		prompt.WriteString("\n")
	}

	// Additional includes from config
	if len(tg.rules.Includes) > 0 {
		prompt.WriteString("- Also include these headers: ")
		prompt.WriteString(strings.Join(tg.rules.Includes, ", "))
		prompt.WriteString("\n")
	}

	// Methods to test
	if methodsList != "" {
		prompt.WriteString("- Focus on testing: ")
		prompt.WriteString(methodsList)
		prompt.WriteString("\n")
	}

	// Strict output format requirements
	prompt.WriteString("\nIMPORTANT OUTPUT REQUIREMENTS:\n")
	prompt.WriteString("- Return ONLY valid C++ test code\n")
	prompt.WriteString("- Do NOT include any explanatory text\n")
	prompt.WriteString("- Do NOT include phrases like 'Here is', 'This test', etc.\n")
	prompt.WriteString("- Start directly with #include statements or TEST macros\n")
	prompt.WriteString("- End with the last closing brace of the test\n")

	if tg.rules.OutputFormat.MarkdownCodeFences {
		prompt.WriteString("- Use markdown code fences (```cpp and ```)\n")
	} else {
		prompt.WriteString("- Do NOT use markdown code fences\n")
	}

	// Add extra prompt if provided
	if extraPrompt != "" {
		prompt.WriteString("\nAdditional requirements:\n")
		prompt.WriteString(extraPrompt)
		prompt.WriteString("\n")
	}

	// Add the code to test
	prompt.WriteString("\nCode to test:\n")
	if tg.rules.OutputFormat.MarkdownCodeFences {
		prompt.WriteString("```cpp\n")
	}
	prompt.WriteString(code)
	if tg.rules.OutputFormat.MarkdownCodeFences {
		prompt.WriteString("\n```")
	}

	// Final instruction
	prompt.WriteString("\n\nOutput only the complete C++ test file code:")
	if tg.rules.OutputFormat.MarkdownCodeFences {
		prompt.WriteString("\n```cpp")
	}

	return prompt.String()
}

// postProcessResponse cleans up the response after receiving it from the LLM
func (tg *TestGenerator) postProcessResponse(response string) string {
	// Remove common explanatory phrases
	explanatoryPhrases := []string{
		"Here is the unit test code",
		"This test file includes",
		"The test file contains",
		"These tests cover",
		"The maximum total tests are",
		"as per the requirement",
		"This covers",
		"The tests include",
	}

	lines := strings.Split(response, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Skip lines that contain explanatory phrases
		skipLine := false
		for _, phrase := range explanatoryPhrases {
			if strings.Contains(line, phrase) {
				skipLine = true
				break
			}
		}

		if !skipLine {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// tryModelsWithRetries tries multiple models with retry logic
func (tg *TestGenerator) tryModelsWithRetries(req api.GenerateRequest, modelsToTry []string, methods []string) (string, error) {
	var lastErr error

	for _, model := range modelsToTry {
		req.Model = model
		log.Printf("Trying model: %s", model)

		// Try with retries for this model
		for attempt := 1; attempt <= tg.rules.ModelConfig.MaxRetries; attempt++ {
			log.Printf("Attempt %d/%d with model %s", attempt, tg.rules.ModelConfig.MaxRetries, model)

			result, err := tg.callModel(req)
			if err == nil {
				log.Printf("Successfully generated tests with model %s on attempt %d", model, attempt)
				return result, nil
			}

			lastErr = err
			log.Printf("Attempt %d failed with model %s: %v", attempt, model, err)

			// Wait before retry (exponential backoff)
			if attempt < tg.rules.ModelConfig.MaxRetries {
				waitTime := time.Duration(attempt) * time.Second
				log.Printf("Waiting %v before retry", waitTime)
				time.Sleep(waitTime)
			}
		}

		log.Printf("All attempts failed for model %s", model)
	}

	return "", fmt.Errorf("failed to generate tests with all models. Last error: %v", lastErr)
}

// callModel makes the actual API call to the model
func (tg *TestGenerator) callModel(req api.GenerateRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(tg.rules.ModelConfig.TimeoutMinutes)*time.Minute)
	defer cancel()

	var result strings.Builder

	err := tg.client.Generate(ctx, &req, func(resp api.GenerateResponse) error {
		result.WriteString(resp.Response)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}

	response := result.String()
	if response == "" {
		return "", fmt.Errorf("empty response from model")
	}

	log.Printf("Raw response length: %d bytes", len(response))

	// Post-process to remove explanatory text
	response = tg.postProcessResponse(response)

	// ALWAYS extract code from markdown, regardless of configuration
	// because LLMs often return markdown even when not requested
	response = tg.extractCodeFromMarkdown(response)

	// Final cleanup
	response = strings.TrimSpace(response)

	// Validate that we have actual C++ code
	if !tg.isValidCppCode(response) {
		return "", fmt.Errorf("response does not contain valid C++ code")
	}

	log.Printf("Final cleaned response length: %d bytes", len(response))
	return response, nil
}

// isValidCppCode performs basic validation that the response contains C++ code
func (tg *TestGenerator) isValidCppCode(code string) bool {
	// Must contain at least one of these C++ patterns
	requiredPatterns := []string{
		"#include",
		"TEST(",
		"EXPECT_",
		"ASSERT_",
	}

	for _, pattern := range requiredPatterns {
		if strings.Contains(code, pattern) {
			return true
		}
	}

	return false
}

// extractCodeFromMarkdown extracts C++ code from markdown code blocks
func (tg *TestGenerator) extractCodeFromMarkdown(content string) string {
	// Debug: log what we're processing
	log.Printf("Extracting code from markdown. Content starts with: %.50s", content)

	// Look for markdown code blocks and extract content between them
	lines := strings.Split(content, "\n")
	var codeLines []string
	insideCodeBlock := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line starts or ends a code block
		if strings.HasPrefix(trimmedLine, "```") {
			if !insideCodeBlock {
				// Starting a code block
				insideCodeBlock = true
				log.Printf("Found code block start at line %d: %s", i, trimmedLine)
				continue
			} else {
				// Ending a code block
				insideCodeBlock = false
				log.Printf("Found code block end at line %d: %s", i, trimmedLine)
				continue
			}
		}

		// If we're inside a code block, collect the line
		if insideCodeBlock {
			codeLines = append(codeLines, line)
		} else if !strings.HasPrefix(trimmedLine, "```") {
			// If we're not in a code block and this isn't a fence marker,
			// it might be plain code without markdown
			codeLines = append(codeLines, line)
		}
	}

	// If we didn't find any code blocks, return the original content
	// after removing any stray markdown fence markers
	if len(codeLines) == 0 {
		log.Printf("No code blocks found, cleaning fence markers from original content")
		result := content
		result = strings.ReplaceAll(result, "```cpp", "")
		result = strings.ReplaceAll(result, "```c++", "")
		result = strings.ReplaceAll(result, "```", "")
		return strings.TrimSpace(result)
	}

	// Join the code lines
	result := strings.Join(codeLines, "\n")

	// Remove leading and trailing empty lines
	result = strings.TrimSpace(result)

	log.Printf("Extracted code length: %d bytes", len(result))
	return result
}

func (tg *TestGenerator) extractImportsFromCode(code string) []string {
	var imports []string
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#include") {
			imports = append(imports, trimmed)
		}
	}

	return imports
}

// generateTestFilename generates the test filename based on the source file, preserving folder structure
func (tg *TestGenerator) generateTestFilename(sourceFile string) string {
	// Get the relative path from the codebase directory
	relPath, err := filepath.Rel(tg.rules.Paths.CodebaseDir, sourceFile)
	if err != nil {
		// If we can't get relative path, just use the base name
		log.Printf("Warning: Could not get relative path for %s: %v", sourceFile, err)
		baseName := filepath.Base(sourceFile)
		return tg.convertToTestFilename(baseName)
	}

	// Get the directory part and filename part
	dir := filepath.Dir(relPath)
	filename := filepath.Base(relPath)

	// Convert filename to test filename
	testFilename := tg.convertToTestFilename(filename)

	// If the file is in a subdirectory, preserve that structure
	if dir != "." {
		return filepath.Join(dir, testFilename)
	}

	return testFilename
}

// convertToTestFilename converts a source filename to test filename
func (tg *TestGenerator) convertToTestFilename(filename string) string {
	if strings.HasSuffix(filename, ".cpp") {
		return strings.Replace(filename, ".cpp", "_test.cc", 1)
	} else if strings.HasSuffix(filename, ".cc") {
		return strings.Replace(filename, ".cc", "_test.cc", 1)
	} else if strings.HasSuffix(filename, ".h") {
		return strings.Replace(filename, ".h", "_test.cc", 1)
	} else if strings.HasSuffix(filename, ".hpp") {
		return strings.Replace(filename, ".hpp", "_test.cc", 1)
	}
	return filename + "_test.cc"
}

// saveTestFile saves the generated test code to a file
func (tg *TestGenerator) saveTestFile(outputPath, testCode string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write the test code to file
	if err := os.WriteFile(outputPath, []byte(testCode), 0644); err != nil {
		return fmt.Errorf("failed to write test file %s: %v", outputPath, err)
	}

	log.Printf("Successfully saved test file: %s", outputPath)
	return nil
}
