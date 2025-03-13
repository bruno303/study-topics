package process

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	log "github.com/bruno303/go-toolkit/pkg/log"
)

type Command struct {
	commands   []internalCommand
	signal     os.Signal
	currentCmd *exec.Cmd
}

type internalCommand struct {
	executable string
	args       []string
}

func NewCommand(commands []string, signal os.Signal) *Command {
	var internalCommands []internalCommand
	for _, c := range commands {
		parts := strings.Fields(c)
		if len(parts) == 0 {
			log.Log().Error(context.Background(), "Empty command received", errors.New("empty command"))
			os.Exit(1)
		}
		executable := parts[0]
		args := parts[1:]
		internalCommands = append(internalCommands, internalCommand{executable: executable, args: args})
	}

	return &Command{commands: internalCommands, signal: signal}
}

func (c *Command) Run(ctx context.Context) {
	c.stopPreviousCommand(ctx)

	for _, command := range c.commands {
		c.runSingleCommand(ctx, command)
	}
}

func (c *Command) runSingleCommand(ctx context.Context, command internalCommand) {
	cmd := exec.Command(command.executable, command.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.currentCmd = cmd

	log.Log().Info(ctx, "Executing command: %s %s\n", command.executable, strings.Join(command.args, " "))
	err := cmd.Start()
	if err != nil {
		log.Log().Error(ctx, "Error starting command:", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Log().Error(ctx, "Error executing command:", err)
	} else {
		log.Log().Debug(ctx, "Command completed successfully")
	}
}

func (c *Command) stopPreviousCommand(ctx context.Context) {
	if c.currentCmd != nil && c.currentCmd.Process != nil {
		log.Log().Debug(ctx, "Stopping previous command...")
		err := c.currentCmd.Process.Signal(c.signal)
		if err != nil && !errors.Is(err, os.ErrProcessDone) {
			log.Log().Error(ctx, "Error stopping previous command:", err)
		}
	}
}

func (c *Command) Stop(ctx context.Context) {
	c.stopPreviousCommand(ctx)
}
