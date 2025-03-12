package process

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
			log.Fatal("Error: Empty command")
		}
		executable := parts[0]
		args := parts[1:]
		internalCommands = append(internalCommands, internalCommand{executable: executable, args: args})
	}

	return &Command{commands: internalCommands, signal: signal}
}

func (c *Command) Run() {
	c.stopPreviousCommand()

	for _, command := range c.commands {
		c.runSingleCommand(command)
	}
}

func (c *Command) runSingleCommand(command internalCommand) {
	cmd := exec.Command(command.executable, command.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.currentCmd = cmd

	fmt.Printf("Executing command: %s %s\n", command.executable, strings.Join(command.args, " "))
	err := cmd.Start()
	if err != nil {
		log.Println("Error starting command:", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Println("Error executing command:", err)
	} else {
		fmt.Println("Command completed successfully")
	}
}

func (c *Command) stopPreviousCommand() {
	if c.currentCmd != nil && c.currentCmd.Process != nil {
		fmt.Println("Stopping previous command...")
		err := c.currentCmd.Process.Signal(c.signal)
		if err != nil && !errors.Is(err, os.ErrProcessDone) {
			log.Println("Error stopping previous command:", err)
		}
	}
}

func (c *Command) Stop() {
	c.stopPreviousCommand()
}
