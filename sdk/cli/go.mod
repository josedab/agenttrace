module github.com/agenttrace/agenttrace-cli

go 1.21

require (
	github.com/agenttrace/agenttrace-go v0.1.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/google/uuid v1.6.0
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.16.0 // indirect
)

replace github.com/agenttrace/agenttrace-go => ../go
