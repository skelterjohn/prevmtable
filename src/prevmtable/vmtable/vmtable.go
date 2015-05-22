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
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/rogpeppe/rjson"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute_v1 "google.golang.org/api/compute/v1"
)

type VMTable struct {
	Config  Config
	compute *compute_v1.Service
	project string

	ZoneInstances map[string][]*compute_v1.Instance
}

func NewVMTable() (*VMTable, error) {
	t := &VMTable{
		ZoneInstances: map[string][]*compute_v1.Instance{},
	}
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
	if err != nil {
		log.Printf("error refreshing config: %v", err)
	}
	return err
}

func (t *VMTable) RefreshVMs() {
	t.ZoneInstances = map[string][]*compute_v1.Instance{}
	for _, zone := range t.Config.AllowedZones {
		if err := t.RefreshVMsInZone(zone); err != nil {
			log.Printf("error refreshing zone %s: %v", zone, err)
		}
	}
}

func (t *VMTable) RefreshVMsInZone(zone string) error {
	result, err := t.compute.Instances.List(t.project, zone).Do()
	if err != nil {
		return err
	}
	for _, item := range result.Items {
		// skip instances we don't know about
		if !strings.HasPrefix(item.Name, t.Config.Prefix) {
			continue
		}

		// skip instances that aren't running or won't be running
		if item.Status == "STOPPING" || item.Status == "TERMINATED" {
			continue
		}
		zis, _ := t.ZoneInstances[zone]
		zis = append(zis, item)
		t.ZoneInstances[zone] = zis
	}
	return nil
}

func (t *VMTable) RightSize() {
	activeVMs := 0
	for _, zis := range t.ZoneInstances {
		activeVMs += len(zis)
	}

	if activeVMs == t.Config.Target {
		return
	}

	if activeVMs < t.Config.Target {
		t.createVMs(t.Config.Target - activeVMs)
	} else {
		t.deleteVMs(activeVMs - t.Config.Target)
	}
}

func (t *VMTable) createVMs(count int) {
	var wg sync.WaitGroup
	wg.Add(count)
	offset := rand.Int()
	for i := 0; i < count; i++ {
		zone := t.Config.AllowedZones[(offset+i)%len(t.Config.AllowedZones)]
		go func(zone string) {
			t.createVM(zone)
			wg.Done()
		}(zone)
	}
	wg.Wait()
}

func (t *VMTable) deleteVMs(count int) {
	type nz struct {
		name string
		zone string
	}

	var allInstances []nz
	for z, zis := range t.ZoneInstances {
		for _, i := range zis {
			allInstances = append(allInstances, nz{i.Name, z})
		}
	}
	var wg sync.WaitGroup
	wg.Add(count)
	for count > 0 {
		i := rand.Intn(len(allInstances))
		go func(i nz) {
			t.deleteVM(i.zone, i.name)
			wg.Done()
		}(allInstances[i])
		allInstances = append(allInstances[:i], allInstances[i+1:]...)
		count--
	}

	wg.Wait()
}

func (t *VMTable) createVM(zone string) {
	id := uuid.NewV4()
	name := fmt.Sprintf("%s%s", t.Config.Prefix, id)

	i := &compute_v1.Instance{}
	instanceData := string(t.Config.Instance)
	instanceData = strings.Replace(instanceData, "{project}", t.project, -1)
	instanceData = strings.Replace(instanceData, "{zone}", zone, -1)
	instanceData = strings.Replace(instanceData, "{name}", name, -1)

	if err := rjson.Unmarshal([]byte(instanceData), i); err != nil {
		log.Printf("error decoding instance template: %s", err)
		return
	}

	log.Printf("inserting instance %s/%s/%s", t.project, zone, i.Name)
	op, err := t.compute.Instances.Insert(t.project, zone, i).Do()
	if err != nil {
		log.Printf("error inserting instance: %s", err)
	}

	for range time.Tick(2 * time.Second) {
		op, err := t.compute.ZoneOperations.Get(t.project, zone, op.Name).Do()
		if err != nil {
			log.Printf("error fetching operation: %v", err)
			return
		}
		if op.Status == "RUNNING" {
			break
		}
		if op.Status != "PENDING" {
			log.Printf("unexpected operation status: %s/%s/%s = %s", t.project, zone, op.Name, op.Status)
			break
		}
	}
}

func (t *VMTable) deleteVM(zone, name string) {
	log.Printf("deleting instance %s/%s/%s", t.project, zone, name)
	op, err := t.compute.Instances.Delete(t.project, zone, name).Do()
	if err != nil {
		log.Printf("error deleting instance: %s", err)
	}
	for range time.Tick(2 * time.Second) {
		op, err := t.compute.ZoneOperations.Get(t.project, zone, op.Name).Do()
		if err != nil {
			log.Printf("error fetching operation: %v", err)
			return
		}
		if op.Status == "RUNNING" {
			break
		}
		if op.Status != "PENDING" {
			log.Printf("unexpected operation status: %s/%s/%s = %s", t.project, zone, op.Name, op.Status)
			break
		}
	}
}
