#! /bin/bash

# Copyright 2015 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "compiling the binary..."
wgo install prevmtable || exit

echo "building the container..."
docker build -t deploy . &> /dev/null || exit

project=$(gcloud config list core/project --format=text | cut -d ' ' -f 2)

echo "running the faux metadata container..."
metadata=$(mktemp)
cat > $metadata << EOM
computeMetadata:
  v1: &V1
    project:
      projectId: &PROJECT-ID
        $project
      numericProjectId: 1234
      attributes:
        prevmtable: |
          {
            secondsToRest: 30
            secondsForExhaustion: 120
            prefix: "delete-"
            allowedzones: [
              "us-central1-b"
            ]
            targetVMCount: 1
            instance: {
              metadata: {
                items: [
                  {
                    key: "startup-script"
                    value: "docker run --rm -p 8080:8080 skelterjohn/http"
                  }
                ]
              }
              tags: {
                items: [
                  "prevmtable-http"
                ]
              }
              machineType: "https://www.googleapis.com/compute/v1/projects/{project}/zones/{zone}/machineTypes/f1-micro"
              name: "{name}"
              disks: [
                {
                  autoDelete: true
                  boot: true
                  initializeParams: {
                    sourceImage: "https://www.googleapis.com/compute/v1/projects/coreos-cloud/global/images/coreos-stable-647-0-0-v20150512"
                  }
                  mode: "READ_WRITE"
                  type: "PERSISTENT"
                }
              ]
              networkInterfaces: [
                {
                  accessConfigs: [
                    {
                      name: "external-nat"
                      type: "ONE_TO_ONE_NAT"
                    }
                  ]
                  network: "https://www.googleapis.com/compute/v1/projects/{project}/global/networks/default"
                }
              ]
              scheduling: {
                automaticRestart: false
                preemptible: true
              }
            }
          }
    instance:
      projectId: *PROJECT-ID
      hostname: deploy_machine
      machineType: n1-standard-1
      maintenanceEvent: NONE
      serviceAccounts:
        default: *DEFAULT
        prevmtable@googleserviceaccount.com: &DEFAULT
          email: prevmtable@googleserviceaccount.com
          scopes:
            - https://www.googleapis.com/auth/cloud-platform
      zone: us-central1-a
EOM

metadata_id=$(docker run \
 -d \
 --name metadata \
 -v $metadata:/prevmtable/manifest.yaml \
 gcr.io/_b_containers_qa/faux-metadata:latest \
  -manifest_file=/prevmtable/manifest.yaml \
  -refresh_token=$(gcloud auth print-refresh-token))

docker run \
 --rm \
 --link metadata:metadata.google.internal \
 --env GCE_METADATA_HOST=metadata.google.internal \
 --env PREVMTABLE_ATTRIBUTE=prevmtable \
 deploy

docker rm -f metadata &> /dev/null
