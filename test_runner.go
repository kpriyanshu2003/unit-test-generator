package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CheckAndBuildGoogleTest ensures Google Test is properly built
func CheckAndBuildGoogleTest() error {
	fmt.Println("🔧 Setting up Google Test...")

	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	gtestDir := filepath.Join(projectRoot, "external", "googletest")
	buildDir := filepath.Join(gtestDir, "build")

	// Check if we can find the built libraries
	possibleLibPaths := []string{
		filepath.Join(buildDir, "lib", "libgtest.a"),
		filepath.Join(buildDir, "lib", "libgtest_main.a"),
		filepath.Join(buildDir, "googletest", "libgtest.a"),
		filepath.Join(buildDir, "googletest", "libgtest_main.a"),
	}

	libsExist := true
	for _, libPath := range possibleLibPaths[:2] { // Check first two (main lib paths)
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			libsExist = false
			break
		}
	}

	if !libsExist {
		fmt.Println("📦 Building Google Test libraries...")

		// Create build directory
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			return fmt.Errorf("failed to create build directory: %v", err)
		}

		// Run cmake with proper flags
		cmakeCmd := exec.Command("cmake", "..", "-DCMAKE_BUILD_TYPE=Release")
		cmakeCmd.Dir = buildDir
		if output, err := cmakeCmd.CombinedOutput(); err != nil {
			fmt.Printf("❌ CMake failed:\n%s\n", string(output))
			return fmt.Errorf("cmake failed: %v", err)
		}

		// Run make with parallel jobs
		makeCmd := exec.Command("make", "-j4")
		makeCmd.Dir = buildDir
		if output, err := makeCmd.CombinedOutput(); err != nil {
			fmt.Printf("❌ Make failed:\n%s\n", string(output))
			return fmt.Errorf("make failed: %v", err)
		}

		fmt.Println("✅ Google Test built successfully!")
	} else {
		fmt.Println("✅ Google Test libraries found!")
	}

	return nil
}

// FindGoogleTestLibraries locates the Google Test library files
func FindGoogleTestLibraries() (string, string, error) {
	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to get project root: %v", err)
	}

	buildDir := filepath.Join(projectRoot, "external", "googletest", "build")

	// Possible library locations
	libPaths := []string{
		filepath.Join(buildDir, "lib"),
		filepath.Join(buildDir, "googletest"),
		filepath.Join(buildDir, "googlemock", "gtest"),
	}

	for _, libPath := range libPaths {
		gtestLib := filepath.Join(libPath, "libgtest.a")
		gtestMainLib := filepath.Join(libPath, "libgtest_main.a")

		if _, err := os.Stat(gtestLib); err == nil {
			if _, err := os.Stat(gtestMainLib); err == nil {
				return gtestLib, gtestMainLib, nil
			}
		}
	}

	return "", "", fmt.Errorf("Google Test libraries not found")
}

// ListCppTestFiles finds all C++ test files in the given directory
func ListCppTestFiles(dir string) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filename := strings.ToLower(info.Name())
			if strings.HasSuffix(filename, "_test.cpp") ||
				strings.HasSuffix(filename, "test.cpp") ||
				strings.HasSuffix(filename, "_test.cc") ||
				strings.HasSuffix(filename, "test.cc") {
				testFiles = append(testFiles, path)
			}
		}

		return nil
	})

	return testFiles, err
}

// SelectTestFile displays test files and allows user to select one
func SelectTestFile(testFiles []string) (string, error) {
	if len(testFiles) == 0 {
		return "", fmt.Errorf("no C++ test files found")
	}

	fmt.Println("\n📋 Available C++ test files:")
	for i, file := range testFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Print("\nSelect a test file (enter number): ")
	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}

	choice := strings.TrimSpace(scanner.Text())
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(testFiles) {
		return "", fmt.Errorf("invalid selection")
	}

	return testFiles[index-1], nil
}

// CompileAndRunCppTest compiles and runs a C++ test file with Google Test
func CompileAndRunCppTest(testFile string) error {
	fmt.Printf("🔨 Compiling %s...\n", testFile)

	// Get absolute path to ensure file exists
	absTestFile, err := filepath.Abs(testFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absTestFile); os.IsNotExist(err) {
		return fmt.Errorf("test file does not exist: %s", absTestFile)
	}

	// Extract filename without extension for executable name
	baseFile := strings.TrimSuffix(filepath.Base(testFile), filepath.Ext(testFile))
	executableName := baseFile + "_executable"

	// Get the directory where we'll run the compilation
	testDir := filepath.Dir(absTestFile)

	// Get the project root directory
	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	// Define Google Test paths
	gtestInclude := filepath.Join(projectRoot, "external", "googletest", "googletest", "include")
	gmockInclude := filepath.Join(projectRoot, "external", "googletest", "googlemock", "include")

	// Find Google Test libraries
	gtestLib, gtestMainLib, err := FindGoogleTestLibraries()
	if err != nil {
		return fmt.Errorf("failed to find Google Test libraries: %v", err)
	}

	// Check if Google Test directories exist
	if _, err := os.Stat(gtestInclude); os.IsNotExist(err) {
		return fmt.Errorf("Google Test include directory not found: %s", gtestInclude)
	}

	// Create the compile command with Google Test
	compileArgs := []string{
		"-std=c++17",
		"-I" + gtestInclude,
		"-I" + gmockInclude,
		"-pthread",
		"-o", executableName,
		absTestFile,
		gtestLib,
		gtestMainLib,
	}

	compileCmd := exec.Command("g++", compileArgs...)
	compileCmd.Dir = testDir

	fmt.Printf("🔧 Running: g++ %s\n", strings.Join(compileArgs, " "))
	fmt.Printf("📁 Working directory: %s\n", testDir)

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Compilation failed:\n%s\n", string(compileOutput))
		return fmt.Errorf("compilation failed: %v", err)
	}

	fmt.Println("✅ Compilation successful!")

	// Check if executable was created
	executablePath := filepath.Join(testDir, executableName)
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("executable was not created: %s", executablePath)
	}

	// Run the compiled test
	fmt.Printf("🚀 Running tests from %s...\n", testFile)

	runCmd := exec.Command("./" + executableName)
	runCmd.Dir = testDir

	runOutput, err := runCmd.CombinedOutput()
	fmt.Printf("📊 Test output:\n%s\n", string(runOutput))

	// Clean up executable
	cleanupCmd := exec.Command("rm", executableName)
	cleanupCmd.Dir = testDir
	cleanupCmd.Run() // Ignore cleanup errors

	if err != nil {
		return fmt.Errorf("test execution failed: %v", err)
	}

	fmt.Println("✅ Tests completed!")
	return nil
}

// RunCppTestWorkflow orchestrates the entire test running process
func RunCppTestWorkflow(testsDir string) error {
	// First, ensure Google Test is built
	if err := CheckAndBuildGoogleTest(); err != nil {
		return fmt.Errorf("failed to setup Google Test: %v", err)
	}

	// List all C++ test files in the tests directory
	testFiles, err := ListCppTestFiles(testsDir)
	if err != nil {
		return fmt.Errorf("failed to list test files: %v", err)
	}

	// Let user select a test file
	selectedFile, err := SelectTestFile(testFiles)
	if err != nil {
		return fmt.Errorf("failed to select test file: %v", err)
	}

	// Compile and run the selected test
	return CompileAndRunCppTest(selectedFile)
}
