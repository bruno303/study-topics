package main

import (
	"file-watcher/internal/config"
	"file-watcher/internal/fs"
	"file-watcher/internal/process"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	cfg := parseInput()

	cmd := process.NewCommand(cfg.Commands, cfg.Signal)

	// Handle interrupt signals (e.g., Ctrl+C) to clean up resources
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		cmd.Stop()
		os.Exit(0)
	}()

	go func() {
		cmd.Run()
	}()

	fs.StartMonitor(cfg.FilesToWatch, cfg.Delay, func() error {
		cmd.Run()
		return nil
	})
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
