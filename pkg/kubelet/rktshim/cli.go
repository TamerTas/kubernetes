/*
Copyright 2016 The Kubernetes Authors.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rktshim

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	utilexec "k8s.io/kubernetes/pkg/util/exec"
)

var (
	errFlagTagNotFound           = errors.New("arg: given field doesn't have a `flag` tag")
	errStructFieldNotInitialized = errors.New("arg: given field is unitialized")
)

// TODO(tmrts): refactor these into an util pkg
// Uses reflection to retrieve the `flag` tag of a field.
// The value of the `flag` field with the value of the field is
// used to construct a POSIX long flag argument string.
func getLongFlagFormOfField(fieldValue reflect.Value, fieldType reflect.StructField) (string, error) {
	flagTag := fieldType.Tag.Get("flag")
	if flagTag == "" {
		return "", errFlagTagNotFound
	}

	if fieldValue.IsValid() {
		return "", errStructFieldNotInitialized
	}

	switch fieldValue.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("--%v", flagTag), nil
	case reflect.Int:
		return fmt.Sprintf("--%v=%v", flagTag, fieldValue.Int()), nil
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		var args []string
		for i := 0; i < fieldValue.Len(); i++ {
			args = append(args, fieldValue.Index(i).String())
		}

		return fmt.Sprintf("--%v=%v", flagTag, strings.Join(args, ",")), nil
	}

	return fmt.Sprintf("--%v=%v", flagTag, fieldValue.String()), nil
}

// Uses reflection to transform a struct containing fields with `flag` tags
// to a string slice of POSIX compliant long form arguments.
func getArgumentFormOfStruct(strt interface{}) (flags []string) {
	numberOfFields := reflect.ValueOf(strt).NumField()

	for i := 0; i < numberOfFields; i++ {
		fieldValue := reflect.ValueOf(strt).Field(i)
		fieldType := reflect.TypeOf(strt).Field(i)

		flagFormOfField, err := getLongFlagFormOfField(fieldValue, fieldType)
		if err != nil {
			continue
		}

		flags = append(flags, flagFormOfField)
	}

	return
}

func getFlagFormOfStruct(strt interface{}) (flags []string) {
	return getArgumentFormOfStruct(strt)
}

type CLIConfig struct {
	Debug bool `flag:"debug"`

	Dir             string `flag:"dir"`
	LocalConfigDir  string `flag:"local-config"`
	UserConfigDir   string `flag:"user-config"`
	SystemConfigDir string `flag:"system-config"`

	InsecureOptions string `flag:"insecure-options"`
}

func (cfg *CLIConfig) Merge(newCfg CLIConfig) {
	newCfgVal := reflect.ValueOf(newCfg)

	numberOfFields := newCfgVal.NumField()

	for i := 0; i < numberOfFields; i++ {
		fieldValue := newCfgVal.Field(i)

		if !fieldValue.IsValid() {
			continue
		}

		newCfgVal.FieldByName(fieldValue.Name()).Set(fieldValue)
	}
}

type CLI interface {
	With(CLIConfig) CLI
	RunCommand(string, ...string) ([]string, error)
}

type cli struct {
	rktPath string
	config  CLIConfig
	execer  utilexec.Interface
}

func (c *cli) With(cfg CLIConfig) CLI {
	newC := NewRktCLI(c.rktPath, c.config, c.execer)

	newC.config.Merge(cfg)

	return newC
}

func (c *cli) RunCommand(subcmd string, args ...string) ([]string, error) {
	globalFlags := GetFlagFormOfStruct(cmd.config)

	args := append(globalFlags, args...)

	cmd := cmd.execer.Command(c.rktPath, append([]string{subcmd}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	//glog.V(4).Infof("rkt: Run command: %q with args: %#v", subcmd, args)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run %v: %v\nstdout: %v\nstderr: %v", args, err, stdout.String(), stderr.String())
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}

func NewRktCLI(rktPath string, cfg Config, exec utilexec.Interface) CLI {
	return &cli{rktPath: rktPath, config: cfg, execer: exec}
}
