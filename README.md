# MCP Files Server

A lightweight HTTP streamable MCP (Model Context Protocol) server implemented in Go that provides secure file operations within a designated workspace.

## Goal

The MCP Files Server enables AI assistants and automation tools to safely read and write text files through the standardized MCP protocol. This server provides:

- **Workspace Isolation**: All file operations are contained within a configurable workspace directory via `LOCAL_WORKSPACE_FOLDER`
- **Path Traversal Protection**: Advanced security validation prevents access outside the workspace folder (blocks `../`, absolute paths, etc.)
- **Secure File Access**: Proper path validation and cleaning to prevent directory traversal attacks
- **Simple API**: Eight essential tools for file manipulation - read, write, delete files, create/delete directories, list directory contents, and tree view operations
- **HTTP Streaming**: Built on the MCP streamable HTTP protocol for real-time communication
- **Zero Dependencies**: Minimal external dependencies for easy deployment and maintenance

## Use Cases

- **Content Management**: Allow AI assistants to read configuration files, documentation, or data files
- **Code Generation**: Enable automated code writing and modification within project boundaries
- **Data Processing**: Read input files and write processed results
- **Template Systems**: Read templates and write generated content
- **Log Analysis**: Read log files and write analysis reports
- **Backup Operations**: Read source files and write backup copies

## Tools

### `read_file`
Reads the content of a text file within the workspace.

**Parameters:**
- `file_path` (string, required): Relative path to the file within the workspace (automatically validated for security)

**Returns:** The complete file content as text

**Security:** Path is automatically validated to prevent access outside the workspace folder.

### `write_file`
Writes content to a text file within the workspace.

**Parameters:**
- `file_path` (string, required): Relative path to the file within the workspace (automatically validated for security)
- `content` (string, required): Text content to write to the file

**Returns:** Success message with file path and byte count

**Security:** Path is automatically validated to prevent access outside the workspace folder. Directories are created automatically if needed.

### `delete_file`
Deletes a file from the filesystem within the workspace.

**Parameters:**
- `file_path` (string, required): Relative path to the file within the workspace (automatically validated for security)

**Returns:** Success message confirming file deletion

**Security:** Path is automatically validated to prevent access outside the workspace folder.

### `create_directory`
Creates a directory and all necessary parent directories within the workspace.

**Parameters:**
- `directory_path` (string, required): Relative path to the directory to create within the workspace (automatically validated for security)

**Returns:** Success message confirming directory creation

**Security:** Path is automatically validated to prevent access outside the workspace folder. Parent directories are created automatically if needed.

### `delete_directory`
Deletes a directory and all its contents from the filesystem within the workspace.

**Parameters:**
- `directory_path` (string, required): Relative path to the directory to delete within the workspace (automatically validated for security)

**Returns:** Success message confirming directory deletion

**Security:** Path is automatically validated to prevent access outside the workspace folder. Includes verification that the path is actually a directory before deletion.

### `list_directory`
Lists the contents of a directory within the workspace.

**Parameters:**
- `directory_path` (string, required): Relative path to the directory to list within the workspace (automatically validated for security)

**Returns:** Formatted list of directory contents with file sizes (directories marked with trailing /)

**Security:** Path is automatically validated to prevent access outside the workspace folder. Includes verification that the path is actually a directory.

### `tree_view`
Displays a tree view of a directory structure within the workspace.

**Parameters:**
- `directory_path` (string, required): Relative path to the directory to display as tree within the workspace (automatically validated for security)
- `max_depth` (number, optional): Maximum depth to traverse (default: unlimited)

**Returns:** Formatted tree view with file sizes and proper tree structure characters (directories marked with trailing /)

**Security:** Path is automatically validated to prevent access outside the workspace folder. Includes verification that the path is actually a directory. Depth limiting prevents excessive resource usage.

## Architecture

The server follows the MCP protocol specification and provides:

1. **Session Management**: Each client connection gets a unique session ID
2. **Tool Registration**: All eight file tools are registered with proper parameter validation
3. **Error Handling**: Comprehensive error responses for missing files, permissions, etc.
4. **Health Monitoring**: `/health` endpoint for service monitoring
5. **Workspace Security**: All file paths are validated and contained within `LOCAL_WORKSPACE_FOLDER`
6. **Path Validation**: Comprehensive protection against directory traversal attacks including:
   - Resolution of relative path components (`../`, `./`)
   - Removal of absolute path prefixes (`/`, `\`)
   - Verification that final resolved paths remain within the workspace
   - Prevention of symlink-based escapes


## Configuration

Configure the server using environment variables in `mcp.server.env`:

```bash
MCP_HTTP_PORT=9096                          # HTTP port for the server
LOCAL_WORKSPACE_FOLDER=/path/to/workspace   # Base directory for all file operations
```

## Docker Support

The project includes Docker configuration for easy deployment.

## Use the Docker Image

Image: https://hub.docker.com/repository/docker/k33g/mcp-files-server/tags


```yaml
services:

  mcp-files-server:
    image: k33g/mcp-files-server:0.0.1
    ports:
      - 9090:6060
    environment:
      - MCP_HTTP_PORT=6060
      - LOCAL_WORKSPACE_FOLDER=/app/workspace
    volumes:
      - ./workspace:/app/workspace
```

Start the server with:

```bash
docker compose up
```
