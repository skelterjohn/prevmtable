#! /bin/bash

echo "compiling the binary..."
wgo install prevmtable || exit

echo "building the container..."
docker build -t deploy -f Dockerfile.deploy . &> /dev/null || exit

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
            secondsToRest: 5
            prefix: "delete-"
            allowedzones: [
              "us-central1-b"
            ]
            target: 1
            instance: {
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
              machineType: "https://www.googleapis.com/compute/v1/projects/{project}/zones/{zone}/machineTypes/f1-micro"
              name: "{name}"
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
              serviceAccounts: [
                {
                  email: "default"
                  scopes: [
                    "https://www.googleapis.com/auth/computeaccounts.readonly"
                    "https://www.googleapis.com/auth/devstorage.read_only"
                    "https://www.googleapis.com/auth/logging.write"
                  ]
                }
              ]
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
 deploy

docker rm -f metadata &> /dev/null
