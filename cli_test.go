package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/consul-template/test"
)

func TestRun_printsErrors(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)
	args := strings.Split("consul-template -bacon delicious", " ")

	status := cli.Run(args)
	if status == ExitCodeOK {
		t.Fatal("expected not OK exit code")
	}

	expected := "flag provided but not defined: -bacon"
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("expected %q to eq %q", errStream.String(), expected)
	}
}

func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)
	args := strings.Split("consul-template -version", " ")

	status := cli.Run(args)
	if status != ExitCodeOK {
		t.Errorf("expected %q to eq %q", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("consul-template v%s", Version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("expected %q to eq %q", errStream.String(), expected)
	}
}

func TestRun_parseError(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)
	args := strings.Split("consul-template -bacon delicious", " ")

	status := cli.Run(args)
	if status != ExitCodeParseFlagsError {
		t.Errorf("expected %q to eq %q", status, ExitCodeParseFlagsError)
	}

	expected := "flag provided but not defined: -bacon"
	if !strings.Contains(errStream.String(), expected) {
		t.Fatalf("expected %q to contain %q", errStream.String(), expected)
	}
}

func TestRun_waitFlagError(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)
	args := strings.Split("consul-template -wait=watermelon:bacon", " ")

	status := cli.Run(args)
	if status != ExitCodeParseWaitError {
		t.Errorf("expected %q to eq %q", status, ExitCodeParseWaitError)
	}

	expected := "time: invalid duration watermelon"
	if !strings.Contains(errStream.String(), expected) {
		t.Fatalf("expected %q to contain %q", errStream.String(), expected)
	}
}

func TestRun_onceFlag(t *testing.T) {
	template := test.CreateTempfile([]byte(`
	{{range service "consul"}}{{.Name}}{{end}}
  `), t)
	defer test.DeleteTempfile(template, t)

	out := test.CreateTempfile(nil, t)
	defer test.DeleteTempfile(out, t)

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)

	command := fmt.Sprintf("consul-template -consul demo.consul.io -template %s:%s -once", template.Name(), out.Name())
	args := strings.Split(command, " ")

	ch := make(chan int, 1)
	go func() {
		ch <- cli.Run(args)
	}()

	select {
	case status := <-ch:
		if status != ExitCodeOK {
			t.Errorf("expected %d to eq %d", status, ExitCodeOK)
			t.Errorf("stderr: %s", errStream.String())
		}
	case <-time.After(2 * time.Second):
		t.Errorf("expected exit, did not exit after 2 seconds")
	}
}

func TestReload_sighup(t *testing.T) {
	template := test.CreateTempfile([]byte("initial value"), t)
	defer test.DeleteTempfile(template, t)

	out := test.CreateTempfile(nil, t)
	defer test.DeleteTempfile(out, t)

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := NewCLI(outStream, errStream)

	command := fmt.Sprintf("consul-template -template %s:%s", template.Name(), out.Name())
	args := strings.Split(command, " ")

	go func(args []string) {
		if exit := cli.Run(args); exit != 0 {
			t.Fatal("bad exit code: %d", exit)
		}
	}(args)
	defer cli.stop()

	// Ensure we have run at least once
	test.WaitForFileContents(out.Name(), []byte("initial value"), t)

	newValue := []byte("new value")
	ioutil.WriteFile(template.Name(), newValue, 0644)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	test.WaitForFileContents(out.Name(), []byte("new value"), t)
}
