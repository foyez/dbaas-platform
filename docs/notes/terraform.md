# Terraform

!!! abstract "Where it's used in this project"
    Every piece of infrastructure is Terraform: OpenStack VMs ([`infra/openstack/*/tf`](https://github.com/foyez/dbaas-platform/tree/main/infra/openstack)), the SKE cluster ([`infra/ske-cluster/terraform`](https://github.com/foyez/dbaas-platform/tree/main/infra/ske-cluster/terraform)) and ZITADEL ([`infra/zitadel/tf`](https://github.com/foyez/dbaas-platform/tree/main/infra/zitadel/tf)), which `make dev` also applies locally.

## Key ideas

**Infrastructure as Code** means describing infrastructure in versioned files instead of clicking or typing one-off commands. Terraform is a **declarative provisioning** tool: you write the end state, and it works out what to create, change or delete.

| Concept | Meaning |
|---|---|
| Provider | Plugin that talks to one API (OpenStack, STACKIT, ZITADEL…) |
| Resource | Something Terraform creates and owns |
| Data source | Something that already exists and Terraform only reads |
| State | Terraform's record of what it manages, mapping config to real IDs |
| Backend | Where state lives: local file or remote (S3, with locking) |
| Module | A reusable folder of configuration with inputs and outputs |
| Plan / apply | Show the diff, then make it real |

**Where Terraform fits next to other tools:**

- **Provisioning vs configuration management.** Terraform creates servers, networks and clusters. Ansible configures software on servers that already exist. This project uses both: Terraform for the VMs, Ansible for Kubernetes on them.
- **Declarative vs procedural.** Re-running a declarative config is safe because it converges on the same state. Re-running a procedural script ("add 10 servers") repeats the action.
- **Immutable vs mutable.** Replacing resources avoids the configuration drift you get from patching servers in place.
- Terraform is **not** a write-once-deploy-anywhere tool. Each provider's resources are different; only the workflow is portable.

## What I learned building it

- **Same API, different bookkeeping.** Rebuilding the manual OpenStack lab in Terraform called exactly the same Nova and Neutron endpoints. What changes is who tracks dependency order, remembers what exists and reverses it safely on teardown: you, or the tool.
- **`plan` has no manual equivalent.** No `openstack … create` command can show you what it's about to change. That preview is the main safety net.
- **Terraform stops at the guest OS.** Formatting a disk or installing packages inside a VM isn't an API call. Use cloud-init `user_data`, or hand over to Ansible.
- **Never put credentials in `.tf` files.** OpenStack credentials come from `clouds.yaml` and STACKIT credentials from a service-account key. State files can contain secrets too: the ZITADEL module's API key ends up in state, which is why creating it in Terraform is opt-in.
- **Use remote state for anything shared.** The SKE setup stores state in a STACKIT Object Storage bucket through the S3 backend, so it isn't lost with a laptop and can be locked against concurrent applies.
- **Use data sources to bridge the gap with things created outside Terraform.** The local ZITADEL setup looks up the org and admin user that ZITADEL created at startup, instead of hardcoding their IDs.
- **`count` can't depend on sensitive values.** Making Google login optional needed `nonsensitive(var.google_client_id != "")`.
- **Refactoring needs `moved` blocks.** Adding `count` to an existing resource changes its address; a `moved` block stops Terraform from planning a destroy and recreate.

## Review questions

??? question "What's in Terraform state, and why does it matter where it's stored?"
    The mapping from each resource in your config to the real object's ID and attributes, sometimes including secrets. If it's lost, Terraform no longer knows what it manages. If two people apply against different copies, they overwrite each other. Remote state with locking solves both.

??? question "Resource vs data source?"
    A resource is created, updated and destroyed by Terraform. A data source only reads something that already exists, such as an image ID or an org created by another tool.

??? question "Why did this project use Terraform and Ansible together instead of one of them?"
    Terraform is good at creating cloud resources through APIs and tracking them. Ansible is good at configuring the OS and software inside machines over SSH. Installing Kubernetes is the second kind of job.

??? question "You add `count` to an existing resource and `plan` wants to destroy and recreate it. Why, and what's the fix?"
    The resource's address changed from `x.y` to `x.y[0]`, so Terraform sees one resource removed and another added. A `moved { from = x.y  to = x.y[0] }` block tells it they're the same object.
