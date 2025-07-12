package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

type App struct {
	client *api.Client
	rules  *Rules
	debug  bool
}

func main() {
	app := &App{
		debug: os.Getenv("DEBUG") == "true",
	}

	// Configure logging based on debug mode
	if !app.debug {
		log.SetOutput(io.Discard)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if err := app.initialize(); err != nil {
		app.printError("Initialization failed: %v", err)
		os.Exit(1)
	}

	app.runCLI()
}

func (app *App) initialize() error {
	app.printInfo("🔧 Initializing application...")

	// Load rules
	rules, err := LoadRules("rules.yaml")
	if err != nil {
		app.printWarning("Failed to load rules.yaml, using defaults: %v", err)
		rules = GetDefaultRules()
	}
	app.rules = rules

	if app.debug {
		app.printDebug("Using rules: Language=%s, Framework=%s, Model=%s",
			rules.Language, rules.Framework, rules.ModelConfig.PrimaryModel)
	}

	// Load extra prompt if available
	_, err = LoadExtraPrompt("extra_prompt.txt")
	if err != nil && app.debug {
		app.printDebug("Failed to load extra_prompt.txt: %v", err)
	}

	// Initialize Ollama client
	client, err := app.initializeOllamaClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Ollama client: %v", err)
	}
	app.client = client

	// Check Ollama server status
	resp, err := client.List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama server: %v", err)
	}

	if app.debug {
		app.printDebug("Ollama server running, available models: %v", resp.Models)
	}

	app.printSuccess("Application initialized successfully")
	return nil
}

func (app *App) runCLI() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		app.printMenu()

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			app.generateTests()
		case "2":
			app.runTests()
		case "3":
			app.runBuild()
		case "0", "exit", "quit":
			app.printInfo("👋 Goodbye!")
			return
		default:
			app.printWarning("Invalid choice. Please try again.")
		}

		fmt.Println() // Add spacing between operations
	}
}

func (app *App) printMenu() {
	fmt.Println("\nC++ Unit Test Generator")
	fmt.Println("[1] 🏗️  Generate C++ Tests")
	fmt.Println("[2] 🏃 Run Tests")
	fmt.Println("[3] 🔨 Build C++ Project")
	fmt.Println("[0] 🚪 Exit")
	fmt.Print("Enter your choice: ")
}

func (app *App) generateTests() {
	app.printInfo("🏗️  Starting test generation...")

	// Read codebase
	files, err := ReadCodebase(app.rules.Paths.CodebaseDir, app.rules.Paths.FoldersToScan)
	if err != nil {
		app.printError("Failed to read codebase: %v", err)
		return
	}

	// Create tests directory if it doesn't exist
	if err := os.MkdirAll(app.rules.Paths.TestsDir, 0755); err != nil {
		app.printError("Failed to create tests directory: %v", err)
		return
	}

	if app.debug {
		app.printDebug("Tests directory ready: %s", app.rules.Paths.TestsDir)
	}

	// Generate unit tests
	generator := NewTestGenerator(app.client, app.rules)

	startTime := time.Now()
	err = generator.ProcessFiles(files)
	duration := time.Since(startTime)

	if err != nil {
		app.printError("Failed to process files: %v", err)
		return
	}

	app.printSuccess("Test generation completed successfully in %v", duration)
}

func (app *App) runTests() {
	app.printInfo("🏃 Running C++ tests...")

	var cmd *exec.Cmd

	// Check for different C++ test setups
	if _, err := os.Stat("build/tests"); err == nil {
		// Look for test executables in build directory
		cmd = exec.Command("find", "build/tests", "-type", "f", "-executable", "-exec", "{}", ";")
	} else if _, err := os.Stat("CMakeLists.txt"); err == nil {
		// CMake project - try to run ctest
		if _, err := exec.LookPath("ctest"); err == nil {
			cmd = exec.Command("ctest", "--test-dir", "build", "--verbose")
		} else {
			app.printWarning("CMake project detected but ctest not found. Building first...")
			app.buildCMakeProject()
			cmd = exec.Command("ctest", "--test-dir", "build", "--verbose")
		}
	} else if _, err := os.Stat("Makefile"); err == nil {
		// Makefile project - check for test target
		cmd = exec.Command("make", "test")
	} else if _, err := os.Stat("test"); err == nil {
		// Look for test executables in test directory
		cmd = exec.Command("find", "test", "-type", "f", "-executable", "-exec", "{}", ";")
	} else {
		// Look for any test executables in current directory
		cmd = exec.Command("find", ".", "-name", "*test*", "-type", "f", "-executable", "-exec", "{}", ";")
	}

	if cmd == nil {
		app.printWarning("Could not determine C++ test command. Please build tests first.")
		return
	}

	app.printInfo("Executing: %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		app.printError("Tests failed after %v: %v", duration, err)
	} else {
		app.printSuccess("Tests completed successfully in %v", duration)
	}
}

func (app *App) runBuild() {
	app.printInfo("🔨 Building C++ project...")

	var cmd *exec.Cmd

	// Check for C++ build systems
	if _, err := os.Stat("CMakeLists.txt"); err == nil {
		app.buildCMakeProject()
		return
	} else if _, err := os.Stat("Makefile"); err == nil {
		cmd = exec.Command("make", "all")
	} else if _, err := os.Stat("build.sh"); err == nil {
		cmd = exec.Command("./build.sh")
	} else if _, err := os.Stat("configure"); err == nil {
		app.printInfo("Running configure script first...")
		configCmd := exec.Command("./configure")
		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr
		if err := configCmd.Run(); err != nil {
			app.printError("Configure failed: %v", err)
			return
		}
		cmd = exec.Command("make")
	} else {
		// Try to find and compile .cpp files directly
		app.printInfo("No build system found. Attempting direct compilation...")
		app.directCompile()
		return
	}

	if cmd == nil {
		app.printWarning("Could not determine build command for this C++ project")
		return
	}

	app.printInfo("Executing: %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		app.printError("Build failed after %v: %v", duration, err)
	} else {
		app.printSuccess("Build completed successfully in %v", duration)
	}
}

func (app *App) initializeOllamaClient() (*api.Client, error) {
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
		if app.debug {
			app.printDebug("OLLAMA_HOST not set, using default: %s", ollamaURL)
		}
	} else if app.debug {
		app.printDebug("Using OLLAMA_HOST: %s", ollamaURL)
	}

	url, err := url.Parse(ollamaURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama URL %s: %v", ollamaURL, err)
	}

	client := api.NewClient(url, http.DefaultClient)

	if app.debug {
		app.printDebug("Ollama client initialized")
	}

	return client, nil
}

func (app *App) buildCMakeProject() {
	app.printInfo("🏗️  Building CMake project...")

	// Create build directory if it doesn't exist
	if err := os.MkdirAll("build", 0755); err != nil {
		app.printError("Failed to create build directory: %v", err)
		return
	}

	// Configure with CMake
	configCmd := exec.Command("cmake", "..", "-DCMAKE_BUILD_TYPE=Debug")
	configCmd.Dir = "build"
	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr

	app.printInfo("Configuring CMake...")
	if err := configCmd.Run(); err != nil {
		app.printError("CMake configuration failed: %v", err)
		return
	}

	// Build the project
	buildCmd := exec.Command("cmake", "--build", ".")
	buildCmd.Dir = "build"
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	app.printInfo("Building project...")
	startTime := time.Now()
	err := buildCmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		app.printError("Build failed after %v: %v", duration, err)
	} else {
		app.printSuccess("CMake build completed successfully in %v", duration)
	}
}

func (app *App) directCompile() {
	app.printInfo("🔍 Looking for C++ source files...")

	// Find all .cpp files
	findCmd := exec.Command("find", ".", "-name", "*.cpp", "-o", "-name", "*.cc", "-o", "-name", "*.cxx")
	output, err := findCmd.Output()
	if err != nil {
		app.printError("Failed to find C++ files: %v", err)
		return
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 0 || files[0] == "" {
		app.printWarning("No C++ source files found")
		return
	}

	app.printInfo("Found %d C++ files", len(files))

	// Try to compile with g++ (assuming it's available)
	compiler := "g++"
	if _, err := exec.LookPath(compiler); err != nil {
		compiler = "clang++"
		if _, err := exec.LookPath(compiler); err != nil {
			app.printError("No C++ compiler found (tried g++ and clang++)")
			return
		}
	}

	// Create build directory
	if err := os.MkdirAll("build", 0755); err != nil {
		app.printError("Failed to create build directory: %v", err)
		return
	}

	// Compile each file
	for _, file := range files {
		if strings.TrimSpace(file) == "" {
			continue
		}

		// Skip test files for now
		if strings.Contains(file, "test") {
			continue
		}

		outputFile := "build/" + strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(file, ".cpp"), ".cc"), ".cxx")

		compileCmd := exec.Command(compiler, "-std=c++17", "-Wall", "-g", "-o", outputFile, file)
		compileCmd.Stdout = os.Stdout
		compileCmd.Stderr = os.Stderr

		app.printInfo("Compiling %s...", file)
		if err := compileCmd.Run(); err != nil {
			app.printWarning("Failed to compile %s: %v", file, err)
		} else {
			app.printSuccess("Compiled %s successfully", file)
		}
	}
}

func (app *App) printSuccess(format string, args ...interface{}) {
	fmt.Printf("✅ %s\n", fmt.Sprintf(format, args...))
}

func (app *App) printError(format string, args ...interface{}) {
	fmt.Printf("❌ %s\n", fmt.Sprintf(format, args...))
}

func (app *App) printWarning(format string, args ...interface{}) {
	fmt.Printf("⚠️  %s\n", fmt.Sprintf(format, args...))
}

func (app *App) printInfo(format string, args ...interface{}) {
	fmt.Printf("🔵 %s\n", fmt.Sprintf(format, args...))
}

func (app *App) printDebug(format string, args ...interface{}) {
	if app.debug {
		fmt.Printf("🐛 DEBUG: %s\n", fmt.Sprintf(format, args...))
	}
}
