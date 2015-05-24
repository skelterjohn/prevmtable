#prevmtable#

Because not having enough VMs is preventable.

Prevmtable manages a pool of preemptible Google Compute Engine (GCE) VMs across zones. The VMs are created from a provided template, and prevmtable will balance them (round-robin style) across zones.

If a zone's preemtible machines are exhausted, prevmtable will leave that zone alone for a set amount of time, and create new VMs in the remaining zones in the meantime.

Prevmtable is effectively stateless. If it goes down, you can bring it up and it will continue on as before. The only loss is that VMs killed while prevmtable was not running will not be reported.

##configuration##

Configuration of the pool is managed through GCE project metadata. The project metadata can be updated at any time to dynamically change prevmtable's configuration - it will be checked again during the next poll cycle.

The project metadata attribute "prevmtable" (overridable with the environment variable PREVMTABLE_ATTRIBUTE) must be an rjson (https://github.com/rogpeppe/rjson) document that matches the following structure.

    type Config struct {
      // Seconds between updates.
      SecondsToRest int

      // Seconds to wait before retrying a zone that got exhausted.
      SecondsForExhaustion int

      // Prefix to put on the name of each VM.
      Prefix string

      // The zones in which VMs may be created.
      AllowedZones []string

      // Project metadata attributes containing script hooks.
      Hooks struct {
        Create    string
        Delete    string
        Vanished  string
        Exhausted string
      }

      // Number of VMs to maintain. If there are more, delete. If there are fewer, create.
      TargetVMCount int

      // Template to use for instance creation.
      Instance rjson.RawMessage
    }

If the Hooks are given values, prevmtable will look in project metadata attributes for scripts to run.

The `Instance` component will be decoded into an instance of type http://godoc.org/google.golang.org/api/compute/v1#Instance.

If the strings "{project}", "{zone}", and "{name}" are somewhere in the instance template, they will be replaced with the appropriate project, zone, and name during individual instance creation. See the "example config" section below for a starting point.

###hooks###

Prevmtable currently has four hooks, for instace creation, deletion, loss, and for zone exhaustion. Put a script (don't forget, eg, the "#!/bin/bash" at the top) in a project metadata attribute pointed to by the hooks in the config.

The script hook will be downloaded and run each time the hook fires, so the project metadata can be safely changed during prevmtable operation.

####Create####

$PROJECT: The GCP project.
$ZONE: The GCE zone.
$NAME: The GCE instance name.

The create hook is called whenever prevmtable creates a new instance.

####Delete####

$PROJECT: The GCP project.
$ZONE: The GCE zone.
$NAME: The GCE instance name.

The delete hook is called whenever prevmtable deletes an old instance.

####Vanished####

$PROJECT: The GCP project.
$ZONE: The GCE zone.
$NAME: The GCE instance name.

The vanished hook is called whenever prevmtable notices that an instance disappeared from one update to the next, and it was not deleted by prevmtable.

####Exhausted####

$PROJECT: The GCP project.
$ZONE: The GCE zone.

The exhausted hook is called whenever prevmtable tries to create an instance in a zone, but the operation fails with ZONE_RESOURCE_POOL_EXHAUSTED.

##building##

Revision pinning is done with https://github.com/skelterjohn/wgo. Run `wgo restore` in the cloned github repo to fetch dependencies, and `wgo install prevmtable` to build.

Or, set GOPATH to be the root of this repository, and test your luck with `go get prevmtable`. Maybe it will work?

###docker integration###

After building the binary for linux 64bit (GOOS=linux, GOARCH=amd64, rebuild go, rebuild the binary), the Dockerfile can be used to create a container that will run prevmtable. 

##running##

Running either the binary or the container in context with GCE metadata and metadata-provided credentials will allow prevmtable to manage a VM pool.

The `run_deploy.bash` script demonstrates a way to have GCE metadata context without running from GCE, using a false metadata container that is linked with the prevmtable container. But, for something reliable, you'd probably want kubernetes (or something) to keep the prevmtable container going on a GCE VM.

###example config###

The config below will keep one preemtible f1-micro coreos instance running in either us-central1-b or us-central1-f. Additionally, it has a startup script that runs a very simple "Hello, world!" http server, and a tag that can be used to manage its firewall status.

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
