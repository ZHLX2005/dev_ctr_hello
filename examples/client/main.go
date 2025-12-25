// Package main provides file server client example
package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	// Server URL
	serverURL = "http://localhost:8080"
	// Auth token - change this to your actual token
	authToken = "your-secret-token-change-me"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go upload <file_path> [TTL]")
			return
		}
		filePath := os.Args[2]
		ttl := "1h"
		if len(os.Args) > 3 {
			ttl = os.Args[3]
		}
		uploadFile(filePath, ttl)

	case "download":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go download <file_id> [output_path]")
			return
		}
		fileID := os.Args[2]
		outputPath := ""
		if len(os.Args) > 3 {
			outputPath = os.Args[3]
		}
		downloadFile(fileID, outputPath)

	case "metadata":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go metadata <file_id>")
			return
		}
		fileID := os.Args[2]
		getMetadata(fileID)

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go delete <file_id>")
			return
		}
		fileID := os.Args[2]
		deleteFile(fileID)

	case "health":
		checkHealth()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("File Server Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go upload <file_path> [TTL]      - Upload file")
	fmt.Println("  go run main.go download <file_id> [output]  - Download file")
	fmt.Println("  go run main.go metadata <file_id>           - Get file metadata")
	fmt.Println("  go run main.go delete <file_id>             - Delete file")
	fmt.Println("  go run main.go health                        - Health check")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go upload document.pdf 2h")
	fmt.Println("  go run main.go download a1b2c3d4e5f6...")
	fmt.Println("  go run main.go metadata a1b2c3d4e5f6...")
	fmt.Println()
	fmt.Println("Note: Update 'authToken' constant with your actual token")
}

// uploadFile Upload a file
func uploadFile(filePath, ttl string) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, _ := file.Stat()
	fileName := fileInfo.Name()

	// Create request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		return
	}
	io.Copy(part, file)

	// Add TTL
	writer.WriteField("ttl", ttl)
	writer.Close()

	// Create request
	url := serverURL + "/api/v1/upload"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Upload failed (status: %d): %s\n", resp.StatusCode, string(respBody))
		return
	}

	fmt.Println("Upload successful!")
	fmt.Println(string(respBody))
}

// downloadFile Download a file
func downloadFile(fileID, outputPath string) {
	url := fmt.Sprintf("%s/api/v1/download/%s", serverURL, fileID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Download request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("File not found")
		return
	}

	if resp.StatusCode == http.StatusGone {
		fmt.Println("File has expired")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Download failed (status: %d)\n", resp.StatusCode)
		return
	}

	// Get filename from response headers
	fileName := resp.Header.Get("X-File-Name")
	if fileName == "" {
		fileName = fileID
	}

	// Determine output path
	if outputPath == "" {
		outputPath = fileName
	}

	// Create file
	out, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer out.Close()

	// Write file
	io.Copy(out, resp.Body)

	fmt.Printf("Download successful: %s\n", outputPath)
}

// getMetadata Get file metadata
func getMetadata(fileID string) {
	url := fmt.Sprintf("%s/api/v1/file/%s/metadata", serverURL, fileID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("File not found")
		return
	}

	if resp.StatusCode == http.StatusGone {
		fmt.Println("File has expired")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed (status: %d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Println("File metadata:")
	fmt.Println(string(body))
}

// deleteFile Delete a file
func deleteFile(fileID string) {
	url := fmt.Sprintf("%s/api/v1/file/%s", serverURL, fileID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Delete failed (status: %d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Println("Delete successful!")
}

// checkHealth Health check
func checkHealth() {
	url := serverURL + "/health"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Service unhealthy (status: %d)\n", resp.StatusCode)
		return
	}

	fmt.Println("Service healthy:")
	fmt.Println(string(body))
}
