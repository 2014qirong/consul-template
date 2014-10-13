package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

/// ------------------------- ///

// Exit codes are int valuse that represent an exit code for a particular error.
// Sub-systems may check this unique error to determine the cause of an error
// without parsing the output or help text.
const (
	ExitCodeOK int = 0

	// Errors start at 500
	ExitCodeError = 500 + iota
	ExitCodeParseFlagsError
	ExitCodeParseWaitError
	ExitCodeParseConfigError
)

/// ------------------------- ///

type CLI struct {
	// outSteam and errStream are the standard out and standard error streams to
	// write messages from the CLI.
	outStream, errStream io.Writer
}

// Run accepts a list of arguments and returns an int representing the exit
// status from the command.
func (c *CLI) Run(args []string) int {
	var dry, version bool
	config := &Config{}

	cmd := filepath.Base(args[0])

	flags := flag.NewFlagSet("consul-template", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintf(c.errStream, helpText, cmd)
	}
	flags.StringVar(&config.Consul, "consul", "127.0.0.1:8500",
		"address of the Consul instance")
	flags.Var((*configTemplateVar)(&config.ConfigTemplates), "template",
		"new template declaration")
	flags.StringVar(&config.Token, "token", "abcd1234",
		"a consul API token")
	flags.StringVar(&config.WaitRaw, "wait", "",
		"the minimum(:maximum) to wait before rendering a new template")
	flags.StringVar(&config.Path, "config", "",
		"the path to a config file on disk")
	flags.BoolVar(&config.Once, "once", false,
		"do not run as a daemon")
	flags.BoolVar(&dry, "dry", false,
		"write generated templates to stdout")
	flags.BoolVar(&version, "version", false, "display the version")

	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(c.errStream, "%s\n", err)
		flags.Usage()
		return ExitCodeParseError
	}

	// If the version was requested, print and exit
	if version {
		fmt.Fprintf(c.errStream, "%s v%s\n", cmd, Version)
		return ExitCodeOK
	}

	// Parse the raw wait value into a Wait object
	if config.WaitRaw != "" {
		wait, err := ParseWait(config.WaitRaw)
		if err != nil {
			fmt.Fprintf(c.errStream, "%s\n", err)
			return ExitCodeParseWaitError
		}
		config.Wait = wait
	}

	// Merge a path config with the command line options. Command line options
	// take precedence over config file options for easy overriding.
	if config.Path != "" {
		fileConfig, err := ParseConfig(config.Path)
		if err != nil {
			fmt.Fprintf(c.errStream, "%s\n", err)
			return ExitCodeParseConfigError
		}
		fileConfig.Merge(config)
		config = fileConfig
	}

	return ExitCodeOK
}

const helpText = `
Usage: %s [options]

  Watches a series of templates on the file system, writing new changes when
  Consul is updated. It runs until an interrupt is received unless the -once
  flag is specified.

Options:

  -consul=<address>        Sets the address of the Consul instance
  -token=<token>           Sets the Consul API token
  -template=<template>      Adds a new template to watch on disk in the format
                           'templatePath:outputPath(:command)'.
  -wait=<duration>         Sets the 'minumum(:maximum)' amount of time to wait
                           before writing a template (and triggering a command)
  -config=<path>           Sets the path to a configuration file on disk

  -dry                     Dump generated templates to stdout
  -once                    Do not run the process as a daemon
  -version                 Print the version of this daemon
`

/// ------------------------- ///

// configTemplateVar implements the Flag.Value interface and allows the user
// to specify multiple -template keys in the CLI where each option is parsed
// as a template.
type configTemplateVar []*ConfigTemplate

func (ctv configTemplateVar) String() string {
	buff := new(bytes.Buffer)
	for _, template := range ctv {
		fmt.Fprintf(buff, "%s", template.Source)
		if template.Destination != "" {
			fmt.Fprintf(buff, ":%s", template.Destination)

			if template.Command != "" {
				fmt.Fprintf(buff, ":%s", template.Command)
			}
		}
	}

	return buff.String()
}

func (ctv *configTemplateVar) Set(value string) error {
	template, err := ParseConfigTemplate(value)
	if err != nil {
		return err
	}

	if *ctv == nil {
		*ctv = make([]*ConfigTemplate, 0, 1)
	}
	*ctv = append(*ctv, template)

	return nil
}
