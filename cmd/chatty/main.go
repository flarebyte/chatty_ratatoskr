package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flarebyte/chatty-ratatoskr/internal/httpapi"
	"github.com/flarebyte/chatty-ratatoskr/internal/runtimeconfig"
	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
	"github.com/spf13/cobra"
)

const repoURL = "https://github.com/flarebyte/chatty-ratatoskr"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type cliOptions struct {
	serveReady func(string)
	listen     func(network, address string) (net.Listener, error)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := execute(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		mustWrite(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return executeWithOptions(ctx, args, stdout, stderr, cliOptions{})
}

func executeWithOptions(ctx context.Context, args []string, stdout, stderr io.Writer, options cliOptions) error {
	root := newRootCmd(ctx, stdout, stderr, options)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		return fmt.Errorf("%s; run 'chatty --help' for usage", err.Error())
	}
	return err
}

func newRootCmd(ctx context.Context, stdout, stderr io.Writer, options cliOptions) *cobra.Command {
	var versionFlag bool

	root := &cobra.Command{
		Use:           "chatty",
		Short:         "chatty runs the Yggdrasil mock-server CLI",
		Long:          "chatty is the CLI for the flarebyte/chatty-ratatoskr project.\nIt is intended to run a lightweight Yggdrasil mock server for local development and CI.\nRepository: " + repoURL,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				_, err := io.WriteString(cmd.OutOrStdout(), versionString())
				return err
			}
			return cmd.Help()
		},
	}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.Flags().BoolVar(&versionFlag, "version", false, "Print version information")
	root.AddCommand(newVersionCmd(stdout))
	root.AddCommand(newServeCmd(ctx, stdout, options))

	return root
}

func newVersionCmd(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print version, commit, and build time",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := io.WriteString(stdout, versionString())
			return err
		},
	}
	cmd.SetOut(stdout)
	return cmd
}

func newServeCmd(ctx context.Context, stdout io.Writer, options cliOptions) *cobra.Command {
	var (
		configPath string
		listenAddr string
	)

	cmd := &cobra.Command{
		Use:           "serve",
		Short:         "Start the mock-server HTTP process",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeconfig.LoadServeConfig(ctx, configPath)
			if err != nil {
				return err
			}
			if listenAddr != "" {
				cfg.Listen = listenAddr
			}
			return runServeWithOptions(ctx, stdout, cfg, options)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to a strict JSON config file")
	cmd.Flags().StringVar(&listenAddr, "listen", "", "HTTP listen address override")
	return cmd
}

func runServeWithOptions(ctx context.Context, stdout io.Writer, cfg runtimeconfig.ServeConfig, options cliOptions) error {
	if err := runtimeconfig.ValidateServeConfig(cfg); err != nil {
		return err
	}

	listenFn := options.listen
	if listenFn == nil {
		listenFn = net.Listen
	}

	listener, err := listenFn("tcp", cfg.Listen)
	if err != nil {
		return fmt.Errorf("listen on %q: %w", cfg.Listen, err)
	}
	defer func() {
		_ = listener.Close()
	}()

	server := &http.Server{
		Handler: newServerMux(cfg),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if _, err := io.WriteString(stdout, fmt.Sprintf("chatty serve listening on %s\n", listener.Addr().String())); err != nil {
		return err
	}
	if options.serveReady != nil {
		options.serveReady(listener.Addr().String())
	}

	err = server.Serve(listener)
	<-shutdownDone
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func versionString() string {
	return fmt.Sprintf("chatty version=%s commit=%s date=%s\n", Version, Commit, Date)
}

func newServerMux(cfg runtimeconfig.ServeConfig) *http.ServeMux {
	mux := http.NewServeMux()
	store := snapshot.NewInMemoryStore()
	httpapi.NewSnapshotAPI(store).Register(mux)
	httpapi.NewNodeAPI(store).Register(mux)
	httpapi.NewCreateAPI().Register(mux)
	httpapi.NewAdminAPI(store).Register(mux)
	if cfg.WebSocketEnabled {
		httpapi.NewEventsAPI([]string{
			"tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		}).Register(mux)
	}
	return mux
}

func mustWrite(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		panic(err)
	}
}
