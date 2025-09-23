package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {

	// Create MCP server
	s := server.NewMCPServer(
		"mcp-files-server",
		"0.0.0",
	)

	// Read file tool
	readFileTool := mcp.NewTool("read_file",
		mcp.WithDescription("Read the content of a text file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to read"),
		),
	)
	s.AddTool(readFileTool, readFileHandler)

	// Write file tool
	writeFileTool := mcp.NewTool("write_file",
		mcp.WithDescription("Write content to a text file"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to write"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to write to the file"),
		),
	)
	s.AddTool(writeFileTool, writeFileHandler)

	// generateFileTool := mcp.NewTool("generate_file",
	// 	mcp.WithDescription("Write content to a text file"),
	// 	mcp.WithString("file_path",
	// 		mcp.Required(),
	// 		mcp.Description("Path to the file to write"),
	// 	),
	// 	mcp.WithString("content",
	// 		mcp.Required(),
	// 		mcp.Description("Content to write to the file"),
	// 	),
	// )
	// s.AddTool(generateFileTool, writeFileHandler)	

	// Delete file tool
	deleteFileTool := mcp.NewTool("delete_file",
		mcp.WithDescription("Delete a file from the filesystem"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to delete"),
		),
	)
	s.AddTool(deleteFileTool, deleteFileHandler)

	// Create directory tool
	createDirectoryTool := mcp.NewTool("create_directory",
		mcp.WithDescription("Create a directory and its parent directories if they don't exist"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to create"),
		),
	)
	s.AddTool(createDirectoryTool, createDirectoryHandler)

	// Delete directory tool
	deleteDirectoryTool := mcp.NewTool("delete_directory",
		mcp.WithDescription("Delete a directory and all its contents from the filesystem"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to delete"),
		),
	)
	s.AddTool(deleteDirectoryTool, deleteDirectoryHandler)

	// List directory tool
	listDirectoryTool := mcp.NewTool("list_directory",
		mcp.WithDescription("List the contents of a directory"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to list"),
		),
	)
	s.AddTool(listDirectoryTool, listDirectoryHandler)

	// Tree view tool
	treeViewTool := mcp.NewTool("tree_view",
		mcp.WithDescription("Display a tree view of a directory structure"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to the directory to display as tree"),
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Maximum depth to traverse (default: unlimited)"),
		),
	)
	s.AddTool(treeViewTool, treeViewHandler)

	// Start the HTTP server
	httpPort := os.Getenv("MCP_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9090"
	}

	log.Println("MCP Files Server is running on port", httpPort)

	// Create a custom mux to handle both MCP and health endpoints
	mux := http.NewServeMux()

	// Add healthcheck endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Add MCP endpoint
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/mcp"),
	)

	// Register MCP handler with the mux
	mux.Handle("/mcp", httpServer)

	// Start the HTTP server with custom mux
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

// validatePath ensures that the file path is within the workspace folder and prevents path traversal attacks
func validatePath(userPath string) (string, error) {
	workspaceFolder := os.Getenv("LOCAL_WORKSPACE_FOLDER")
	if workspaceFolder == "" {
		return "", fmt.Errorf("LOCAL_WORKSPACE_FOLDER environment variable is not set")
	}

	// Get absolute path of workspace folder
	absWorkspace, err := filepath.Abs(workspaceFolder)
	if err != nil {
		return "", fmt.Errorf("error resolving workspace path: %v", err)
	}

	// Clean the user provided path
	cleanUserPath := filepath.Clean(userPath)

	// Remove any leading slashes to ensure it's treated as relative
	cleanUserPath = strings.TrimPrefix(cleanUserPath, "/")
	cleanUserPath = strings.TrimPrefix(cleanUserPath, "\\")

	// Join with workspace folder
	fullPath := filepath.Join(absWorkspace, cleanUserPath)

	// Get absolute path to resolve any remaining .. or . components
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("error resolving file path: %v", err)
	}

	// Check if the resolved path is still within the workspace
	if !strings.HasPrefix(absFullPath, absWorkspace) {
		return "", fmt.Errorf("access denied: path is outside workspace folder")
	}

	return absFullPath, nil
}

func createDirectoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	directoryPathArg, exists := args["directory_path"]
	if !exists || directoryPathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'directory_path'")
	}

	directoryPath, ok := directoryPathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'directory_path' must be a string")
	}

	// Validate and secure the directory path
	cleanPath, err := validatePath(directoryPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory path: %v", err)), nil
	}

	// Create the directory and all parent directories
	err = os.MkdirAll(cleanPath, 0755)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating directory: %v", err)), nil
	}

	log.Printf("Successfully created directory: %s", cleanPath)
	return mcp.NewToolResultText(fmt.Sprintf("Successfully created directory: %s", cleanPath)), nil
}

func deleteDirectoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	directoryPathArg, exists := args["directory_path"]
	if !exists || directoryPathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'directory_path'")
	}

	directoryPath, ok := directoryPathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'directory_path' must be a string")
	}

	// Validate and secure the directory path
	cleanPath, err := validatePath(directoryPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory path: %v", err)), nil
	}

	// Check if directory exists
	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory not found: %s", cleanPath)), nil
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing directory: %v", err)), nil
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("Path is not a directory: %s", cleanPath)), nil
	}

	// Delete the directory and all its contents
	err = os.RemoveAll(cleanPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error deleting directory: %v", err)), nil
	}

	log.Printf("Successfully deleted directory: %s", cleanPath)
	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted directory: %s", cleanPath)), nil
}

func listDirectoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	directoryPathArg, exists := args["directory_path"]
	if !exists || directoryPathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'directory_path'")
	}

	directoryPath, ok := directoryPathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'directory_path' must be a string")
	}

	// Validate and secure the directory path
	cleanPath, err := validatePath(directoryPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory path: %v", err)), nil
	}

	// Check if directory exists
	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory not found: %s", cleanPath)), nil
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing directory: %v", err)), nil
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("Path is not a directory: %s", cleanPath)), nil
	}

	// Read directory contents
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading directory: %v", err)), nil
	}

	// Build the result
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of directory: %s\n\n", directoryPath))

	if len(entries) == 0 {
		result.WriteString("(empty directory)")
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				result.WriteString(fmt.Sprintf("%s/\n", entry.Name()))
			} else {
				info, err := entry.Info()
				if err == nil {
					result.WriteString(fmt.Sprintf("%s (%d bytes)\n", entry.Name(), info.Size()))
				} else {
					result.WriteString(fmt.Sprintf("%s\n", entry.Name()))
				}
			}
		}
	}

	log.Printf("Successfully listed directory: %s (%d entries)", cleanPath, len(entries))
	return mcp.NewToolResultText(result.String()), nil
}

func treeViewHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	directoryPathArg, exists := args["directory_path"]
	if !exists || directoryPathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'directory_path'")
	}

	directoryPath, ok := directoryPathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'directory_path' must be a string")
	}

	// Parse max_depth parameter
	maxDepth := -1 // unlimited by default
	if maxDepthArg, exists := args["max_depth"]; exists && maxDepthArg != nil {
		if depth, ok := maxDepthArg.(float64); ok {
			maxDepth = int(depth)
		}
	}

	// Validate and secure the directory path
	cleanPath, err := validatePath(directoryPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory path: %v", err)), nil
	}

	// Check if directory exists
	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory not found: %s", cleanPath)), nil
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing directory: %v", err)), nil
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("Path is not a directory: %s", cleanPath)), nil
	}

	// Build tree view
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Tree view of directory: %s\n\n", directoryPath))

	err = buildTreeView(&result, cleanPath, "", 0, maxDepth)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error building tree view: %v", err)), nil
	}

	log.Printf("Successfully generated tree view for directory: %s", cleanPath)
	return mcp.NewToolResultText(result.String()), nil
}

func buildTreeView(result *strings.Builder, path string, prefix string, currentDepth int, maxDepth int) error {
	if maxDepth >= 0 && currentDepth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1

		// Determine the tree characters
		var connector, newPrefix string
		if isLast {
			connector = "└── "
			newPrefix = prefix + "    "
		} else {
			connector = "├── "
			newPrefix = prefix + "│   "
		}

		// Write current entry
		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("%s%s%s/\n", prefix, connector, entry.Name()))
			// Recursively process subdirectory
			subPath := filepath.Join(path, entry.Name())
			err := buildTreeView(result, subPath, newPrefix, currentDepth+1, maxDepth)
			if err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err == nil {
				result.WriteString(fmt.Sprintf("%s%s%s (%d bytes)\n", prefix, connector, entry.Name(), info.Size()))
			} else {
				result.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, entry.Name()))
			}
		}
	}

	return nil
}

func readFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filePathArg, exists := args["file_path"]
	if !exists || filePathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'file_path'")
	}

	filePath, ok := filePathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'file_path' must be a string")
	}

	// Validate and secure the file path
	cleanPath, err := validatePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file path: %v", err)), nil
	}

	// Read the file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", cleanPath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}

	log.Printf("Successfully read file: %s (%d bytes)", cleanPath, len(content))
	return mcp.NewToolResultText(string(content)), nil
}

func writeFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filePathArg, exists := args["file_path"]
	if !exists || filePathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'file_path'")
	}

	filePath, ok := filePathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'file_path' must be a string")
	}

	contentArg, exists := args["content"]
	if !exists || contentArg == nil {
		return nil, fmt.Errorf("missing required parameter 'content'")
	}

	content, ok := contentArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'content' must be a string")
	}

	// Validate and secure the file path
	cleanPath, err := validatePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file path: %v", err)), nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating directory: %v", err)), nil
	}

	// Write the file
	err = os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing file: %v", err)), nil
	}

	log.Printf("Successfully wrote file: %s (%d bytes)", cleanPath, len(content))
	return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), cleanPath)), nil
}

func deleteFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filePathArg, exists := args["file_path"]
	if !exists || filePathArg == nil {
		return nil, fmt.Errorf("missing required parameter 'file_path'")
	}

	filePath, ok := filePathArg.(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'file_path' must be a string")
	}

	// Validate and secure the file path
	cleanPath, err := validatePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid file path: %v", err)), nil
	}

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", cleanPath)), nil
	}

	// Delete the file
	err = os.Remove(cleanPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error deleting file: %v", err)), nil
	}

	log.Printf("Successfully deleted file: %s", cleanPath)
	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted file: %s", cleanPath)), nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"server": "mcp-files-server",
	}
	json.NewEncoder(w).Encode(response)
}
