// Command check-clickhouse-migrations inspects and normalizes ClickHouse migration state.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/agenttrace/agenttrace/api/internal/migration"
)

func main() {
	databaseURL := flag.String("database-url", "", "ClickHouse migration database URL")
	printMigrationURL := flag.Bool("print-migration-url", false, "print a normalized golang-migrate URL")
	flag.Parse()

	if *databaseURL == "" {
		fmt.Fprintln(os.Stderr, "database-url is required")
		os.Exit(2)
	}

	normalizedURL, err := migration.NormalizeClickHouseURL(*databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid ClickHouse database URL: %v\n", err)
		os.Exit(2)
	}
	if *printMigrationURL {
		fmt.Println(normalizedURL)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	state, err := migration.InspectClickHouseSchema(ctx, normalizedURL)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(state)
}
