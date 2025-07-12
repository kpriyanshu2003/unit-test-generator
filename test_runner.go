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

// ListSourceFiles finds all C++ source files in the given directory
func ListSourceFiles(dir string) ([]string, error) {
	var sourceFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filename := strings.ToLower(info.Name())
			if strings.HasSuffix(filename, ".cpp") ||
				strings.HasSuffix(filename, ".cc") ||
				strings.HasSuffix(filename, ".c") {
				// Exclude test files from source files
				if !strings.Contains(filename, "test") {
					sourceFiles = append(sourceFiles, path)
				}
			}
		}

		return nil
	})

	return sourceFiles, err
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

// GenerateCoverageSummary captures coverage and produces a command-line summary report.
func GenerateCoverageSummary(testDir string, sourceDir string) error {
	fmt.Println("📊 Generating coverage summary...")

	// --- Step 1: Capture coverage data using a robust lcov command ---
	rawInfoFile := filepath.Join(testDir, "coverage.raw.info")
	projectRoot, _ := filepath.Abs(".")

	// Define patterns to exclude from the very beginning.
	excludePatterns := []string{
		filepath.Join(projectRoot, "external", "*"),  // Exclude Google Test
		filepath.Join(projectRoot, "tests-new", "*"), // Exclude the test files themselves
		"/usr/include/*",        // Exclude system headers (Linux)
		"/Applications/*",       // Exclude Xcode/macOS system headers
		"*/Library/Developer/*", // Exclude macOS developer tools headers
	}

	// Build the lcov command arguments
	lcovArgs := []string{
		"--capture",
		"--directory", testDir,
		"--output-file", rawInfoFile,
		"--ignore-errors", "unsupported,inconsistent,unused",
	}
	for _, p := range excludePatterns {
		lcovArgs = append(lcovArgs, "--exclude", p)
	}

	captureCmd := exec.Command("lcov", lcovArgs...)
	if output, err := captureCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lcov capture failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("   [1/2] Raw coverage data collected and filtered.")

	// --- Step 2: Manually parse the raw info file to calculate coverage ---
	file, err := os.Open(rawInfoFile)
	if err != nil {
		fmt.Println("⚠️  No coverage data was generated for the source files. This may be because they were fully excluded or the source directory is incorrect.")
		return nil
	}
	defer file.Close()

	totalLines := 0
	coveredLines := 0
	var currentFile string
	isSourceFile := false

	absSourceDir, _ := filepath.Abs(sourceDir)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "SF:") {
			currentFile = strings.TrimPrefix(line, "SF:")
			isSourceFile = strings.HasPrefix(currentFile, absSourceDir)
		}
		if isSourceFile && strings.HasPrefix(line, "DA:") {
			parts := strings.Split(strings.TrimPrefix(line, "DA:"), ",")
			if len(parts) == 2 {
				totalLines++
				hitCount, err := strconv.Atoi(parts[1])
				if err == nil && hitCount > 0 {
					coveredLines++
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading coverage file: %v", err)
	}

	// Clean up the temporary raw info file immediately after parsing
	os.Remove(rawInfoFile)

	fmt.Println("   [2/2] Coverage data parsed.")

	// --- Step 3: Format the summary and save it to a file ---
	var summaryContent string
	if totalLines == 0 {
		summaryContent = `
---------------------
Code Coverage Summary
---------------------
⚠️  No executable lines were found for the source files.
   Please check if the 'source_directory' argument is correct.
---------------------
`
	} else {
		var coveragePercentage float64 = (float64(coveredLines) / float64(totalLines)) * 100
		summaryContent = fmt.Sprintf(`
---------------------
Code Coverage Summary
---------------------
Total lines:    %d
Covered lines:  %d
Coverage:       %.2f%%
Uncovered lines: %d
---------------------
`, totalLines, coveredLines, coveragePercentage, totalLines-coveredLines)
	}

	// Print the summary to the console
	fmt.Print(summaryContent)

	// Define the path for the output file
	coverageDir := filepath.Join(testDir, "coverage")
	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		return fmt.Errorf("could not create coverage directory: %v", err)
	}
	summaryFilePath := filepath.Join(coverageDir, "coverage_summary.txt")

	// Write the summary to the file
	if err := os.WriteFile(summaryFilePath, []byte(strings.TrimSpace(summaryContent)), 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %v", err)
	}

	fmt.Printf("\n✅ Summary saved to: %s\n", summaryFilePath)

	return nil
}

// CleanupTestDirectory removes all intermediate files generated during compilation and testing.
func CleanupTestDirectory(testDir string, executableName string) {
	fmt.Println("🧹 Cleaning up intermediate files...")

	// Patterns for files to remove
	patterns := []string{
		filepath.Join(testDir, "*.gcno"),
		filepath.Join(testDir, "*.gcda"),
		filepath.Join(testDir, executableName),
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err == nil {
			for _, file := range files {
				os.Remove(file)
			}
		}
	}

	// Also clean up .dSYM directories on macOS
	dsymPattern := filepath.Join(testDir, "*.dSYM")
	dsymFiles, err := filepath.Glob(dsymPattern)
	if err == nil {
		for _, dsymFile := range dsymFiles {
			os.RemoveAll(dsymFile)
		}
	}
}

// CompileAndRunCppTest compiles and runs a C++ test, then generates a coverage report.
func CompileAndRunCppTest(testFile string, sourceDir string) error {
	fmt.Printf("🔨 Compiling %s with coverage...\n", testFile)

	absTestFile, err := filepath.Abs(testFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for test file: %v", err)
	}
	if _, err := os.Stat(absTestFile); os.IsNotExist(err) {
		return fmt.Errorf("test file does not exist: %s", absTestFile)
	}

	baseFile := strings.TrimSuffix(filepath.Base(testFile), filepath.Ext(testFile))
	executableName := baseFile + "_executable"
	testDir := filepath.Dir(absTestFile)

	// Clean up from any previous runs before we start
	CleanupTestDirectory(testDir, executableName)

	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	// Google Test paths
	gtestInclude := filepath.Join(projectRoot, "external", "googletest", "googletest", "include")
	gmockInclude := filepath.Join(projectRoot, "external", "googletest", "googlemock", "include")
	gtestLib, gtestMainLib, err := FindGoogleTestLibraries()
	if err != nil {
		return fmt.Errorf("failed to find Google Test libraries: %v", err)
	}

	// Source files
	sourceFiles, err := ListSourceFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to list source files: %v", err)
	}
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source directory: %v", err)
	}

	// --- Compile Command ---
	compileArgs := []string{
		"-std=c++17",
		"-g",
		"-O0",        // No optimization for accurate line numbers
		"--coverage", // This flag combines -fprofile-arcs and -ftest-coverage
		"-I" + gtestInclude,
		"-I" + gmockInclude,
		"-I" + absSourceDir,
		"-pthread",
		"-o", executableName,
		absTestFile,
	}
	// Add all source files to compilation
	for _, sourceFile := range sourceFiles {
		absSourceFile, err := filepath.Abs(sourceFile)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not get absolute path for %s: %v\n", sourceFile, err)
			continue
		}
		compileArgs = append(compileArgs, absSourceFile)
	}
	compileArgs = append(compileArgs, gtestLib, gtestMainLib)

	compileCmd := exec.Command("g++", compileArgs...)
	compileCmd.Dir = testDir // Run compilation in the test directory

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Compilation failed:\n%s\n", string(compileOutput))
		return fmt.Errorf("compilation failed: %v", err)
	}
	fmt.Println("✅ Compilation successful!")

	// --- Run Test Executable ---
	fmt.Printf("🚀 Running tests from %s...\n", testFile)
	executablePath := filepath.Join(testDir, executableName)
	runCmd := exec.Command(executablePath)
	runCmd.Dir = testDir

	runOutput, runErr := runCmd.CombinedOutput()
	fmt.Printf("📊 Test output:\n%s\n", string(runOutput))

	// --- Generate Report ---
	// Only generate report if tests ran (even if they failed)
	if coverageErr := GenerateCoverageSummary(testDir, sourceDir); coverageErr != nil {
		fmt.Printf("⚠️  Coverage summary generation failed: %v\n", coverageErr)
	}

	// --- Final Cleanup ---
	CleanupTestDirectory(testDir, executableName)

	if runErr != nil {
		return fmt.Errorf("test execution failed: %v", runErr)
	}

	fmt.Println("✅ Tests and coverage generation completed!")
	return nil
}

// RunCppTestWorkflow orchestrates the entire test running process with coverage
func RunCppTestWorkflow(testsDir string, sourceDir string) error {
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

	// Compile and run the selected test with source files and coverage
	return CompileAndRunCppTest(selectedFile, sourceDir)
}
