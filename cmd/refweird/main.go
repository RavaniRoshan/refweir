// Command refweird is the primary server daemon for refweir.dev.
// It hosts the Smart-HTTP Git transport, REST API, Model Context Protocol (MCP) server,
// and supports offline administrator bootstrap.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"refweir.dev/internal/app"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/config"
	"refweir.dev/internal/security"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("refweird %s\n", version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	cfg.Version = version

	// Check for CLI subcommands
	if len(os.Args) > 1 && os.Args[1] == "hook" {
		if len(os.Args) > 2 && os.Args[2] == "pre-receive" {
			closeAudit := openHookAudit()
			hookErr := security.PreReceiveHook(os.Stdin, os.Stderr)
			closeAudit()
			if hookErr != nil {
				os.Exit(1)
			}
			return
		}
		fmt.Fprintln(os.Stderr, "Usage: refweird hook pre-receive")
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "bootstrap" {
		fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
		email := fs.String("email", "", "Admin user email address (required)")
		namespace := fs.String("namespace", "", "Initial admin namespace slug (required)")
		dataDir := fs.String("data-dir", cfg.DataDir, "Path to refweir.dev data directory")
		if err := fs.Parse(os.Args[2:]); err != nil {
			log.Fatalf("bootstrap: %v", err)
		}

		if *email == "" || *namespace == "" {
			fmt.Fprintln(os.Stderr, "Usage: refweird bootstrap --email <email> --namespace <slug> [--data-dir <path>]")
			os.Exit(1)
		}

		if os.Getenv("REFWEIR_TELEMETRY_OPT_IN") == "" {
			if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
				fmt.Print("Send an anonymous activation metric? [y/N] ")
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				choice := "0"
				switch strings.ToLower(strings.TrimSpace(answer)) {
				case "y", "yes":
					choice = "1"
				}
				_ = os.MkdirAll(*dataDir, 0o700)
				_ = os.WriteFile(filepath.Join(*dataDir, "telemetry_opt_in"), []byte(choice), 0o600)
			}
		}

		cfg.DataDir = *dataDir
		res, err := app.Bootstrap(context.Background(), cfg, *email, *namespace)
		if err != nil {
			log.Fatalf("bootstrap failed: %v", err)
		}

		fmt.Println("refweir.dev bootstrap complete!")
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("User:       %s (id: %s)\n", res.User.Email, res.User.ID)
		fmt.Printf("Namespace:  %s (id: %s)\n", res.Namespace.Slug, res.Namespace.ID)
		fmt.Printf("Repository: %s\n", res.Repo.FullName)
		fmt.Printf("API Key:    %s\n", res.RawKey)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("IMPORTANT: Store this API key securely. It will not be shown again.")
		fmt.Println()
		fmt.Println("=== Agent Configuration Snippets ===")
		fmt.Println()
		fmt.Println("Claude Code (~/.claude/settings.json or command line):")
		fmt.Printf("  claude mcp add refweir %s/mcp --header \"Authorization: Bearer %s\"\n\n", cfg.PublicURL, res.RawKey)
		fmt.Println("  Or add to ~/.claude/settings.json:")
		fmt.Printf(`  {
    "mcpServers": {
      "refweir": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        }
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Cursor (.cursor/mcp.json):")
		fmt.Printf(`  {
    "mcpServers": {
      "refweir": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        }
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Cline (cline_mcp_settings.json):")
		fmt.Printf(`  {
    "mcpServers": {
      "refweir": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        },
        "disabled": false,
        "autoApprove": []
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Codex (TOML):")
		fmt.Printf(`  [mcp.servers.refweir]
  url = "%s/mcp"
  headers = { Authorization = "Bearer %s" }`+"\n", cfg.PublicURL, res.RawKey)
		return
	}

	// Durable audit trail: append every event to a JSON-lines file alongside the
	// database. The in-memory ring buffer remains available to the admin API.
	// Set REFWEIR_AUDIT_LOG to a path to override, or "off" to disable the file.
	auditPath := os.Getenv("REFWEIR_AUDIT_LOG")
	if auditPath == "" {
		auditPath = filepath.Join(cfg.DataDir, "audit.log")
	}
	if auditPath != "off" {
		if mkErr := os.MkdirAll(filepath.Dir(auditPath), 0o700); mkErr != nil {
			log.Printf("refweird: cannot create audit log directory: %v", mkErr)
		}
		auditLogger, aerr := audit.NewFileLogger(auditPath)
		if aerr != nil {
			log.Printf("refweird: file audit log unavailable (%v); audit stays on stderr and in memory", aerr)
		} else {
			audit.SetDefaultLogger(auditLogger)
			// Child processes (the pre-receive hook invoked by git) inherit the
			// environment; point them at the same durable log.
			_ = os.Setenv("REFWEIR_AUDIT_LOG", auditPath)
			defer func() {
				if cerr := auditLogger.Close(); cerr != nil {
					log.Printf("refweird: error closing audit log: %v", cerr)
				}
			}()
			fmt.Printf("refweird audit log: %s\n", auditPath)
		}
	}

	application, err := app.Open(cfg)
	if err != nil {
		log.Fatalf("refweird: failed to initialize application: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("refweird: error closing application: %v", err)
		}
	}()

	if cfg.SSHAddr != "" && cfg.SSHAddr != "none" {
		if err := application.StartSSH(); err != nil {
			log.Printf("refweird: failed to start SSH server on %s: %v", cfg.SSHAddr, err)
		} else {
			fmt.Printf("refweird SSH listening on %s\n", cfg.SSHAddr)
		}
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("refweird listening on %s (publicURL=%s gitURL=%s dataDir=%s sshAddr=%s)\n",
			cfg.ListenAddr, cfg.PublicURL, cfg.GitURL, cfg.DataDir, cfg.SSHAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Fatalf("refweird: server error: %v", err)
	case sig := <-shutdown:
		fmt.Printf("\nrefweird: received signal %v, shutting down gracefully...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("refweird: shutdown error: %v", err)
		}
	}
}

// openHookAudit opens the shared audit log for the pre-receive subprocess. The
// hook is spawned by git from the daemon, so it inherits REFWEIR_AUDIT_LOG (set
// by the parent) or REFWEIR_DATA_DIR. It returns a no-op when no durable log is
// configured, leaving events on stderr.
func openHookAudit() func() {
	path := os.Getenv("REFWEIR_AUDIT_LOG")
	if path == "" {
		if dataDir := os.Getenv("REFWEIR_DATA_DIR"); dataDir != "" {
			path = filepath.Join(dataDir, "audit.log")
		}
	}
	if path == "" || path == "off" {
		return func() {}
	}
	logger, err := audit.NewFileLogger(path)
	if err != nil {
		return func() {}
	}
	audit.SetDefaultLogger(logger)
	return func() { _ = logger.Close() }
}
