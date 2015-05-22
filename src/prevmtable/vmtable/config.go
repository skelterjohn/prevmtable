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
	"strings"

	"github.com/rogpeppe/rjson"
	"google.golang.org/cloud/compute/metadata"
)

type Config struct {
	SecondsToRest float64
	Prefix        string
	AllowedZones  []string
	MachineType   string
	GCEImage      string

	Target int

	Instance rjson.RawMessage
}

func ConfigFromMetadata() (Config, error) {
	cfgData, err := metadata.ProjectAttributeValue("prevmtable")
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	if err := rjson.NewDecoder(strings.NewReader(cfgData)).Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Project() (string, error) {
	return metadata.ProjectID()
}
