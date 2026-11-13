# From manual kubeadm → Ansible automation

You already know *what* needs to happen on each VM (containerd, kubeadm, kubelet,
kubectl, `kubeadm init`, Flannel). Ansible's only job is to run those exact steps
for you, on every node, reliably, every time. Nothing conceptually new — just a
different way of executing the same shell commands.

## 1. Ansible basics you need before writing any code

**Push-based, agentless.** Unlike Terraform (which talks to the OpenStack API),
Ansible just SSHes into machines and runs commands/modules. No agent to install
on the target — only Python needs to exist there (it does, on Ubuntu).

**Inventory** — a file listing which hosts to manage, grouped:
```ini
[control_plane]
10.0.10.28

[workers]
10.0.10.29

[k8s_cluster:children]
control_plane
workers
```
`[k8s_cluster:children]` creates a group-of-groups so you can target "all
k8s nodes" in one play.

**Playbook** — a YAML file listing *plays*. Each play = "on these hosts, run
these tasks."

**Task** — one action, almost always via a *module* (not raw bash), e.g.:
```yaml
- name: Install containerd
  ansible.builtin.apt:
    name: containerd
    state: present
    update_cache: true
```
Modules are idempotent — running the playbook twice doesn't break anything or
reinstall things unnecessarily. This is the biggest mental shift from bash
scripts: you declare *desired state*, not a sequence of commands. Use
`ansible.builtin.command`/`shell` only as an escape hatch when no module fits
(e.g. `kubeadm init`).

**Roles** — a directory convention for organizing reusable tasks
(`roles/<name>/tasks/main.yml`, `roles/<name>/templates/`, etc.). This project
uses one role per logical unit: `common`, `containerd`, `kube_pkgs`,
`control_plane`, `cni_flannel`, `worker_join`.

**Variables** — `group_vars/all.yml` applies to every host; you can override
per-host or per-group. Used here for the Kubernetes version, pod CIDR, etc.,
so you change it in one place.

**Templates (Jinja2, `.j2`)** — files rendered with variables substituted in,
via the `template` module. Used here for `/etc/containerd/config.toml` and for
the inventory file itself.

**Handlers** — tasks that only run when notified by another task (e.g.
"restart containerd" only fires if the config file actually changed).

**`become: true`** — the Ansible equivalent of `sudo` in front of a task.

**Facts** — auto-gathered info about each host (IP, OS, etc.), available as
`ansible_facts` / `ansible_default_ipv4.address`. This is how the worker node
learns the control-plane's IP without you hardcoding it.

**`delegate_to` / `hostvars`** — how one host's play can reference another
host's data. This is exactly the mechanism used to grab the `kubeadm join`
command generated on the control-plane and hand it to the workers.

**`ansible.cfg`** — local config (inventory path, disables SSH host-key
checking so first-run automation doesn't hang on a prompt).

**Ad-hoc commands** — a one-off module run without a playbook, good for
sanity-checking connectivity before you write anything:
```sh
ansible all -i ansible/inventory/hosts.ini -m ping
```
If that doesn't return `pong` from every host, nothing else will work — fix
SSH/inventory first.

**Terraform vs Ansible, in one line:** Terraform creates *infrastructure*
(the VM exists or doesn't). Ansible configures *what's running inside* it.
They're chained: Terraform's output (the floating IP) becomes Ansible's input
(the inventory).

## 2. Project layout

```
k8s-ansible-automation/
├── terraform/
│   ├── single-node/        # 1 VM
│   └── two-node/           # control-plane VM + worker VM
├── ansible/
│   ├── ansible.cfg
│   ├── generate_inventory.sh   # terraform output -> inventory/hosts.ini
│   ├── group_vars/all.yml      # k8s version, pod CIDR, etc.
│   ├── site.yml                # the master playbook
│   └── roles/
│       ├── common/          # prep: packages, disable swap, kernel modules
│       ├── containerd/      # container runtime
│       ├── kube_pkgs/       # kubeadm / kubelet / kubectl, pinned + held
│       ├── control_plane/   # kubeadm init, kubeconfig, join-command capture
│       ├── cni_flannel/     # pod networking
│       └── worker_join/     # kubeadm join (two-node only)
└── bootstrap.sh              # the single command
```

## 3. One command, start to finish

```sh
# Single node
./bootstrap.sh single-node

# Two nodes (control-plane + worker)
./bootstrap.sh two-node
```
This runs `terraform apply -auto-approve`, waits for SSH to come up, generates
the inventory from Terraform's outputs, waits for `ansible -m ping` to succeed,
then runs `ansible-playbook site.yml`. When it finishes, `kubectl get nodes`
against the printed floating IP shows all nodes `Ready`.

## 4. What's different from your manual notes (and why)

- **swap is disabled and `br_netfilter`/`overlay` kernel modules + sysctl
  bridge settings are added in `common`.** Your manual notes didn't spell
  these out, but `kubeadm init`/`join` fail without them on a fresh VM — I
  added them since they're required, not optional.
- Package versions and pod CIDR are variables in `group_vars/all.yml`, not
  hardcoded, so upgrading Kubernetes later means editing one file.
- The `kubeadm join` command is captured automatically from the control-plane
  node's `kubeadm init` output and pushed to the worker(s) — you never
  copy/paste a token by hand.