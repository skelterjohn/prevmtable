#! /bin/bash

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
 -v $metadata:/argo/manifest.yaml \
 gcr.io/_b_containers_qa/faux-metadata:latest \
  -manifest_file=/argo/manifest.yaml \
  -refresh_token=$(gcloud auth print-refresh-token))

docker run \
 --rm \
 --link metadata:metadata.google.internal \
 --env GCE_METADATA_HOST=metadata.google.internal \
 deploy

docker rm -f metadata
