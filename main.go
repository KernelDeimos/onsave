package main

//go:generate genfor-interp-a bindings.go
//go:generate gofmt -w bindings.go

import (
	"io/ioutil"
	"os"

	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/KernelDeimos/gottagofast/toolparse"
	"github.com/KernelDeimos/onsave/definitions"
	"github.com/KernelDeimos/onsave/watcher"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	args := os.Args[:1]
	if len(args) > 0 && args[0] == "-d" {
		logrus.SetLevel(logrus.DebugLevel)
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
