package main

import (
	"context"
	"file-watcher/internal/config"
	"file-watcher/internal/monitor"
	"file-watcher/internal/process"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/bruno303/go-toolkit/pkg/log"
)

type commandsFlag []string

func (c *commandsFlag) String() string {
	return strings.Join(*c, ", ")
}

func (c *commandsFlag) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func main() {
	ctx := context.Background()
	cfg := parseInput()

	log.SetLogger(log.NewSlogAdapter(log.SlogAdapterOpts{
		Level:      log.LevelDebug,
		FormatJson: false,
		Source:     "default",
	}))

	cmd := process.NewCommand(cfg.Commands, cfg.Signal)
	m := monitor.NewMonitor(cfg.FilesToWatch, cfg.Delay, func() error {
		cmd.Run(ctx)
		return nil
	})

	manager := monitor.NewMonitorManager(m, cmd)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		manager.Stop(ctx)
		os.Exit(0)
	}()

	manager.Start(ctx)
}

func parseInput() config.Config {
	filesFlag := flag.String("files", "", "Comma-separated list of files or folders to watch (supports glob patterns)")
	var commands commandsFlag
	flag.Var(&commands, "command", "Command to execute when a file changes (can be used multiple times)")
	delayFlag := flag.Int("delay", 0, "Delay in seconds before re-executing the command")
	signalFlag := flag.String("signal", "SIGINT", "Signal to send to the child process (e.g., SIGKILL, SIGTERM)")
	flag.Parse()

	if *filesFlag == "" || len(commands) == 0 {
		fmt.Println("Error: --files and --command flags are required")
		flag.Usage()
		os.Exit(1)
	}

	files := strings.Split(*filesFlag, ",")

	var signalToSend os.Signal
	switch *signalFlag {
	case "SIGKILL":
		signalToSend = syscall.SIGKILL
	case "SIGTERM":
		signalToSend = syscall.SIGTERM
	case "SIGINT":
		signalToSend = syscall.SIGINT
	default:
		fmt.Println("Error: Unsupported signal. Use SIGKILL, SIGTERM, or SIGINT")
		os.Exit(1)
	}

	return config.NewConfig(files, commands, *delayFlag, signalToSend)
}
