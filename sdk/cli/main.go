// AgentTrace CLI - Trace any command-line tool
package main

import (
	"os"

	"github.com/agenttrace/agenttrace-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
