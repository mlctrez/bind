//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Clean removes the temp directory
func Clean() error {
	fmt.Println("Cleaning temp directory...")
	return os.RemoveAll("temp")
}

// Test runs unit tests with code coverage and generates an HTML report
func Test() error {
	mg.Deps(Clean)

	// Create a temp directory for coverage files
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	coverageFile := filepath.Join(tempDir, "coverage.out")
	coverageHTML := filepath.Join(tempDir, "coverage.html")

	fmt.Println("Running tests with coverage...")

	// Run tests with coverage
	if err := sh.Run("go", "test", "-v", "-coverprofile="+coverageFile, "./..."); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Generate HTML coverage report
	fmt.Printf("Generating HTML coverage report at %s...\n", coverageHTML)
	if err := sh.Run("go", "tool", "cover", "-html="+coverageFile, "-o", coverageHTML); err != nil {
		return fmt.Errorf("failed to generate HTML coverage report: %w", err)
	}

	// Show coverage summary
	fmt.Println("Coverage summary:")
	if err := sh.Run("go", "tool", "cover", "-func="+coverageFile); err != nil {
		return err
	}

	fmt.Printf("\nCoverage files generated:\n")
	fmt.Printf("- Profile: %s\n", coverageFile)
	fmt.Printf("- HTML Report: %s\n", coverageHTML)

	// Try to open coverage report in browser
	fmt.Println("\nOpening coverage report...")
	if err := openBrowser(coverageHTML); err != nil {
		fmt.Printf("Could not auto-open browser. Open manually: %s\n", coverageHTML)
	}

	return nil
}

// Doc starts a local godoc server for this project and opens it in the browser
func Doc() error {
	fmt.Println("Starting godoc server for this project...")
	fmt.Println("Server will be available at: http://localhost:6060/pkg/github.com/mlctrez/bind/")
	fmt.Println("Press Ctrl+C to stop the server")

	// Try to open the documentation URL in browser after a brief delay
	go func() {
		// Give godoc a moment to start up
		sh.Run("sleep", "2")
		docURL := "http://localhost:6060/pkg/github.com/mlctrez/bind/"
		fmt.Printf("Opening documentation at: %s\n", docURL)
		if err := sh.Run("xdg-open", docURL); err != nil {
			fmt.Printf("Could not auto-open browser. Open manually: %s\n", docURL)
		}
	}()

	// Start godoc server on port 6060
	return sh.Run("godoc", "-http=:6060")
}

// openBrowser opens the coverage HTML file in the default browser using xdg-open
func openBrowser(file string) error {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	return sh.Run("xdg-open", absPath)
}
