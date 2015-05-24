#!/bin/bash

set -x


gcloud compute forwarding-rules delete -q us-central1/prevmtable-rule

gcloud compute target-pools delete -q us-central1/prevmtable-pool
gcloud compute target-pools create us-central1/prevmtable-pool || exit

gcloud compute forwarding-rules create us-central1/prevmtable-rule \
  --target-pool us-central1/prevmtable-pool || exit

gcloud compute firewall-rules create prevmtable-http \
  --target-tags prevmtable-http \
  --source-ranges 0.0.0.0/0 \
  --allow tcp:8080

gcloud compute project-info add-metadata \
  --metadata-from-file prevmtable=prevmtable-config.rjson || exit
gcloud compute project-info add-metadata \
  --metadata-from-file prevmtable-create-hook=create-hook.bash || exit

gcloud compute instances create us-central1-b/prevmtable-master \
  --image container-vm \
  --metadata-from-file google-container-manifest=prevmtable-cvm.yaml \
  --machine-type f1-micro \
  --scopes https://www.googleapis.com/auth/cloud-platform || exit
