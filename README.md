#prevmtable#

Because not having enough VMs is preventable.

Prevmtable manages a pool of preemptible VMs across zones. The VMs are created from a provided template, and prevmtable will balance them across zones.

If a zone's preemtible machines are exhausted, prevmtable will load all the VMs up in the other zones.

##configuration##

Configuration of the pool is managed through GCE metadata. The project metadata attribute "prevmtable" (overridable with the environment variable PREVMTABLE_ATTRIBUTE) must be an rjson (https://github.com/rogpeppe/rjson) document with that matches the following structure.

    type Config struct {
      // Seconds between updates.
      SecondsToRest int

      // Seconds to wait before retrying a zone that got exhausted.
      SecondsForExhaustion int

      // Prefix to put on the name of each VM.
      Prefix string

      // The zones to create VMs in.
      AllowedZones []string

      // Number of VMs to maintain. If there are more, delete. If there are fewer, create.
      TargetVMCount int

      // Template to use for instance creation.
      Instance rjson.RawMessage
    }

The `Instance` component will be decoded into an instance of type http://godoc.org/google.golang.org/api/compute/v1#Instance.

If the strings "{project}", "{zone}", and "{name}" are somewhere in the instance template, they will be replaced with the appropriate project, zone, and name during individual instance creation. See the "example config" section below for a starting point.

##building##

Revision pinning is done with https://github.com/skelterjohn/wgo. Run `wgo restore` in the cloned github repo to fetch dependencies, and `wgo install prevmtable` to build.

##docker integration##

After building the binary, the Dockerfile can be used to create a container that will run prevmtable. 

##running##

Running either the binary or the container in context with GCE metadata and metadata-provided credentials will allow prevmtable to manage a VM pool.

The `run_deploy.bash` script demonstrates a way to have GCE metadata context without running from GCE, using a false metadata container that is linked with the prevmtable container. But, for something reliable, you'd probably want kubernetes (or something) to keep the prevmtable container going on a GCE VM.

##example config##

The config below will keep one preemtible f1-micro coreos instance running in either us-central1-b or us-central1-f.

    {
      secondsToRest: 5
      secondsForExhaustion: 120
      prefix: "delete-"
      allowedzones: [
        "us-central1-b"
        "us-central1-f"
      ]
      targetVMCount: 1
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