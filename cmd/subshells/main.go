package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/alecthomas/kong"
	"github.com/efficientgo/core/runutil"
	"github.com/go-kit/log/level"
	"github.com/metalmatze/signal/healthcheck"
	"github.com/metalmatze/signal/internalserver"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/kakkoyun/subshells/pkg/logger"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type Flags struct {
	LogLevel string        `default:"info" enum:"error,warn,info,debug" help:"log level."`
	Address  string        `default:":8080" help:"Address string for internal server"`
	Interval time.Duration `default:"1s" help:"Interval between each shell execution"`
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
	registry.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
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
		level.Info(logger).Log("msg", "running the loop in a shell")
		repeatInterval := flags.Interval
		return runutil.Repeat(repeatInterval, ctx.Done(), func() error {
			level.Info(logger).Log("msg", "new shell")

			ctx, cancel := context.WithTimeout(ctx, repeatInterval-(10*time.Millisecond))
			defer cancel()

			if err := runShell(ctx, fmt.Sprintf("./bin/infiniteloop --log-level=%s", flags.LogLevel)); err != nil {
				level.Debug(logger).Log("msg", "shell failed", "err", err)
			}
			level.Info(logger).Log("msg", "shell finished")
			return nil
		})
	}, func(_ error) {
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

func runShell(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
