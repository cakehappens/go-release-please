package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/cakehappens/go-release-please/internal/cmds/root"
)

func main() {
	var (
		ctx    = context.Background()
		args   = os.Args[1:]
		stdin  = os.Stdin
		stdout = os.Stdout
		stderr = os.Stderr
		err    = exec(ctx, args, stdin, stdout, stderr)
	)
	switch {
	case err == nil, errors.Is(err, ff.ErrHelp), errors.Is(err, ff.ErrNoExec):
		// no problem
	case err != nil:
		fmt.Fprintf(stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func exec(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	var g run.Group
	g.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))
	g.Add(run.ContextHandler(ctx))

	var (
		rootCmd = root.New(stdout, stderr)
		//_    = createcmd.New(root)
	)

	// always print help
	defer func() {
		if err != nil {
			fmt.Fprintf(stderr, "\n%s\n", ffhelp.Command(rootCmd.Command))
		}
	}()

	rootCtx, cancel := context.WithCancel(ctx)
	g.Add(func() error {
		if err := rootCmd.Command.Parse(
			args,
			ff.WithEnvVars(),
		); err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		if err := rootCmd.Command.Run(rootCtx); err != nil {
			return fmt.Errorf("run: %w", err)
		}
		return nil
	}, func(err error) {
		cancel()
	})

	return g.Run()
}
