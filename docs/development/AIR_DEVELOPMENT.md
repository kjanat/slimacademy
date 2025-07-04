# Air Live Reload Development Guide

This guide explains how to use Air for live reloading during SlimAcademy development.

## What is Air?

Air is a live reload utility for Go applications. When you change your code, Air automatically rebuilds and restarts your application, making development faster and more efficient.

## Installation

Air is automatically installed when you run the development setup:

```bash
./scripts/dev-setup.sh
```

Or install manually:

```bash
go install github.com/air-verse/air@latest
```

## Usage

### Basic Usage

Start Air in the project root:

```bash
air
```

This will:
1. Watch for changes in Go files
2. Automatically rebuild the `slim` binary when changes are detected
3. Run `slim list source/` to show available books
4. Display build errors if compilation fails

### Custom Configuration

The project includes a pre-configured `.air.toml` file optimized for SlimAcademy:

```bash
air -c .air.toml
```

### Configuration Details

The `.air.toml` file is configured to:

- **Build Command**: `go build -o ./bin/slim ./cmd/slim`
- **Binary Location**: `./bin/slim`
- **Default Args**: `list source/` (lists available books)
- **Watched Directories**: `cmd/`, `internal/`, `pkg/`
- **Excluded Directories**: `bin/`, `source/`, `outputs/`, `.git/`, `tmp/`
- **File Extensions**: `.go`, `.tpl`, `.tmpl`, `.html`, `.json`

### Development Workflow

1. **Start Air**:
   ```bash
   air
   ```

2. **Make Changes**: Edit any Go file in the project

3. **Automatic Rebuild**: Air detects changes and rebuilds

4. **See Results**: The CLI automatically runs with the configured command

### Customizing Air Behavior

#### Run Different Commands

Edit `.air.toml` to change the default command:

```toml
[build]
  # Change this to test different commands
  args_bin = ["convert", "--format", "html", "source/Homeostase Colleges", "--output", "/tmp/test.html"]
```

#### Debug Mode

Run Air with debug output:

```bash
air -d
```

#### Pass Arguments to Your Binary

```bash
# This runs: ./bin/slim check source/Homeostase Colleges
air check source/Homeostase Colleges

# Use -- to pass flags
air -- --help
```

### Common Use Cases

#### Testing Document Conversion

```toml
# .air.toml
[build]
  args_bin = ["convert", "--format", "html", "source/Homeostase Colleges", "--output", "/tmp/test.html"]
```

#### Running Tests on Change

```toml
# .air.toml
[build]
  cmd = "go test ./... && go build -o ./bin/slim ./cmd/slim"
```

#### Checking All Books

```toml
# .air.toml
[build]
  args_bin = ["check", "--all"]
```

### Troubleshooting

#### Air Command Not Found

If `air` is not found after installation:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Build Errors Not Showing

Check the `build-errors.log` file:

```bash
tail -f build-errors.log
```

#### Too Many Open Files

On macOS, you might need to increase the file descriptor limit:

```bash
ulimit -n 2048
```

### Tips and Tricks

1. **Clear Screen on Rebuild**: The config is set to clear the terminal on each rebuild for cleaner output

2. **Timestamps**: Logs include timestamps to track when rebuilds occur

3. **Custom Environment**: Create a `.env` file for development-specific variables

4. **Multiple Configs**: Create different Air configs for different scenarios:
   ```bash
   air -c .air.test.toml  # Run tests on change
   air -c .air.bench.toml # Run benchmarks on change
   ```

5. **Docker Development**: Use Air with Docker:
   ```bash
   docker run -it --rm \
     -w /app \
     -v $(pwd):/app \
     -p 8080:8080 \
     cosmtrek/air
   ```

## Integration with VS Code

The project includes VS Code tasks that work well with Air:

1. Open VS Code: `code slimacademy.code-workspace`
2. Terminal → Run Task → Choose "Start Air"
3. Edit files and see live updates in the terminal

## Best Practices

1. **Keep Build Fast**: Exclude unnecessary directories from watching
2. **Use Appropriate Commands**: Set `args_bin` to commands that provide quick feedback
3. **Monitor Resources**: Air uses inotify on Linux; ensure adequate limits
4. **Clean Regularly**: Run `go clean -cache` if builds become slow

## Next Steps

- Customize `.air.toml` for your workflow
- Create task-specific Air configurations
- Integrate with your preferred editor/IDE
- Share useful configurations with the team
