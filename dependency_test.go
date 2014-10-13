package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseServiceDependency_emptyString(t *testing.T) {
	_, err := ParseServiceDependency("")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "cannot specify empty service dependency"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error %q to contain %q", err.Error(), expected)
	}
}

func TestParseServiceDependency_name(t *testing.T) {
	sd, err := ParseServiceDependency("webapp")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{Name: "webapp"}
	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}

func TestParseServiceDependency_nameTag(t *testing.T) {
	sd, err := ParseServiceDependency("release.webapp")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{
		Name: "webapp",
		Tag:  "release",
	}

	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}

func TestParseServiceDependency_nameTagDataCenter(t *testing.T) {
	sd, err := ParseServiceDependency("release.webapp@nyc1")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{
		Name:       "webapp",
		Tag:        "release",
		DataCenter: "nyc1",
	}

	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}

func TestParseServiceDependency_nameTagDataCenterPort(t *testing.T) {
	sd, err := ParseServiceDependency("release.webapp@nyc1:8500")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{
		Name:       "webapp",
		Tag:        "release",
		DataCenter: "nyc1",
		Port:       8500,
	}

	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}

func TestParseServiceDependency_dataCenterOnly(t *testing.T) {
	_, err := ParseServiceDependency("@nyc1")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "invalid service dependency format"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error %q to contain %q", err.Error(), expected)
	}
}

func TestParseServiceDependency_nameAndPort(t *testing.T) {
	sd, err := ParseServiceDependency("webapp:8500")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{
		Name: "webapp",
		Port: 8500,
	}

	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}

func TestParseServiceDependency_nameAndDataCenter(t *testing.T) {
	sd, err := ParseServiceDependency("webapp@nyc1")
	if err != nil {
		t.Fatal(err)
	}

	expected := &ServiceDependency{
		Name:       "webapp",
		DataCenter: "nyc1",
	}

	if !reflect.DeepEqual(sd, expected) {
		t.Errorf("expected %#v to equal %#v", sd, expected)
	}
}
