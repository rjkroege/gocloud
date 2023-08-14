Some tools to help use Google Cloud compute nodes.

* Run `sessionender` as part of launching a new node that you intend to use interactively.
It shuts the node down when the login tty has been idle for more than a configurable
value (default is 15 minutes) to save money.

* `makecloudconfig` assembles a single cloudconfig file for a GCP COS node out of a list
of services. This makes it easy to pack a number of Docker containers into a single node.
Presumably I could have used Kubernetes or `docker compose` or something similar to do
this but this was easy, predictable and cheap.

* `gocloud`  tool to launch a GCP node without using `gcloud`. Smaller and faster than
using `gcloud` but does so much less. How to use:

	* make a configuration file in [TOML: Tom's Obvious Minimal Language](https://toml.io/en/) toml format

		```toml
		defaultzone = "us-east1-b"
		projectid = "liquiorg"

		[instance.smallnodisk]
			hardware = "e2-small"
			family = "cos-cloud"
			userdatafile = "path to your couldconfig file name"
			githost = "Git repository to checkout for system setup"
			postsshconfig = "Script to run on node bringup"
		```
	
	* Make one:

		```shell
		gocloud make smallnodisk myinstance
		```
	
	* I have some related tooling to provision the node. A bare node needs
	a useful `cloudconfig` file, a configured service account, etc.

	* I use the [mkconfig](https://github.com/rjkroege/mkconfig) tool to setup a node from
	within the cloudconfig file.

	* On MacOS, `gocloud` wants to read selected configuration that it will push to the GCP
	metadata service from the MacOS KeyChain. The `gocloud show-meta` subcommand
	will show if this is configured correctly.

