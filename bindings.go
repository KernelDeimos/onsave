package main

import (
	"errors"

	"github.com/KernelDeimos/anything-gos/interp_a"
)

func BindFunctions(ii interp_a.HybridEvaluator) {

	//::gen gen-binding ii gotmpl GoTmpl (strings templateFile string dataFile string outputFile string) (strings err error)
	ii.AddOperation("gotmpl", func(
		args []interface{}) ([]interface{}, error) {

		if len(args) < 3 {
			return nil, errors.New("gotmpl requires at least 3 arguments")
		}

		var templateFile string
		var dataFile string
		var outputFile string
		{
			var ok bool
			templateFile, ok = args[0].(string)
			if !ok {
				return nil, errors.New("gotmpl: argument 0: templateFile; must be type string")
			}
			dataFile, ok = args[1].(string)
			if !ok {
				return nil, errors.New("gotmpl: argument 1: dataFile; must be type string")
			}
			outputFile, ok = args[2].(string)
			if !ok {
				return nil, errors.New("gotmpl: argument 2: outputFile; must be type string")
			}
		}
		err := GoTmpl(templateFile, dataFile, outputFile)
		if err != nil {
			return nil, err
		}
		return []interface{}{}, nil
	})
	//::end

	//::gen gen-binding ii exec Exec (strings cmd string argsI ...interface{}) (strings err error)
	ii.AddOperation("exec", func(
		args []interface{}) ([]interface{}, error) {

		if len(args) < 1 {
			return nil, errors.New("exec requires at least 1 arguments")
		}

		var cmd string
		{
			var ok bool
			cmd, ok = args[0].(string)
			if !ok {
				return nil, errors.New("exec: argument 0: cmd; must be type string")
			}
		}
		argsI := args[1:]
		err := Exec(cmd, argsI...)
		if err != nil {
			return nil, err
		}
		return []interface{}{}, nil
	})
	//::end

}
