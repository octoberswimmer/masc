package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// global state for serve command
var (
	servePort       int
	serveDir        string
	currentBuildDir string
	buildMutex      sync.RWMutex
)

// indexHTML is the HTML template served for the Thunder app root.
const indexHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Main App</title>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("bundle.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });
    </script>
</head>
<body>
    <div id="app"></div>
</body>
</html>`

// root command
var rootCmd = &cobra.Command{Use: "masc"}

// serve command
var serveCmd = &cobra.Command{
	Use:   "serve [dir]",
	Short: "Build and serve the masc app locally",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServe,
}

func init() {
	// serve flags (port only; app dir is optional positional arg)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8000, "Port to serve on")
	// add subcommands
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildWASM compiles the Go app in appDir to WebAssembly and prepares assets.
func buildWASM(appDir string) (string, error) {
	// create temporary build directory
	buildDir, err := os.MkdirTemp("", "masc-build-*")
	if err != nil {
		return "", err
	}
	// build WASM binary
	outWasm := filepath.Join(buildDir, "bundle.wasm")
	cmd := exec.Command("go", "build", "-o", outWasm, "-tags", "dev")

	// Set up environment with smart GOWORK handling
	env := append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if shouldDisableWorkspace(appDir) {
		env = append(env, "GOWORK=off")
		fmt.Printf("Note: Disabling go.work for standalone module build\n")
	}
	cmd.Env = env

	absPath, err := filepath.Abs(appDir)
	if err != nil {
		return "", fmt.Errorf("failed to set app dir: %w", err)
	}
	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	// copy wasm_exec.js from Go SDK
	gorootCmd := exec.Command("go", "env", "GOROOT")
	gorootBytes, err := gorootCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOROOT: %w", err)
	}
	goroot := strings.TrimSpace(string(gorootBytes))
	wasmExecSrc := filepath.Join(goroot, "lib", "wasm", "wasm_exec.js")
	wasmExecDst := filepath.Join(buildDir, "wasm_exec.js")
	if err := copyFile(wasmExecSrc, wasmExecDst); err != nil {
		return "", err
	}
	return buildDir, nil
}

// findGoWork searches for go.work file starting from dir and walking up
func findGoWork(dir string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		workFile := filepath.Join(absDir, "go.work")
		if _, err := os.Stat(workFile); err == nil {
			return workFile
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			break // reached root
		}
		absDir = parent
	}
	return ""
}

// parseWorkspaceModules parses go.work file and returns the list of module directories
func parseWorkspaceModules(workFile string) ([]string, error) {
	file, err := os.Open(workFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", err)
		}
	}()

	var modules []string
	scanner := bufio.NewScanner(file)
	inUseBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "use (") {
			inUseBlock = true
			continue
		}
		if strings.HasPrefix(line, "use ") && !strings.Contains(line, "(") {
			// Single line use directive
			module := strings.TrimSpace(strings.TrimPrefix(line, "use"))
			if !strings.HasPrefix(module, "//") && module != "" {
				modules = append(modules, module)
			}
			continue
		}
		if inUseBlock {
			if strings.HasPrefix(line, ")") {
				inUseBlock = false
				continue
			}
			if strings.HasPrefix(line, "//") || line == "" {
				continue
			}
			modules = append(modules, line)
		}
	}

	// Convert relative paths to absolute paths relative to workspace file
	workDir := filepath.Dir(workFile)
	for i, module := range modules {
		if !filepath.IsAbs(module) {
			modules[i] = filepath.Join(workDir, module)
		}
	}

	return modules, scanner.Err()
}

// shouldDisableWorkspace determines if GOWORK should be disabled for the target directory
func shouldDisableWorkspace(targetDir string) bool {
	workFile := findGoWork(targetDir)
	if workFile == "" {
		return false // No workspace to disable
	}

	modules, err := parseWorkspaceModules(workFile)
	if err != nil {
		// If we can't parse the workspace, be conservative and don't disable it
		return false
	}

	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return false
	}

	// Check if target directory matches any workspace module
	for _, module := range modules {
		absModule, err := filepath.Abs(module)
		if err != nil {
			continue
		}
		if absTarget == absModule {
			return false // Target is in workspace, keep GOWORK enabled
		}
	}

	return true // Target not in workspace, disable GOWORK
}

// findFreePort finds and reserves a free port, returning both the port and listener
func findFreePort(preferredPort int) (int, net.Listener, error) {
	// First try the preferred port
	if preferredPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", preferredPort))
		if err == nil {
			return preferredPort, ln, nil
		}
		fmt.Printf("Port %d is in use, finding alternative...\n", preferredPort)
	}

	// Let OS assign a free port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, nil, err
	}

	port := ln.Addr().(*net.TCPAddr).Port
	return port, ln, nil
}

// copyFile copies a file from src to dst, creating parent directories.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := in.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close input file: %v\n", err)
		}
	}()
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close output file: %v\n", err)
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// watchFiles watches Go source files and calls the provided callback on changes.
func watchFiles(appDir string, onRebuild func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up file watcher: %w", err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close watcher: %v\n", err)
		}
	}()

	// watch app and local module directories recursively
	// Determine module directories via `go list`
	gomodcacheBytes, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting GOMODCACHE: %v\n", err)
	}
	gomodcache := strings.TrimSpace(string(gomodcacheBytes))

	listCmd := exec.Command("go", "list", "-C", appDir, "-m", "-mod=readonly", "-f", "{{.Dir}}", "all")

	// Use same environment setup as build commands for consistency
	env := os.Environ()
	if shouldDisableWorkspace(appDir) {
		env = append(env, "GOWORK=off")
	}
	listCmd.Env = env

	out, err := listCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing modules: %v\n", err)
	}
	roots := make(map[string]struct{})
	// always include the app directory
	roots[appDir] = struct{}{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// skip modules in GOMODCACHE
		if strings.HasPrefix(line, gomodcache) {
			continue
		}
		roots[line] = struct{}{}
	}
	// Walk and watch each root directory
	for root := range roots {
		err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				if watchErr := watcher.Add(path); watchErr != nil {
					fmt.Fprintf(os.Stderr, "Error watching %s: %v\n", path, watchErr)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s for file watching: %v\n", root, err)
		}
	}

	// Debounce mechanism for rebuilds
	rebuildTimer := time.NewTimer(0)
	if !rebuildTimer.Stop() {
		<-rebuildTimer.C
	}
	rebuildPending := false
	debounceDelay := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				ext := filepath.Ext(event.Name)
				if ext == ".go" || ext == ".mod" || ext == ".sum" {
					fmt.Printf("File changed (%s), scheduling rebuild...\n", event.Name)
					// Reset the timer to debounce multiple rapid changes
					if !rebuildTimer.Stop() && rebuildPending {
						<-rebuildTimer.C
					}
					rebuildTimer.Reset(debounceDelay)
					rebuildPending = true
				}
			}
		case <-rebuildTimer.C:
			if rebuildPending {
				err := onRebuild()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error during rebuild: %v\n", err)
				}
				rebuildPending = false
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
}

// watchAndRebuild watches Go source files and rebuilds the WASM bundle on change.
func watchAndRebuild(appDir string) {
	err := watchFiles(appDir, func() error {
		fmt.Println("Rebuilding...")
		newBuildDir, err := buildWASM(appDir)
		if err != nil {
			return fmt.Errorf("error rebuilding WASM: %w", err)
		}
		buildMutex.Lock()
		old := currentBuildDir
		currentBuildDir = newBuildDir
		buildMutex.Unlock()
		if err := os.RemoveAll(old); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old build directory: %v\n", err)
		}
		fmt.Println("Rebuild complete")
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
	}
}

// wasmHandler serves the bundle.wasm file from the current build directory.
func wasmHandler(w http.ResponseWriter, r *http.Request) {
	buildMutex.RLock()
	dirPath := currentBuildDir
	buildMutex.RUnlock()
	http.ServeFile(w, r, filepath.Join(dirPath, "bundle.wasm"))
}

// wasmExecHandler serves the wasm_exec.js file from the current build directory.
func wasmExecHandler(w http.ResponseWriter, r *http.Request) {
	buildMutex.RLock()
	dirPath := currentBuildDir
	buildMutex.RUnlock()
	http.ServeFile(w, r, filepath.Join(dirPath, "wasm_exec.js"))
}

// indexHandler serves the indexHTML template directly.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Only serve index for root path and paths that don't match other handlers
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(indexHTML)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write response: %v\n", err)
	}
}

// runServe builds the WASM bundle and serves the app with auto-rebuild.
func runServe(cmd *cobra.Command, args []string) error {
	// Determine app directory (optional positional argument)
	if len(args) > 0 {
		serveDir = args[0]
	} else {
		serveDir = "."
	}
	// Validate app directory
	info, err := os.Stat(serveDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid app directory: %s", serveDir)
	}

	// Set up environment for package validation
	env := os.Environ()
	if shouldDisableWorkspace(serveDir) {
		env = append(env, "GOWORK=off")
	}

	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  serveDir,
		Env:  env,
	}
	pkgs, _ := packages.Load(cfg, ".")
	if len(pkgs) == 0 || pkgs[0].Name != "main" {
		return fmt.Errorf("serve directory %s is not package main", serveDir)
	}
	fmt.Printf("Building WASM bundle in %s...\n", serveDir)
	buildDir, err := buildWASM(serveDir)
	if err != nil {
		return fmt.Errorf("error building WASM: %w", err)
	}
	buildMutex.Lock()
	currentBuildDir = buildDir
	buildMutex.Unlock()
	go watchAndRebuild(serveDir)

	// Find and reserve a free port
	actualPort, listener, err := findFreePort(servePort)
	if err != nil {
		return fmt.Errorf("error finding free port: %w", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close listener: %v\n", err)
		}
	}()

	fmt.Printf("Serving Thunder app on port %d (watching %s)...\n", actualPort, serveDir)

	// Set up HTTP handlers
	http.HandleFunc("/bundle.wasm", wasmHandler)
	http.HandleFunc("/wasm_exec.js", wasmExecHandler)
	http.HandleFunc("/", indexHandler)

	// Start the server in a goroutine so we can open browser after it starts
	server := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverStarted := make(chan bool, 1)
	serverErr := make(chan error, 1)

	go func() {
		// Signal that we're about to start the server
		serverStarted <- true
		serverErr <- server.Serve(listener)
	}()

	// Wait for server to start, then open browser
	urlStr := fmt.Sprintf("http://localhost:%d", actualPort)
	go func() {
		select {
		case <-serverStarted:
			// Give server a moment to fully initialize
			time.Sleep(100 * time.Millisecond)
			// Try to open browser
			if err := open(urlStr); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", err)
			}
		case err := <-serverErr:
			// Server failed to start immediately, don't open browser
			if err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
			}
		}
	}()

	// Wait for server to finish or fail
	return <-serverErr
}

var openCommands = map[string][]string{
	"windows": []string{"cmd", "/c", "start"},
	"darwin":  []string{"open"},
	"linux":   []string{"xdg-open"},
}

func open(uri string) error {
	run, ok := openCommands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	if len(run) == 0 {
		return fmt.Errorf("no command specified for platform %s", runtime.GOOS)
	}
	if runtime.GOOS == "windows" {
		uri = strings.ReplaceAll(uri, "&", "^&")
	}
	run = append(run, uri)
	cmd := exec.Command(run[0], run[1:]...) //nolint:gosec // G204: Safe to execute platform-specific open commands with validated arguments
	return cmd.Start()
}
