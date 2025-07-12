package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ReadCodebase reads all C++ files from the specified directory, but only from folders listed in toScan
func ReadCodebase(dir string, toScan []string) (map[string]string, error) {
	filesContent := make(map[string]string)
	log.Printf("Reading codebase directory: %s", dir)
	log.Printf("Scanning only folders: %v", toScan)

	// Convert to absolute path for consistent handling
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %v", dir, err)
	}
	fmt.Println("Absolute directory path:", absDir)

	// Create a map for quick lookup of folders to scan
	foldersToScan := make(map[string]bool)
	for _, folder := range toScan {
		foldersToScan[folder] = true
	}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return err
		}

		// Skip if it's a directory
		if info.IsDir() {
			return nil
		}

		// Only process C++ files
		if !isCppFile(info.Name()) {
			return nil
		}

		// Get the relative path from the base directory
		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			log.Printf("Error getting relative path for %s: %v", path, err)
			return err
		}

		// Check if the file is in one of the folders we want to scan
		pathParts := strings.Split(relPath, string(filepath.Separator))
		if len(pathParts) == 0 {
			return nil
		}

		// Determine the folder to check
		var folderToCheck string
		if len(pathParts) == 1 {
			// File is directly in the root directory
			folderToCheck = "."
		} else {
			// File is in a subdirectory, check the first directory
			folderToCheck = pathParts[0]
		}

		if !foldersToScan[folderToCheck] {
			// Skip this file as it's not in a folder we want to scan
			return nil
		}

		// Store with relative path from the base directory
		relativePath := filepath.Join(dir, relPath)
		log.Printf("Found file: %s", relativePath)

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading file %s: %v", path, err)
			return err
		}

		filesContent[relativePath] = string(content)
		log.Printf("Successfully read file %s (%d bytes)", relativePath, len(content))
		return nil
	})

	if err != nil {
		log.Printf("Failed to walk codebase directory %s: %v", dir, err)
	} else {
		log.Printf("Found %d files in codebase", len(filesContent))
	}

	return filesContent, err
}

// isCppFile checks if a file is a C++ source or header file
func isCppFile(filename string) bool {
	return strings.HasSuffix(filename, ".cpp") || strings.HasSuffix(filename, ".h")
}
