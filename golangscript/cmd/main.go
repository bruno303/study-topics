package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bruno303/study-topics/golangscript/internal/debugger"
	"github.com/bruno303/study-topics/golangscript/internal/lang"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	fmt.Println("Welcome to GoLangScript!")
	env := lang.NewEnvironment()
	debugger := debugger.NewDebugger()

	scanner := bufio.NewScanner(os.Stdin)
	lines := []string{}

	// Input phase
	fmt.Println("Enter your program. Type `run`")
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		if strings.TrimSpace(line) == "run" {
			break
		}

		lines = append(lines, line)
	}

	// Debug shell before execution
	if *debug {
		debuggerShell(lines, env, debugger)
	} else {
		runProgram(lines, env)
	}
}

func debuggerShell(lines []string, env *lang.Environment, debugger *debugger.Debugger) {
	fmt.Println("Entered debugger shell. Type `start` to run the program.")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("(debug) ")
		cmdLine, _ := reader.ReadString('\n')
		cmd := strings.TrimSpace(cmdLine)

		switch {
		case cmd == "list":
			for i, line := range lines {
				fmt.Printf("%2d: %s\n", i+1, line)
			}

		case strings.HasPrefix(cmd, "break "):
			parts := strings.SplitN(cmd, " ", 2)
			if len(parts) == 2 {
				lineNum, err := strconv.Atoi(parts[1])
				if err == nil {
					debugger.GetBreakpoints()[lineNum] = true
					fmt.Printf("Breakpoint set at line %d\n", lineNum)
				} else {
					fmt.Println("Invalid line number")
				}
			}

		case cmd == "start":
			runProgramDebugAware(lines, env, debugger)
			return

		case cmd == "env":
			for k, v := range env.GetVars() {
				fmt.Printf("%s = %v\n", k, v)
			}

		case cmd == "exit":
			os.Exit(0)

		default:
			fmt.Println("Commands: list | break <line> | env | start | exit")
		}
	}
}

func runProgramDebugAware(lines []string, env *lang.Environment, debugger *debugger.Debugger) {
	for i, line := range lines {
		lineNumber := i + 1
		if debugger.ShouldPause(lineNumber) {
			fmt.Printf("Paused at line %d: %s\n", lineNumber, line)
			debugger.HandleDebuggerCommand(env)
		}
		if err := processCommand(line, env, lineNumber); err != nil {
			continue
		}
	}
}

func runProgram(lines []string, env *lang.Environment) {
	for i, line := range lines {
		if err := processCommand(line, env, i+1); err != nil {
			continue
		}
	}
}

func processCommand(line string, env *lang.Environment, lineNumber int) error {
	result, err := lang.Evaluate(line, env)
	if err != nil {
		fmt.Printf("[Line %d] Error: %v\n", lineNumber, err)
		return err
	}
	if result != nil {
		fmt.Println(result)
	}
	return nil
}
