// Command migrate applies or rolls back AgentTrace database migrations.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/migration"
)

func main() {
	target := flag.String("database", string(migration.TargetAll), "migration target: all, postgres, or clickhouse")
	path := flag.String("path", "./migrations", "path containing postgres and clickhouse migration directories")
	waitTimeout := flag.Duration("wait-timeout", 2*time.Minute, "maximum time to wait for databases")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: migrate [flags] <up|down>")
		os.Exit(2)
	}
	action := flag.Arg(0)
	if action != "up" && action != "down" {
		fmt.Fprintf(os.Stderr, "unsupported migration action %q\n", action)
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load configuration: %v\n", err)
		os.Exit(1)
	}

	runner := migration.Runner{
		Config:      cfg,
		Path:        *path,
		WaitTimeout: *waitTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *waitTimeout+time.Minute)

	switch action {
	case "up":
		err = runner.Up(ctx, migration.Target(*target))
	case "down":
		err = runner.Down(ctx, migration.Target(*target))
	}
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}
