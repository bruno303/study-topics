package debugger

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bruno303/study-topics/golangscript/internal/lang"
)

type Debugger struct {
	breakpoints map[int]bool
	stepping    bool
	currentLine int
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[int]bool),
		stepping:    false,
	}
}

func (d *Debugger) GetBreakpoints() map[int]bool {
	return d.breakpoints
}

func (d *Debugger) ShouldPause(line int) bool {
	d.currentLine = line
	return d.breakpoints[line] || d.stepping
}

func (d *Debugger) HandleDebuggerCommand(env *lang.Environment) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("(debug) ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		switch {
		case cmd == "step":
			d.stepping = true
			return true
		case cmd == "cont":
			d.stepping = false
			return true
		case cmd == "env":
			for k, v := range env.GetVars() {
				fmt.Printf("%s = %v\n", k, v)
			}
		case strings.HasPrefix(cmd, "break "):
			lineStr := strings.TrimPrefix(cmd, "break ")
			line, err := strconv.Atoi(lineStr)
			if err != nil {
				fmt.Println("Invalid line number")
			} else {
				d.breakpoints[line] = true
				fmt.Println("Breakpoint set at line", line)
			}
		case cmd == "exit":
			os.Exit(0)
		default:
			fmt.Println("Commands: step | cont | break <line> | env | exit")
		}
	}
}
