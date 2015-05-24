#user's guide#

The easiest way to put prevmtable into action does not involve building any code. The dockerhub skelterjohn/prevmtable image will be kept reasonably up to date, and at some point may be on autobuild.

If you have base a real system on prevmtable, you should build the image yourself.

The `prevmtable-up.bash` script will maintain one preemptible f1-micro VM in us-central1-b or us-central1-f, and that VM will use kubernetes to keep up a very simple "Hello, world!" web server.

A firewall rule is created to allow TCP on :8080 to connect to those VMs, though load balancing is an exercise left to the reader. Future work is a post-create hook that will allow adding of the instance to a target pool for load balancing.