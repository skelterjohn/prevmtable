#!/bin/bash

gcloud compute target-pools add-instances us-central1/prevmtable-pool \
	--instances /$PROJECT/$ZONE/$NAME
