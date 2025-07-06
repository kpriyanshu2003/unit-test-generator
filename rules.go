package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Rules defines the configuration structure for unit test generation
type Rules struct {
	Language      string `yaml:"language"`
	Framework     string `yaml:"framework"`
	TestFramework string `yaml:"test_framework"`
	Naming        struct {
		TestPrefix             string `yaml:"test_prefix"`
		DescriptiveTestNames   bool   `yaml:"descriptive_test_names"`
		IncludeClassInTestName bool   `yaml:"include_class_in_test_name"`
	} `yaml:"naming"`
	Includes  []string `yaml:"includes"`
	Standards struct {
		CPPStandard string `yaml:"cpp_standard"`
	} `yaml:"standards"`
	TestCaseRules struct {
		PerMethod       int      `yaml:"per_method"`
		TotalTests      int      `yaml:"total_tests"`
		IncludePositive bool     `yaml:"include_positive_case"`
		IncludeNegative bool     `yaml:"include_negative_case"`
		AvoidEdgeCases  []string `yaml:"avoid_edge_cases"`
	} `yaml:"test_case_rules"`
	Assertions struct {
		Preferred              []string `yaml:"preferred"`
		CompleteBracesRequired bool     `yaml:"complete_braces_required"`
	} `yaml:"assertions"`
	MethodsToTest struct {
		Source     string   `yaml:"source"`
		ManualList []string `yaml:"manual_list"`
	} `yaml:"methods_to_test"`
	OutputFormat struct {
		FileType           string `yaml:"file_type"`
		MarkdownCodeFences bool   `yaml:"markdown_code_fences"`
		ExtraText          bool   `yaml:"extra_text"`
		ExampleInPrompt    bool   `yaml:"example_in_prompt"`
	} `yaml:"output_format"`
	LLMPromptGuidance struct {
		RoleDescription       string `yaml:"role_description"`
		StrictFormatting      bool   `yaml:"strict_formatting"`
		ExampleFormatIncluded bool   `yaml:"example_format_included"`
		CodeToTestInPrompt    bool   `yaml:"code_to_test_in_prompt"`
		AvoidCommentsOutside  bool   `yaml:"avoid_comments_outside_code"`
	} `yaml:"llm_prompt_guidance"`
	Coverage struct {
		MinimumThreshold float64 `yaml:"minimum_threshold"`
		Enabled          bool    `yaml:"enabled"`
	} `yaml:"coverage"`
	ModelConfig struct {
		PrimaryModel   string   `yaml:"primary_model"`
		FallbackModels []string `yaml:"fallback_models"`
		MaxRetries     int      `yaml:"max_retries"`
		TimeoutMinutes int      `yaml:"timeout_minutes"`
	} `yaml:"model_config"`
	Paths struct {
		CodebaseDir   string   `yaml:"codebase_dir"`
		TestsDir      string   `yaml:"tests_dir"`
		TempDir       string   `yaml:"temp_dir"`
		FoldersToScan []string `yaml:"folders_to_scan"`
	} `yaml:"paths"`
}

// LoadRules loads configuration from a YAML file
func LoadRules(filePath string) (*Rules, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rules Rules
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

// LoadExtraPrompt loads additional prompt instructions from a file
func LoadExtraPrompt(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// GetDefaultRules returns the default configuration
func GetDefaultRules() *Rules {
	return &Rules{
		Language:      "C++",
		Framework:     "Google Test",
		TestFramework: "gtest",
		Naming: struct {
			TestPrefix             string `yaml:"test_prefix"`
			DescriptiveTestNames   bool   `yaml:"descriptive_test_names"`
			IncludeClassInTestName bool   `yaml:"include_class_in_test_name"`
		}{
			TestPrefix:             "TEST",
			DescriptiveTestNames:   true,
			IncludeClassInTestName: true,
		},
		Includes: []string{
			"#include <gtest/gtest.h>",
			"#include <cmath>",
			"#include <stdexcept>",
			"#include \"example.h\"",
		},
		Standards: struct {
			CPPStandard string `yaml:"cpp_standard"`
		}{
			CPPStandard: "C++17",
		},
		TestCaseRules: struct {
			PerMethod       int      `yaml:"per_method"`
			TotalTests      int      `yaml:"total_tests"`
			IncludePositive bool     `yaml:"include_positive_case"`
			IncludeNegative bool     `yaml:"include_negative_case"`
			AvoidEdgeCases  []string `yaml:"avoid_edge_cases"`
		}{
			PerMethod:       2,
			TotalTests:      4,
			IncludePositive: true,
			IncludeNegative: true,
			AvoidEdgeCases:  []string{"INT_MIN", "INT_MAX"},
		},
		Assertions: struct {
			Preferred              []string `yaml:"preferred"`
			CompleteBracesRequired bool     `yaml:"complete_braces_required"`
		}{
			Preferred:              []string{"EXPECT_EQ", "EXPECT_NE", "EXPECT_TRUE", "EXPECT_FALSE"},
			CompleteBracesRequired: true,
		},
		MethodsToTest: struct {
			Source     string   `yaml:"source"`
			ManualList []string `yaml:"manual_list"`
		}{
			Source:     "manual",
			ManualList: []string{"add", "subtract"},
		},
		OutputFormat: struct {
			FileType           string `yaml:"file_type"`
			MarkdownCodeFences bool   `yaml:"markdown_code_fences"`
			ExtraText          bool   `yaml:"extra_text"`
			ExampleInPrompt    bool   `yaml:"example_in_prompt"`
		}{
			FileType:           "cpp",
			MarkdownCodeFences: false,
			ExtraText:          false,
			ExampleInPrompt:    true,
		},
		LLMPromptGuidance: struct {
			RoleDescription       string `yaml:"role_description"`
			StrictFormatting      bool   `yaml:"strict_formatting"`
			ExampleFormatIncluded bool   `yaml:"example_format_included"`
			CodeToTestInPrompt    bool   `yaml:"code_to_test_in_prompt"`
			AvoidCommentsOutside  bool   `yaml:"avoid_comments_outside_code"`
		}{
			RoleDescription:       "You are an expert C++ programmer tasked with generating unit tests using Google Test for the provided C++ code. Follow these requirements strictly:",
			StrictFormatting:      true,
			ExampleFormatIncluded: true,
			CodeToTestInPrompt:    true,
			AvoidCommentsOutside:  true,
		},
		Coverage: struct {
			MinimumThreshold float64 `yaml:"minimum_threshold"`
			Enabled          bool    `yaml:"enabled"`
		}{
			MinimumThreshold: 80.0,
			Enabled:          true,
		},
		ModelConfig: struct {
			PrimaryModel   string   `yaml:"primary_model"`
			FallbackModels []string `yaml:"fallback_models"`
			MaxRetries     int      `yaml:"max_retries"`
			TimeoutMinutes int      `yaml:"timeout_minutes"`
		}{
			PrimaryModel:   "qwen2.5-coder:7b",
			FallbackModels: []string{},
			MaxRetries:     3,
			TimeoutMinutes: 5,
		},
		Paths: struct {
			CodebaseDir   string   `yaml:"codebase_dir"`
			TestsDir      string   `yaml:"tests_dir"`
			TempDir       string   `yaml:"temp_dir"`
			FoldersToScan []string `yaml:"folders_to_scan"`
		}{
			CodebaseDir: "./codebase",
			TestsDir:    "./tests",
			TempDir:     "",
		},
	}
}
