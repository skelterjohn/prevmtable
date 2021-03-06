// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vmtable

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/rogpeppe/rjson"
	compute_v1 "google.golang.org/api/compute/v1"
	"google.golang.org/cloud/compute/metadata"
)

var (
	NotPreemtibleError = errors.New("instance template in config is not preemptible")
)

type Config struct {
	// Seconds between updates.
	SecondsToRest int

	// Seconds to wait before retrying a zone that got exhausted.
	SecondsForExhaustion int

	// Prefix to put on the name of each VM.
	Prefix string

	// The zones to create VMs in.
	AllowedZones []string

	// Project metadata attributes containing script hooks.
	Hooks struct {
		Create    string
		Delete    string
		Vanished  string
		Exhausted string
	}

	// Number of VMs to maintain. If there are more, delete. If there are fewer, create.
	TargetVMCount int

	// Template to use for instance creation.
	Instance rjson.RawMessage
}

func ConfigFromMetadata() (Config, error) {
	attrName := os.Getenv("PREVMTABLE_ATTRIBUTE")
	if attrName == "" {
		attrName = "prevmtable"
	}

	cfgData, err := metadata.ProjectAttributeValue(attrName)
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	if err := rjson.NewDecoder(strings.NewReader(cfgData)).Decode(&cfg); err != nil {
		return Config{}, err
	}

	i := &compute_v1.Instance{}
	instanceData := string(cfg.Instance)
	instanceData = strings.Replace(instanceData, "{project}", "proj", -1)
	instanceData = strings.Replace(instanceData, "{zone}", "zone", -1)
	instanceData = strings.Replace(instanceData, "{name}", "name", -1)

	if err := rjson.Unmarshal([]byte(instanceData), i); err != nil {
		return cfg, err
	}

	if !i.Scheduling.Preemptible {
		return cfg, NotPreemtibleError
	}

	return cfg, nil
}

func (c Config) CreateHook(project, zone, name string) error {
	return c.execHook(
		"create",
		c.Hooks.Create,
		[]string{
			"PROJECT=" + project,
			"ZONE=" + zone,
			"NAME=" + name,
		})
}

func (c Config) DeleteHook(project, zone, name string) error {
	return c.execHook(
		"delete",
		c.Hooks.Delete,
		[]string{
			"PROJECT=" + project,
			"ZONE=" + zone,
			"NAME=" + name,
		})
}

func (c Config) VanishedHook(project, zone, name string) error {
	return c.execHook(
		"vanished",
		c.Hooks.Vanished,
		[]string{
			"PROJECT=" + project,
			"ZONE=" + zone,
			"NAME=" + name,
		})
}

func (c Config) ExhaustedHook(project, zone string) error {
	return c.execHook(
		"exhausted",
		c.Hooks.Exhausted,
		[]string{
			"PROJECT=" + project,
			"ZONE=" + zone,
		})
}

func (c Config) execHook(hookType, scriptAttribute string, env []string) error {
	if scriptAttribute == "" {
		return nil
	}
	script, err := metadata.ProjectAttributeValue(scriptAttribute)
	if err != nil {
		return err
	}

	scriptFile, err := ioutil.TempFile("", "hook-")
	if err != nil {
		return err
	}
	scriptPath := scriptFile.Name()
	if _, err := scriptFile.WriteString(script); err != nil {
		return err
	}
	scriptFile.Close()
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return err
	}
	log.Printf("executing %s hook with %q: %s", hookType, scriptPath, env)
	cmd := exec.Command(scriptPath)
	cmd.Env = append(env, os.Environ()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Project() (string, error) {
	return metadata.ProjectID()
}
