package main

import (
	"context"
	"errors"
	"os"

	"github.com/alecthomas/kong"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"

	"github.com/kakkoyun/subshells/pkg/logger"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type Flags struct {
	LogLevel string `default:"info" enum:"error,warn,info,debug" help:"log level."`
}

func main() {
	flags := &Flags{}
	_ = kong.Parse(flags)

	var g run.Group
	logger := logger.NewLogger(flags.LogLevel, logger.LogFormatLogfmt, "infiniteloop")
	level.Debug(logger).Log("msg", "subshells initialized",
		"version", version,
		"commit", commit,
		"date", date,
		"builtBy", builtBy,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Add(func() error {
		level.Info(logger).Log("msg", "starting the loop")
		for {
			select {
			case <-ctx.Done():
				level.Debug(logger).Log("msg", "context done")
				return nil
			default:
				iter(ctx, logger)
			}
		}
	}, func(err error) {
		cancel()
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

//go:noinline
func iter(ctx context.Context, logger log.Logger) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	level.Debug(logger).Log("msg", "looping")
}
