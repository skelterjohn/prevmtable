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
	"fmt"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute_v1 "google.golang.org/api/compute/v1"
)

type VMTable struct {
	Config  Config
	compute *compute_v1.Service
	project string
}

func NewVMTable() (*VMTable, error) {
	t := &VMTable{}
	if err := t.RefreshConfig(); err != nil {
		return nil, err
	}
	tokenSource := google.ComputeTokenSource("")
	client := oauth2.NewClient(context.Background(), tokenSource)
	var err error
	if t.compute, err = compute_v1.New(client); err != nil {
		return nil, err
	}
	if t.project, err = Project(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *VMTable) RefreshConfig() error {
	var err error
	t.Config, err = ConfigFromMetadata()
	return err
}

func (t *VMTable) RefreshVMsInZone(zone string) error {
	fmt.Println("fetching VM names...")
	result, err := t.compute.Instances.List(t.project, zone).Do()
	if err != nil {
		return err
	}
	fmt.Printf("got %d results\n", len(result.Items))
	for _, item := range result.Items {
		fmt.Println(item.Name)
	}
	return nil
}
