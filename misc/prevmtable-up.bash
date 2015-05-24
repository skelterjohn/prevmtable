#!/bin/bash

gcloud compute firewall-rules create prvmtable-http \
  --target-tags prevmtable-http \
  --source-ranges 0.0.0.0/0 \
  --allow tcp:8080

gcloud compute project-info add-metadata \
  --metadata-from-file prevmtable=prevmtable-config.rjson || exit

gcloud compute instances create prevmtable-master \
  --image container-vm \
  --metadata-from-file google-container-manifest=prevmtable-cvm.yaml \
  --zone us-central1-f \
  --machine-type f1-micro \
  --scopes https://www.googleapis.com/auth/cloud-platform || exit
