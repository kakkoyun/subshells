package main

import (
	"context"
	"errors"
	"os"

	"github.com/go-kit/log/level"
	"github.com/kakkoyun/subshells/pkg/logger"
	"github.com/oklog/run"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	var g run.Group
	logger := logger.NewLogger("debug", logger.LogFormatLogfmt, "infiniteloop")
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
				level.Info(logger).Log("msg", "context done")
				return nil
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
