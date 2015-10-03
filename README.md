# Gattai

Mission control for Docker

`gattai` is a client for Docker with the orchestration workflow in mind.
It is based on `docker/cli` so its commands are fully compatible with `docker`.
Basically, `gattai` can be a drop-in replacement for `docker` client.

In addition, `gattai` includes `libmachine` for machine provisioning, codes from `swarm` for built-in discovery management. It's planned to include `libcompose` for service composition in future releases.

## Quick Start

You can setup a Docker cluster with `gattai` in just 4 steps.

  * `$ gattai init` will initialize a Gattai repository to store workflow-related files.
  * Open and edit `provision.yml` as follows:

```
---
machines:
  ocean:
    driver: digitalocean
    instances: 2
    options:
      digitalocean-access-token: $DIGITALOCEAN_ACCESS_TOKEN
      digitalocean-region: sgp1
      engine-install-url: "https://experimental.docker.com"

  master:
    from: ocean

```
  * `$ gattai provision` will read `provision.yml` and prepare 3 machines, `ocean-1`, `ocean-2` and `master`, for you.
  * `$ gattai cluster -m master ocean` will form a Docker cluster using `swarm`. The `master` machine will be your `swarm` manager and set active.

 After the `cluster` command, you can now use `gattai` docker commands to place containers on your cluster.

