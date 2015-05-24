#user's guide#

The easiest way to put prevmtable into action does not involve building any code. The dockerhub skelterjohn/prevmtable image will be kept reasonably up to date, and at some point may be on autobuild.

If you have base a real system on prevmtable, you should build the image yourself.

The `prevmtable-up.bash` script will maintain 3 preemptible f1-micro VMs in us-central1-b or us-central1-c, and that VM run a very simple "Hello, world!" web server.

Along with the preemptible VMs, it will also create the prevmtable-master, using kubernetes to keep it going. The prevmtable master runs the prevmtable service, and will create and delete VMs as needed.

A firewall rule is created to allow TCP on :8080 to connect to those VMs, and a forwarding rule and target pool are created for load balancing. A prevmtable post-create hook adds new instances to the target pool, creating an accessible load-balanced cluster.
