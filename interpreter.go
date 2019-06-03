package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/KernelDeimos/onsave/shorts"
	yaml "gopkg.in/yaml.v2"
)

func GoTmpl(templateFile, dataFile, outputFile string) error {
	tmplBytes, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	dataBytes, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}

	var dataObject interface{}

	ext := filepath.Ext(dataFile)
	switch ext {
	case ".yaml":
		fallthrough
	case ".yml":
		err = yaml.Unmarshal(dataBytes, &dataObject)
	default:
		return errors.New("unrecognized filetype: " + ext)
	}

	if err != nil {
		return err
	}

	output := shorts.RenderHTML(
		string(tmplBytes), dataObject,
	)

	ioutil.WriteFile(outputFile, []byte(output), 0664)
	return nil
}

func Exec(cmd string, args ...interface{}) error {
	aStrings := []string{}
	for _, arg := range args {
		aStrings = append(aStrings, fmt.Sprint(arg))
	}

	result, err := exec.Command(cmd, aStrings...).Output()
	if err != nil {
		return err
	}
	os.Stdout.Write(result)
	return nil
}
