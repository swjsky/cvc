package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
	"h12.me/go-flags"
)

type HelpError struct {
	Message string
}

func (e *HelpError) Error() string { return e.Message }

func Parse(cfg interface{}) error {
	_, err := parse(cfg)
	return err
}

func ParseCommand(cfg interface{}) (*flags.Command, error) {
	parser, err := parse(cfg)
	if err != nil {
		return nil, err
	}
	if parser.Command.Active == nil {
		var buf bytes.Buffer
		parser.WriteHelp(&buf)
		return nil, &HelpError{Message: buf.String()}
	}
	return parser.Command.Active, nil
}

func parse(cfg interface{}) (*flags.Parser, error) {
	file, err := getConfigFileName()
	if err != nil {
		return nil, err
	}
	if file != "" {
		if err := ParseFile(file, cfg); err != nil {
			return nil, err
		}
	}
	parser := flags.NewParser(cfg, flags.HelpFlag|flags.PassDoubleDash|flags.IgnoreUnknown)
	if _, err := parser.Parse(); err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			return nil, &HelpError{e.Message}
		}
		return nil, err
	}
	return parser, nil
}

func MustParseCommand(cfg interface{}) *flags.Command {
	cmd, err := ParseCommand(cfg)
	if err != nil {
		if _, ok := err.(*HelpError); ok {
			fmt.Println(err)
			os.Exit(0)
		}
		log.Fatal(err)
	}
	return cmd
}

func ExecuteCommand(cfg interface{}) {
	MustParseCommand(cfg)
}

func getConfigFileName() (string, error) {
	var fileConfig struct {
		ConfigFile string `long:"config"`
	}
	if _, err := flags.NewParser(&fileConfig, flags.IgnoreUnknown).Parse(); err != nil {
		return "", err
	}
	if fileConfig.ConfigFile != "" {
		return fileConfig.ConfigFile, nil
	}
	app := path.Base(os.Args[0])
	for _, dir := range []string{
		"",
		path.Dir(os.Args[0]),
		path.Join(os.Getenv("HOME"), "."+app),
		"/etc/" + app,
	} {
		for _, file := range []string{
			"config.yaml",
			"config.yml",
			"config.json",
		} {
			fileName := path.Join(dir, file)
			if fileExists(fileName) {
				return fileName, nil
			}
		}
	}
	return "", nil
}

func ParseFile(file string, cfg interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	switch path.Ext(file) {
	case ".json":
		return json.NewDecoder(f).Decode(cfg)
	case ".yml", ".yaml":
		in, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(in, cfg)
	}
	return errors.New("unsupported config file format: " + file)
}

func fileExists(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
