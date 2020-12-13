Some tools to help use Google Cloud compute nodes.

* Run `sessionender` as part of launching a new node that you intend to use interactively.
It shuts the node down when the login tty has been idle for more than a configurable
value (default is 15 minutes) to save money.

* `makecloudconfig` assembles a single cloudconfig file for a GCP COS node out of a list
of services. This makes it easy to pack a number of Docker containers into a single node.
Presumably I could have used Kubernetes or `docker compose` or something similar to do
this but this was easy, predictable and cheap.

* `gocloud` extremely WIP tool to launch a GCP node without using `gcloud`

Of these, only `sessionender` might be of general interest.