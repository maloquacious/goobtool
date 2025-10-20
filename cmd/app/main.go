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
	"github.com/maloquacious/semver"
	"github.com/spf13/cobra"
)

var (
	version       = semver.Version{Minor: 1, Patch: 1, PreRelease: "alpha", Build: semver.Commit()}
	schemaVersion = "0.1"
	buildDate     = ""
)

var (
	port       int
	adminPort  int
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

	// Check for datastore existence
	storePath := store.GetStorePath()
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

	log.Info("datastore found at path=%s", storePath)

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
		// TODO: implement real readiness checks (store initialized, not in maintenance)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
		// TODO: readiness should return non-OK during maintenance or uninitialized store
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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: publicMux,
	}

	// Bind admin to 127.0.0.1 only (loopback enforcement)
	adminListener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", adminPort))
	if err != nil {
		log.Error("admin listener bind failed (loopback only): %v", err)
		os.Exit(1)
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

// --- DB command stubs ---

func runDBCreate(cmd *cobra.Command, args []string) {
	log.Info("db create: TODO - implement SQLite initialization")
}

func runDBUpgrade(cmd *cobra.Command, args []string) {
	log.Info("db upgrade: TODO - implement migrations")
}

func runDBVerify(cmd *cobra.Command, args []string) {
	log.Info("db verify: TODO - implement schema verification")
}
