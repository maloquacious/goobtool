package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/maloquacious/goobtool/internal/logger"
	"github.com/maloquacious/goobtool/internal/store"
	"github.com/maloquacious/goobtool/internal/store/sqlite"
	"github.com/maloquacious/semver"
	"github.com/spf13/cobra"
)

var (
	version       = semver.Version{Minor: 1, Patch: 3, PreRelease: "alpha", Build: semver.Commit()}
	schemaVersion = "0.1"
	buildDate     = ""
)

var (
	port       int
	adminPort  int
	adminHost  string
	shutdownTO time.Duration
	exitAfter  time.Duration
	publicDir  string
	log        logger.Logger = logger.Default
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Goobergine application server and admin CLI",
	}

	// Global flags
	rootCmd.PersistentFlags().DurationVar(&shutdownTO, "shutdown-timeout", 15*time.Second, "graceful shutdown timeout")
	rootCmd.PersistentFlags().StringVar(&publicDir, "public", "public", "directory for static public assets")

	// serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Goobergine server",
		Run:   runServe,
	}
	serveCmd.Flags().IntVar(&port, "port", 8080, "public HTTP port (HTML/HTMX)")
	serveCmd.Flags().IntVar(&adminPort, "admin-port", 8383, "admin HTTP port (JSON, loopback only)")
	serveCmd.Flags().StringVar(&adminHost, "admin-host", "127.0.0.1", "admin host (127.0.0.1 or ::1, loopback only)")
	serveCmd.Flags().DurationVar(&exitAfter, "exit-after", 0, "optional runtime; if set, server exits after this duration (testing)")

	// db command group
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
	}

	dbCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create and initialize the datastore",
		Run:   runDBCreate,
	}
	dbUpgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Apply migrations to current schema version",
		Run:   runDBUpgrade,
	}
	dbVerifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify schema integrity and version",
		Run:   runDBVerify,
	}

	dbCmd.AddCommand(dbCreateCmd, dbUpgradeCmd, dbVerifyCmd)
	rootCmd.AddCommand(serveCmd, dbCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runServe starts both the public (HTML) and admin (JSON) servers with graceful shutdown.
func runServe(cmd *cobra.Command, args []string) {
	log.Info("starting Goobergine server version=%s schema=%s", version.String(), schemaVersion)

	// Check datastore existence and state
	// NOTE: os.Exit is safe here - we're in initialization phase before any servers start.
	// If startup sequence changes, verify no resources need cleanup before these exits.
	storePath := store.GetStorePath()
	dbPath := store.GetDBPath(storePath)
	exists, err := store.CheckExists(storePath)
	if err != nil {
		log.Error("failed to check datastore: %v", err)
		os.Exit(1)
	}

	if !exists {
		log.Error("datastore not found at path=%s", storePath)
		fmt.Fprintln(os.Stderr, "\nDatastore not initialized.")
		fmt.Fprintf(os.Stderr, "Run: %s db create\n\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	// Open database and check state
	st := sqlite.New(dbPath, schemaVersion)
	if err := st.Open(); err != nil {
		log.Error("failed to open datastore: %v", err)
		os.Exit(1)
	}
	defer st.Close()

	state, err := st.CheckState()
	if err != nil {
		log.Error("failed to check datastore state: %v", err)
		os.Exit(1)
	}

	// Handle uninitialized or mismatched store
	if state == store.StateUninitialized {
		log.Warn("datastore uninitialized (missing schema_migrations table)")
		serveInstallationApp(port, adminPort, adminHost, exitAfter, shutdownTO)
		return
	}

	if state == store.StateVersionMismatch {
		actualVersion, _ := st.GetSchemaVersion()
		log.Warn("datastore version mismatch: expected=%s actual=%s", schemaVersion, actualVersion)
		serveInstallationApp(port, adminPort, adminHost, exitAfter, shutdownTO)
		return
	}

	log.Info("datastore ready path=%s schema=%s", dbPath, schemaVersion)

	publicMux := http.NewServeMux()
	adminMux := http.NewServeMux()

	// --- Public routes (HTML/HTMX) ---
	publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve /public (index.html) by default
		fp := filepath.Join(publicDir, "index.html")
		http.ServeFile(w, r, fp)
	})

	publicMux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	publicMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Store is initialized and ready (we checked state earlier)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	publicMux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"appVersion":    version.String(),
			"schemaVersion": schemaVersion,
			"goVersion":     "go1.25.2",
			"buildDate":     buildDate,
		})
	})

	// Static under /public/* (maps to ./public)
	publicMux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(publicDir))))

	// --- Admin routes (JSON-only, loopback only) ---
	adminMux.Handle("/admin/echo", jsonOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Echo string `json:"echo"`
		}
		if r.Method == http.MethodGet {
			// Support GET with ?q= for simple testing, still require Accept: application/json
			q := r.URL.Query().Get("q")
			_ = json.NewEncoder(w).Encode(map[string]string{"echo": q})
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"echo": payload.Echo})
	})))

	adminMux.Handle("/admin/status", jsonOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC().Format(time.RFC3339)
		resp := map[string]string{
			"version":       version.String(),
			"schemaVersion": schemaVersion,
			"buildDate":     buildDate,
			"time":          now,
			"mode":          "running",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})))

	adminMux.Handle("/admin/shutdown", jsonOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: coordinate shutdown via context cancellation signal channel
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "shutting down"})
		go func() {
			// give the response a moment to flush
			time.Sleep(200 * time.Millisecond)
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(os.Interrupt)
		}()
	})))

	adminMux.Handle("/admin/restart", jsonOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement real restart (requires external supervisor). For now, exit 0.
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
	})))

	// HTTP servers
	publicSrv := &http.Server{
		Addr:    net.JoinHostPort("", fmt.Sprintf("%d", port)),
		Handler: publicMux,
	}

	// Validate admin host is loopback before binding
	// NOTE: os.Exit is safe here - we're in initialization phase before any servers start.
	// If startup sequence changes (e.g., open DB earlier), verify no resources need cleanup.
	adminIP := net.ParseIP(adminHost)
	if adminIP == nil || !adminIP.IsLoopback() {
		log.Error("admin host must be loopback (127.0.0.1 or ::1), got: %s", adminHost)
		os.Exit(1)
	}

	// Bind admin to loopback only (127.0.0.1 for IPv4, ::1 for IPv6)
	adminAddr := net.JoinHostPort(adminHost, fmt.Sprintf("%d", adminPort))
	adminListener, err := net.Listen("tcp", adminAddr)
	if err != nil {
		log.Error("admin listener bind failed (loopback only): %v", err)
		os.Exit(1)
	}

	// Verify loopback-only binding (defense in depth)
	if addr, ok := adminListener.Addr().(*net.TCPAddr); ok {
		if !addr.IP.IsLoopback() {
			log.Error("admin listener bound to non-loopback address: %s", addr.IP)
			adminListener.Close()
			os.Exit(1)
		}
		log.Info("admin listener verified on loopback: %s", addr.String())
	}

	adminSrv := &http.Server{
		Handler: adminMux,
	}

	// Run servers
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		log.Info("public server listening on port=%d", port)
		if err := publicSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("public server error: %w", err)
		}
	}()

	go func() {
		log.Info("admin server listening on 127.0.0.1:%d (JSON-only)", adminPort)
		if err := adminSrv.Serve(adminListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("admin server error: %w", err)
		}
	}()

	// Optional run timer
	if exitAfter > 0 {
		go func() {
			log.Info("exit-after timer set duration=%s", exitAfter)
			time.Sleep(exitAfter)
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(os.Interrupt)
		}()
	}

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("server error: %v", err)
	}

	log.Info("initiating graceful shutdown timeout=%s", shutdownTO)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTO)
	defer cancel()

	_ = publicSrv.Shutdown(shutdownCtx)
	_ = adminSrv.Shutdown(shutdownCtx)
	log.Info("shutdown complete")
}

// serveInstallationApp serves a minimal installation/maintenance page.
func serveInstallationApp(port, adminPort int, adminHost string, exitAfter, shutdownTO time.Duration) {
	log.Info("serving installation app (datastore requires attention)")

	publicMux := http.NewServeMux()
	adminMux := http.NewServeMux()

	// Serve installation page
	publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fp := filepath.Join(publicDir, "install.html")
		http.ServeFile(w, r, fp)
	})

	// Health endpoints
	publicMux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	publicMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Not ready - datastore needs attention
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOT_READY"))
	})

	// Version endpoint
	publicMux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"appVersion":    version.String(),
			"schemaVersion": schemaVersion,
			"goVersion":     "go1.25.2",
			"buildDate":     buildDate,
		})
	})

	// Admin status endpoint (minimal)
	adminMux.Handle("/admin/status", jsonOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC().Format(time.RFC3339)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"version":       version.String(),
			"schemaVersion": schemaVersion,
			"buildDate":     buildDate,
			"time":          now,
			"mode":          "installation",
		})
	})))

	// Setup servers (same as regular runServe)
	publicSrv := &http.Server{
		Addr:    net.JoinHostPort("", fmt.Sprintf("%d", port)),
		Handler: publicMux,
	}

	adminIP := net.ParseIP(adminHost)
	if adminIP == nil || !adminIP.IsLoopback() {
		log.Error("admin host must be loopback (127.0.0.1 or ::1), got: %s", adminHost)
		os.Exit(1)
	}

	adminAddr := net.JoinHostPort(adminHost, fmt.Sprintf("%d", adminPort))
	adminListener, err := net.Listen("tcp", adminAddr)
	if err != nil {
		log.Error("admin listener bind failed (loopback only): %v", err)
		os.Exit(1)
	}

	if addr, ok := adminListener.Addr().(*net.TCPAddr); ok {
		if !addr.IP.IsLoopback() {
			log.Error("admin listener bound to non-loopback address: %s", addr.IP)
			adminListener.Close()
			os.Exit(1)
		}
		log.Info("admin listener verified on loopback: %s", addr.String())
	}

	adminSrv := &http.Server{
		Handler: adminMux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		log.Info("public server listening on port=%d (installation mode)", port)
		if err := publicSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("public server error: %w", err)
		}
	}()

	go func() {
		log.Info("admin server listening on 127.0.0.1:%d (JSON-only)", adminPort)
		if err := adminSrv.Serve(adminListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("admin server error: %w", err)
		}
	}()

	if exitAfter > 0 {
		go func() {
			log.Info("exit-after timer set duration=%s", exitAfter)
			time.Sleep(exitAfter)
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(os.Interrupt)
		}()
	}

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("server error: %v", err)
	}

	log.Info("initiating graceful shutdown timeout=%s", shutdownTO)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTO)
	defer cancel()

	_ = publicSrv.Shutdown(shutdownCtx)
	_ = adminSrv.Shutdown(shutdownCtx)
	log.Info("shutdown complete")
}

// jsonOnly enforces JSON-only contract for admin routes.
func jsonOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Require Accept: application/json (at least for admin)
		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, "application/json") && accept != "" {
			writeJSONError(w, http.StatusNotAcceptable, "not_acceptable", "Accept must include application/json")
			return
		}
		if r.Method != http.MethodGet && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			writeJSONError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSONError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": msg,
	})
}

// --- DB command implementations ---

func runDBCreate(cmd *cobra.Command, args []string) {
	log.Info("creating datastore schema=%s", schemaVersion)

	storePath := store.GetStorePath()
	dbPath := store.GetDBPath(storePath)

	// Check if database already exists
	exists, err := store.CheckExists(storePath)
	if err != nil {
		log.Error("failed to check datastore: %v", err)
		os.Exit(1)
	}

	if exists {
		log.Error("datastore already exists at path=%s", dbPath)
		fmt.Fprintln(os.Stderr, "\nDatastore already exists.")
		fmt.Fprintf(os.Stderr, "Path: %s\n\n", dbPath)
		os.Exit(1)
	}

	// Create and initialize the database
	st := sqlite.New(dbPath, schemaVersion)
	if err := st.Open(); err != nil {
		log.Error("failed to open database: %v", err)
		os.Exit(1)
	}
	defer st.Close()

	if err := st.InitSchema(schemaVersion); err != nil {
		log.Error("failed to initialize schema: %v", err)
		// Clean up partial creation
		st.Close()
		os.Remove(dbPath)
		os.Exit(1)
	}

	log.Info("datastore created successfully path=%s schema=%s", dbPath, schemaVersion)
	fmt.Fprintf(os.Stdout, "\n✓ Datastore created successfully\n")
	fmt.Fprintf(os.Stdout, "  Path: %s\n", dbPath)
	fmt.Fprintf(os.Stdout, "  Schema version: %s\n\n", schemaVersion)
}

func runDBUpgrade(cmd *cobra.Command, args []string) {
	log.Info("db upgrade: TODO - implement migrations")
}

func runDBVerify(cmd *cobra.Command, args []string) {
	log.Info("db verify: TODO - implement schema verification")
}
