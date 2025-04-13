package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func main() {
	file, err := os.Open("example.toy")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	m := ir.NewModule()

	m.TargetTriple = "x86_64-unknown-linux-gnu"
	m.SourceFilename = "example.toy"

	// Declare printf
	printf := m.NewFunc("printf", types.I32, ir.NewParam("", types.NewPointer(types.I8)))
	printf.Sig.Variadic = true

	mainFn := m.NewFunc("main", types.I32)
	block := mainFn.NewBlock("entry")

	vars := map[string]*ir.InstAlloca{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "let ") {
			// Parse: let x = a [+ b];
			tokens := strings.Split(strings.TrimSuffix(line[4:], ";"), "=")
			varName := strings.TrimSpace(tokens[0])
			expr := strings.TrimSpace(tokens[1])

			// Handle expressions like "5", "x + 3"
			parts := strings.Split(expr, "+")
			var lhs value.Value

			part0 := strings.TrimSpace(parts[0])
			if val, err := strconv.Atoi(part0); err == nil {
				lhs = constant.NewInt(types.I32, int64(val))
			} else {
				lhs = block.NewLoad(types.I32, vars[part0])
			}

			var result value.Value = lhs
			if len(parts) > 1 {
				part1 := strings.TrimSpace(parts[1])
				var rhs value.Value
				if val, err := strconv.Atoi(part1); err == nil {
					rhs = constant.NewInt(types.I32, int64(val))
				} else {
					rhs = block.NewLoad(types.I32, vars[part1])
				}
				result = block.NewAdd(lhs, rhs)
			}

			ptr := block.NewAlloca(types.I32)
			block.NewStore(result, ptr)
			vars[varName] = ptr

		} else if strings.HasPrefix(line, "print(") {
			// Extract content inside print(...)
			arg := strings.TrimSuffix(strings.TrimPrefix(line, "print("), ");")
			arg = strings.TrimSpace(arg)

			if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
				// Handle string literal
				strContent := strings.Trim(arg, "\"") + "\n\x00"
				globalName := fmt.Sprintf("str_%d", len(m.Globals))
				strGlobal := m.NewGlobalDef(globalName, constant.NewCharArrayFromString(strContent))
				strGlobal.Immutable = true

				strPtr := block.NewGetElementPtr(
					strGlobal.ContentType,
					strGlobal,
					constant.NewInt(types.I32, 0),
					constant.NewInt(types.I32, 0),
				)

				block.NewCall(printf, strPtr)
			} else {
				// Handle variable print
				val := block.NewLoad(types.I32, vars[arg])

				// Create %d\n format string
				fmtStr := "%d\n\x00"
				globalName := fmt.Sprintf("fmt_%d", len(m.Globals))
				fmtGlobal := m.NewGlobalDef(globalName, constant.NewCharArrayFromString(fmtStr))
				fmtGlobal.Immutable = true

				fmtPtr := block.NewGetElementPtr(
					fmtGlobal.ContentType,
					fmtGlobal,
					constant.NewInt(types.I32, 0),
					constant.NewInt(types.I32, 0),
				)

				block.NewCall(printf, fmtPtr, val)
			}
		}
	}

	block.NewRet(constant.NewInt(types.I32, 0))

	err = os.WriteFile("./bin/output.ll", []byte(m.String()), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Compiled LLVM IR written to output.ll")
}
