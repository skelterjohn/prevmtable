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

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"prevmtable/vmtable"
)

func orExit(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	t, err := vmtable.NewVMTable()
	orExit(err)

	if t.Config.SecondsToRest == 0 {
		t.Config.SecondsToRest = 5
	}

	ticker := time.Tick(time.Duration(int(time.Second) * t.Config.SecondsToRest))
	for {
		select {
		case <-ticker:
			fmt.Fprintf(os.Stderr, "\r[%v] ", time.Now().Format("2006-01-02 15:04:05 -0700"))
			if err := t.RefreshConfig(); err != nil {
				log.Printf("error refreshing config: %v", err)
				continue
			}
			t.RefreshVMs()
			t.RightSize()
		}
	}
}
