package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/go-kit/log/level"
	"github.com/kakkoyun/subshells/pkg/logger"
	"github.com/metalmatze/signal/healthcheck"
	"github.com/metalmatze/signal/internalserver"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type Flags struct {
	LogLevel string `default:"info" enum:"error,warn,info,debug" help:"log level."`
	Address  string `default:":8080" help:"Address string for internal server"`
}

func main() {
	flags := &Flags{}
	_ = kong.Parse(flags)

	logger := logger.NewLogger(flags.LogLevel, logger.LogFormatLogfmt, "subshells")
	level.Debug(logger).Log("msg", "subshells initialized",
		"version", version,
		"commit", commit,
		"date", date,
		"builtBy", builtBy,
	)

	registry := prometheus.NewRegistry()
	healthchecks := healthcheck.NewMetricsHandler(healthcheck.NewHandler(), registry)
	h := internalserver.NewHandler(
		internalserver.WithHealthchecks(healthchecks),
		internalserver.WithPrometheusRegistry(registry),
		internalserver.WithPProf(),
	)
	s := http.Server{
		Addr:    flags.Address,
		Handler: h,
	}

	var g run.Group

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Add(func() error {
		// This is the subshell that runs the command.
		if err := runShell("./bin/infiniteloop"); err != nil {
			return err
		}
		return nil
	}, func(err error) {
		cancel()
	})

	g.Add(func() error {
		level.Info(logger).Log("msg", "starting internal HTTP server", "address", s.Addr)
		return s.ListenAndServe()
	}, func(err error) {
		_ = s.Shutdown(context.Background())
	})

	g.Add(run.SignalHandler(ctx, os.Interrupt, os.Kill))
	if err := g.Run(); err != nil {
		var e run.SignalError
		if errors.As(err, &e) {
			level.Error(logger).Log("msg", "program exited with signal", "err", err, "signal", e.Signal)
		} else {
			level.Error(logger).Log("msg", "program exited with error", "err", err)
		}
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "exited")
}

func runShell(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
