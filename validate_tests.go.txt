package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ollama/ollama/api"
)

type TestCase struct {
	SuiteName string
	TestName  string
	FilePath  string
	Content   string
}

type TestValidator struct {
	client *api.Client // Assuming you have an LLM client interface
	rules  *Rules
	tests  map[string][]TestCase // key: suite_name, value: test cases
}

func NewTestValidator(client *api.Client, rules *Rules) *TestValidator {
	return &TestValidator{
		client: client,
		rules:  rules,
		tests:  make(map[string][]TestCase),
	}
}

func (tv *TestValidator) ProcessFiles(files map[string]string) error {
	// Step 1: Read all test files and extract test cases
	err := tv.extractTestCases(files)
	if err != nil {
		return fmt.Errorf("failed to extract test cases: %w", err)
	}

	// Step 2: Detect and remove duplicates
	tv.removeDuplicates()

	// Step 3: Validate and improve test quality
	err = tv.validateAndImproveTests()
	if err != nil {
		return fmt.Errorf("failed to validate and improve tests: %w", err)
	}

	return nil
}

func (tv *TestValidator) extractTestCases(files map[string]string) error {
	for filePath, content := range files {
		// Only process C++ test files
		if !tv.isTestFile(filePath) {
			continue
		}

		testCases := tv.parseTestCases(filePath, content)
		for _, testCase := range testCases {
			tv.tests[testCase.SuiteName] = append(tv.tests[testCase.SuiteName], testCase)
		}
	}
	return nil
}

func (tv *TestValidator) isTestFile(filePath string) bool {
	// Check if file is in tests directory and has appropriate extension
	if !strings.Contains(filePath, tv.rules.Paths.TestsDir) {
		return false
	}

	ext := filepath.Ext(filePath)
	return ext == ".cpp" || ext == ".cc" || ext == ".cxx"
}

func (tv *TestValidator) parseTestCases(filePath, content string) []TestCase {
	var testCases []TestCase

	// Regex patterns for different test frameworks
	var testPattern *regexp.Regexp
	var suitePattern *regexp.Regexp

	switch tv.rules.TestFramework {
	case "gtest", "googletest":
		testPattern = regexp.MustCompile(`TEST\s*\(\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
		suitePattern = regexp.MustCompile(`TEST_F\s*\(\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
	case "catch2":
		testPattern = regexp.MustCompile(`TEST_CASE\s*\(\s*"([^"]+)"\s*(?:,\s*"([^"]+)")?\s*\)`)
	default:
		// Default to gtest pattern
		testPattern = regexp.MustCompile(`TEST\s*\(\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
		suitePattern = regexp.MustCompile(`TEST_F\s*\(\s*([^,]+)\s*,\s*([^)]+)\s*\)`)
	}

	// Find all TEST matches
	matches := testPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			testCase := TestCase{
				SuiteName: strings.TrimSpace(match[1]),
				TestName:  strings.TrimSpace(match[2]),
				FilePath:  filePath,
				Content:   tv.extractTestContent(content, match[0]),
			}
			testCases = append(testCases, testCase)
		}
	}

	// Find all TEST_F matches (for fixture-based tests)
	if suitePattern != nil {
		fixtureMatches := suitePattern.FindAllStringSubmatch(content, -1)
		for _, match := range fixtureMatches {
			if len(match) >= 3 {
				testCase := TestCase{
					SuiteName: strings.TrimSpace(match[1]),
					TestName:  strings.TrimSpace(match[2]),
					FilePath:  filePath,
					Content:   tv.extractTestContent(content, match[0]),
				}
				testCases = append(testCases, testCase)
			}
		}
	}

	return testCases
}

func (tv *TestValidator) extractTestContent(content, testDeclaration string) string {
	// Find the test declaration and extract the full test body
	startIndex := strings.Index(content, testDeclaration)
	if startIndex == -1 {
		return ""
	}

	// Find the opening brace
	braceIndex := strings.Index(content[startIndex:], "{")
	if braceIndex == -1 {
		return ""
	}

	// Extract the test body by counting braces
	braceCount := 0
	start := startIndex + braceIndex
	end := start

	for i := start; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				end = i + 1
				break
			}
		}
	}

	return content[startIndex:end]
}

func (tv *TestValidator) removeDuplicates() {
	for suiteName, testCases := range tv.tests {
		uniqueTests := make(map[string]TestCase)

		for _, testCase := range testCases {
			key := fmt.Sprintf("%s_%s", testCase.SuiteName, testCase.TestName)

			// If duplicate found, keep the one with better quality
			if existing, exists := uniqueTests[key]; exists {
				if tv.isHigherQuality(testCase, existing) {
					uniqueTests[key] = testCase
				}
			} else {
				uniqueTests[key] = testCase
			}
		}

		// Convert back to slice
		var uniqueTestSlice []TestCase
		for _, testCase := range uniqueTests {
			uniqueTestSlice = append(uniqueTestSlice, testCase)
		}

		tv.tests[suiteName] = uniqueTestSlice
		log.Printf("Suite %s: removed %d duplicates, kept %d tests",
			suiteName, len(testCases)-len(uniqueTestSlice), len(uniqueTestSlice))
	}
}

func (tv *TestValidator) isHigherQuality(test1, test2 TestCase) bool {
	// Quality metrics
	score1 := tv.calculateQualityScore(test1)
	score2 := tv.calculateQualityScore(test2)

	return score1 > score2
}

func (tv *TestValidator) calculateQualityScore(testCase TestCase) int {
	score := 0
	content := testCase.Content

	// Check for assertions
	assertionCount := 0
	for _, assertion := range tv.rules.Assertions.Preferred {
		assertionCount += strings.Count(content, assertion)
	}
	score += assertionCount * 10

	// Check for descriptive test names
	if tv.rules.Naming.DescriptiveTestNames {
		if len(testCase.TestName) > 10 && strings.Contains(testCase.TestName, "_") {
			score += 15
		}
	}

	// Check for positive and negative cases
	if tv.rules.TestCaseRules.IncludePositive {
		if strings.Contains(strings.ToLower(testCase.TestName), "valid") ||
			strings.Contains(strings.ToLower(testCase.TestName), "success") ||
			strings.Contains(strings.ToLower(testCase.TestName), "positive") {
			score += 10
		}
	}

	if tv.rules.TestCaseRules.IncludeNegative {
		if strings.Contains(strings.ToLower(testCase.TestName), "invalid") ||
			strings.Contains(strings.ToLower(testCase.TestName), "fail") ||
			strings.Contains(strings.ToLower(testCase.TestName), "negative") ||
			strings.Contains(strings.ToLower(testCase.TestName), "error") {
			score += 10
		}
	}

	// Check for setup/teardown
	if strings.Contains(content, "SetUp") || strings.Contains(content, "TearDown") {
		score += 5
	}

	// Penalize for edge cases we want to avoid
	for _, edgeCase := range tv.rules.TestCaseRules.AvoidEdgeCases {
		if strings.Contains(strings.ToLower(content), strings.ToLower(edgeCase)) {
			score -= 20
		}
	}

	return score
}

func (tv *TestValidator) validateAndImproveTests() error {
	for suiteName, testCases := range tv.tests {
		log.Printf("Processing test suite: %s with %d tests", suiteName, len(testCases))

		// Check if test count meets requirements
		if len(testCases) < tv.rules.TestCaseRules.PerMethod {
			log.Printf("Suite %s has insufficient tests: %d < %d",
				suiteName, len(testCases), tv.rules.TestCaseRules.PerMethod)
		}

		// Generate improved tests using LLM
		improvedTests, err := tv.GenerateImprovedTests(testCases, tv.rules)
		if err != nil {
			log.Printf("Failed to generate improved tests for suite %s: %v", suiteName, err)
			continue
		}

		// Write improved tests to file
		err = tv.writeImprovedTests(suiteName, improvedTests)
		if err != nil {
			log.Printf("Failed to write improved tests for suite %s: %v", suiteName, err)
			continue
		}

		log.Printf("Successfully improved tests for suite: %s", suiteName)
	}

	return nil
}

func (tv *TestValidator) writeImprovedTests(suiteName, improvedTests string) error {
	// Create output file path
	outputFile := filepath.Join(tv.rules.Paths.TestsDir, fmt.Sprintf("%s_improved_tests.cpp", suiteName))

	// Write to file (assuming you have a WriteFile function)
	return WriteFile(outputFile, improvedTests)
}

func (tv *TestValidator) GetTestSummary() map[string]int {
	summary := make(map[string]int)
	for suiteName, testCases := range tv.tests {
		summary[suiteName] = len(testCases)
	}
	return summary
}

// Helper functions that you'll need to implement based on your existing codebase
func WriteFile(filePath, content string) error {
	// Implementation depends on your existing file writing logic
	panic("implement WriteFile function")
}
