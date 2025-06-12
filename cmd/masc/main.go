package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

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
	wasmExecSrc := filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js")
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
	defer file.Close()

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
	ln, err := net.Listen("tcp", ":0")
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
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// watchFiles watches Go source files and calls the provided callback on changes.
func watchFiles(appDir string, onRebuild func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up file watcher: %w", err)
	}
	defer watcher.Close()

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
		os.RemoveAll(old)
		fmt.Println("Rebuild complete")
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
	}
}

// serve starts an HTTP server on the given port, serving files from dir.
func serve(port int, dir string) error {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, nil)
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
	w.Write([]byte(indexHTML))
}

// settingsHandler serves Thunder Settings from environment variables for dev mode.
func settingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create settings response from environment variables
	settings := map[string]interface{}{
		"Google_Maps_API_Key__c": os.Getenv("GOOGLE_MAPS_API_KEY"),
		"error":                  false,
		"message":                "",
	}

	// If no API key is set, return an error
	if settings["Google_Maps_API_Key__c"] == "" {
		settings["error"] = true
		settings["message"] = "GOOGLE_MAPS_API_KEY environment variable not set"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		http.Error(w, "Failed to encode settings", http.StatusInternalServerError)
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
		return fmt.Errorf("Invalid app directory: %s", serveDir)
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
		return fmt.Errorf("Error building WASM: %w", err)
	}
	buildMutex.Lock()
	currentBuildDir = buildDir
	buildMutex.Unlock()
	go watchAndRebuild(serveDir)

	// Find and reserve a free port
	actualPort, listener, err := findFreePort(servePort)
	if err != nil {
		return fmt.Errorf("Error finding free port: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Serving Thunder app on port %d (watching %s)...\n", actualPort, serveDir)

	// Set up HTTP handlers
	http.HandleFunc("/bundle.wasm", wasmHandler)
	http.HandleFunc("/wasm_exec.js", wasmExecHandler)
	http.HandleFunc("/", indexHandler)

	// Start the server in a goroutine so we can open browser after it starts
	server := &http.Server{}
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

// sanitizeStaticResourceName converts a name to a valid static resource API name (alphanumeric, begins with letter).
func sanitizeStaticResourceName(name string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9]+`)
	name = re.ReplaceAllString(name, "")
	if len(name) == 0 {
		name = "App"
	}
	if !unicode.IsLetter(rune(name[0])) {
		name = "A" + name
	}
	return name
}

// buildProdWASM compiles the Go app in appDir to WebAssembly for production.
func buildProdWASM(appDir string) (string, error) {
	// create temporary build directory
	buildDir, err := os.MkdirTemp("", "masc-deploy-*")
	if err != nil {
		return "", err
	}
	outWasm := filepath.Join(buildDir, "bundle.wasm")
	cmd := exec.Command("go", "build", "-o", outWasm)

	// Set up environment with smart GOWORK handling
	env := append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if shouldDisableWorkspace(appDir) {
		env = append(env, "GOWORK=off")
		fmt.Printf("Note: Disabling go.work for standalone module deployment\n")
	}
	cmd.Env = env

	abs, err := filepath.Abs(appDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	cmd.Dir = abs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buildDir, nil
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
	if runtime.GOOS == "windows" {
		uri = strings.Replace(uri, "&", "^&", -1)
	}
	run = append(run, uri)
	cmd := exec.Command(run[0], run[1:]...)
	return cmd.Start()
}
