package main

//go:generate genfor-interp-a bindings.go
//go:generate gofmt -w bindings.go

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/KernelDeimos/gottagofast/toolparse"
	"github.com/KernelDeimos/onsave/definitions"
	"github.com/KernelDeimos/onsave/watcher"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const helpText = `Usage: onsave [options] [subcommand]

Subcommands:
  gotmpl <template> <data> <output>  Render a Go template with YAML data
  help                               Show this help message

Options:
  -d        Enable debug logging
  --help    Show this help message

Without a subcommand, onsave watches for file changes and runs rules
defined in onsave.yaml in the current directory.
`

func main() {
	args := os.Args[1:]

	// Handle help flags before anything else
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help") {
		fmt.Print(helpText)
		return
	}

	// Strip -d debug flag if present
	if len(args) > 0 && args[0] == "-d" {
		logrus.SetLevel(logrus.DebugLevel)
		args = args[1:]
	}

	// Dispatch subcommands
	if len(args) > 0 {
		switch args[0] {
		case "gotmpl":
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "Usage: onsave gotmpl <template> <data> <output>")
				os.Exit(1)
			}
			err := GoTmpl(args[1], args[2], args[3])
			if err != nil {
				logrus.Fatal(err)
			}
			return
		}
	}

	w, err := watcher.NewDefault()
	if err != nil {
		logrus.Fatal(err)
	}

	confBytes, err := ioutil.ReadFile("onsave.yaml")
	if err != nil {
		logrus.Fatal(err)
	}

	var conf definitions.OnsaveConfig
	err = yaml.Unmarshal(confBytes, &conf)
	if err != nil {
		logrus.Fatal(err)
	}

	// Create instance of the interpreter
	I := interp_a.InterpreterFactoryA{}.MakeExec()
	BindFunctions(I)

	for filename, scriptLines := range conf {
		// Begin a "do" command (executes each argument)
		scriptList := []interface{}{"do"}

		// Make interpreter scope for this file (has filename)
		// TODO: use map builtin to simplify this snippet
		ii := I.MakeChild()
		iiGet := ii.MakeChild()
		ii.AddOperation("get", iiGet.OpEvaluate)
		iiGet.AddOperation("filename", func(args []interface{}) ([]interface{}, error) {
			return []interface{}{filename}, nil
		})

		// Add each command entry as an operand of "do"
		for _, line := range scriptLines {
			// Parse line (list item from YAML file) into tokens
			lineList, err := toolparse.ParseListSimple(line)
			if err != nil {
				logrus.Fatal(err)
			}
			// Add list tokens as operand of "do" command
			scriptList = append(scriptList, lineList)
		}

		logrus.Info(filename, " >> ", scriptList)

		err := w.AddRule(ii, filename, scriptList)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	w.Run()

}
