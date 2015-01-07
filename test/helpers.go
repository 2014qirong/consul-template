package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

type FakeDependencyFetchError struct {
	Name string
}

func (d *FakeDependencyFetchError) Fetch(client *api.Client, options *api.QueryOptions) (interface{}, *api.QueryMeta, error) {
	time.Sleep(50 * time.Millisecond)
	return nil, nil, fmt.Errorf("failed to contact server")
}

func (d *FakeDependencyFetchError) HashCode() string {
	return fmt.Sprintf("FakeDependencyFetchError|%s", d.Name)
}

func (d *FakeDependencyFetchError) Key() string {
	return d.Name
}

func (d *FakeDependencyFetchError) Display() string {
	return "fakedep"
}

type FakeDependencyFetchRetry struct {
	Name    string
	retried bool
}

func (d *FakeDependencyFetchRetry) Fetch(client *api.Client, options *api.QueryOptions) (interface{}, *api.QueryMeta, error) {
	time.Sleep(50 * time.Millisecond)

	if d.retried {
		data := "this is some data"
		qm := &api.QueryMeta{LastIndex: 1}
		return data, qm, nil
	} else {
		d.retried = true
		return nil, nil, fmt.Errorf("failed to contact server (try again)")
	}
}

func (d *FakeDependencyFetchRetry) HashCode() string {
	return fmt.Sprintf("FakeDependencyFetchRetry|%s", d.Name)
}

func (d *FakeDependencyFetchRetry) Key() string {
	return d.Name
}

func (d *FakeDependencyFetchRetry) Display() string {
	return "fakedep"
}

type FakeDependency struct {
	Name string
}

func (d *FakeDependency) Fetch(client *api.Client, options *api.QueryOptions) (interface{}, *api.QueryMeta, error) {
	time.Sleep(50 * time.Millisecond)
	data := "this is some data"
	qm := &api.QueryMeta{LastIndex: 1}
	return data, qm, nil
}

func (d *FakeDependency) HashCode() string {
	return fmt.Sprintf("FakeDependency|%s", d.Name)
}

func (d *FakeDependency) Key() string {
	return d.Name
}

func (d *FakeDependency) Display() string {
	return "fakedep"
}

func DemoConsulClient(t *testing.T) (*api.Client, *api.QueryOptions) {
	config := api.DefaultConfig()
	config.Address = "demo.consul.io"

	client, err := api.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Agent().NodeName(); err != nil {
		t.Fatal(err)
	}

	options := &api.QueryOptions{WaitTime: 10 * time.Second}

	return client, options
}

func CreateTempfile(b []byte, t *testing.T) *os.File {
	f, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	if len(b) > 0 {
		_, err = f.Write(b)
		if err != nil {
			t.Fatal(err)
		}
	}

	return f
}

func DeleteTempfile(f *os.File, t *testing.T) {
	if err := os.Remove(f.Name()); err != nil {
		t.Fatal(err)
	}
}
