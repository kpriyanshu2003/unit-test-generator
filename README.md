# Unit Test Generator

An intelligent, LLM-powered unit test generation tool that automatically creates, verifies, and runs comprehensive test suites for your codebase with coverage analysis.

## Overview

This tool leverages Large Language Models (LLMs) to automatically generate unit tests for your code, then validates and runs them using Docker containerization. The entire process is configuration-driven through a simple `rules.yaml` file, making it easy to customize for different projects and programming languages.

## How It Works

The unit test generator follows a sophisticated 4-stage pipeline:

1. **Code Analysis**: Scans specified directories and reads source files
2. **Test Generation**: Uses LLM to generate comprehensive unit tests
3. **Test Verification**: Validates generated tests through LLM review
4. **Test Execution**: Runs tests in Docker container with coverage analysis

## Features

- **Configuration-Driven**: Everything controlled through `rules.yaml`
- **LLM-Powered**: Supports multiple LLM providers (Ollama, OpenAI)
- **Coverage Analysis**: Automatic test coverage reporting with configurable thresholds
- **Containerized Testing**: Isolated test execution environment
- **Framework Agnostic**: Configurable for different testing frameworks
- **Smart Naming**: Intelligent test naming conventions
- **Customizable Output**: Flexible test file formatting options

## Configuration

All behavior is controlled through the `rules.yaml` file. Here's what you can configure:

### Language & Framework Settings

```yaml
language: "cpp" # Target programming language
framework: "drogon" # Application framework
test_framework: "gtest" # Testing framework
```

### Test Generation Rules

```yaml
test_case_rules:
  per_method: 2 # Tests per method
  total_tests: 4 # Maximum total tests
  include_positive_case: true # Include positive test cases
  include_negative_case: true # Include negative test cases
  avoid_edge_cases: # Edge cases to avoid
    - "INT_MIN"
    - "INT_MAX"
```

### Coverage Requirements

```yaml
coverage:
  minimum_threshold: 80.0 # Minimum coverage percentage
  enabled: true # Enable coverage analysis
```

### LLM Configuration

```yaml
model_config:
  primary_model: "llama3.1:8b" # Primary LLM model
  fallback_models: # Fallback options
    - "gpt-4"
    - "gpt-3.5-turbo"
  max_retries: 3 # Retry attempts
  timeout_minutes: 10 # Request timeout
```

### Project Paths

```yaml
paths:
  codebase_dir: "./orgChartApi" # Source code directory
  tests_dir: "./tests" # Generated tests directory
  temp_dir: "./tmp" # Temporary files
  folders_to_scan: # Directories to analyze
    - "models"
    - "utils"
```

## 🏃Quick Start

1. **Clone the repository**

   ```bash
   git clone https://github.com/kpriyanshu2003/unit-test-generator.git
   cd unit-test-generator
   ```

2. **Configure your project**
   Edit `rules.yaml` to match your project structure and requirements:

   ```yaml
   # Update paths to point to your source code
   paths:
     codebase_dir: "./your-project"
     tests_dir: "./tests"
     folders_to_scan:
       - "src"
       - "lib"

   # Set your preferred LLM
   model_config:
     primary_model: "your-preferred-model"
   ```

3. **Run the generator**

   ```bash
   # The tool will automatically:
   # 1. Scan your specified directories
   # 2. Generate tests using LLM
   # 3. Verify generated tests
   # 4. Run tests with coverage analysis
   go run .
   ```

4. **View results**
   - Generated tests: `./tests/` directory
   - Coverage reports: Available in HTML format
   - Test execution logs: Console output

## Supported Configurations

### Programming Languages

- C++ (with GTest support)
- Easily extensible for other languages

### Testing Frameworks

- Google Test (GTest)
- Framework-agnostic design for easy extension

### LLM Providers

- **Ollama** (Local models like Llama 3.1)
- **OpenAI** (GPT-4, GPT-3.5)
- **Extensible** for other providers

## Coverage Analysis

The tool provides comprehensive coverage analysis:

- **Threshold Enforcement**: Configurable minimum coverage requirements
- **Multiple Formats**: HTML reports and LCOV format
- **Detailed Metrics**: Line-by-line coverage analysis
- **CI/CD Ready**: Machine-readable output formats

## 🎯 Test Generation Intelligence

The LLM-powered test generation includes:

- **Smart Test Cases**: Positive, negative, and edge case scenarios
- **Naming Conventions**: Descriptive, standardized test names
- **Framework Integration**: Proper assertion usage
- **Code Coverage**: Tests designed to maximize coverage

## 🔧 Customization Options

### Naming Conventions

```yaml
naming:
  test_prefix: "TEST" # Test function prefix
  descriptive_test_names: true # Use descriptive names
  include_class_in_test_name: true # Include class names
```

### Assertion Preferences

```yaml
assertions:
  preferred:
    - "EXPECT_EQ" # Preferred assertion types
  complete_braces_required: true # Enforce bracing style
```

### Output Format

```yaml
output_format:
  file_type: ".cpp" # Test file extension
  markdown_code_fences: false # Include markdown formatting
  extra_text: false # Minimize extra text
  example_in_prompt: true # Include examples in LLM prompts
```

## Advanced Usage

### Custom Method Selection

```yaml
methods_to_test:
  source: "dynamic" # Auto-discover methods
  manual_list: [] # Or specify manually
```

### LLM Prompt Customization

```yaml
llm_prompt_guidance:
  role_description: "You are an expert C++ programmer..."
  strict_formatting: true
  example_format_included: true
  code_to_test_in_prompt: true
  avoid_comments_outside_code: true
```

## Benefits

- **Time Saving**: Automates tedious test writing process
- **Comprehensive Coverage**: Ensures thorough test coverage
- **Quality Assurance**: LLM verification reduces test quality issues
- **Reproducible**: Consistent, containerized test environment
- **Measurable**: Clear coverage metrics and reporting
- **Flexible**: Highly configurable for different projects
