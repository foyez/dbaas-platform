# Ansible Fundamentals

## What is Ansible?

**Ansible** is an open-source, **agentless** IT automation tool. No software runs on target machines — it connects via **SSH** (Linux) or **WinRM** (Windows) and does the work.

Used for: server provisioning, software installs, configuration, app deployment, cloud/infra management.

**Why use it:** instead of manually running the same commands on 50 servers, you write one playbook and run:

```bash
ansible-playbook nginx.yml
```

→ consistent, repeatable, less human error, saves time.

---

## The Declarative Approach (core idea)

Ansible doesn't tell the system *how* to do something — it describes the **desired end state**, and Ansible figures out what (if anything) needs to change.

```bash
apt install nginx        # imperative: "run this command"
```
```yaml
state: present            # declarative: "I want nginx to exist"
```

Ansible checks the current state first, then only acts if there's a difference. This is the foundation of **idempotency**.

---

## Idempotency

> Running the same playbook multiple times produces the same end result, without unnecessary changes.

```yaml
- name: Install nginx
  apt:
    name: nginx
    state: present
```

* 1st run → nginx gets installed
* 2nd run → nothing happens (already satisfies desired state)

This makes automation **safe to rerun** anytime.

---

## `state`: the key parameter

`state` tells a module what end result you want. Its allowed values depend on the module.

**`apt` / `yum` / `dnf` (packages)**

| Value | Meaning |
|---|---|
| `present` | ensure installed |
| `absent` | ensure removed |
| `latest` | ensure installed **and** up to date |

**`service` (services)**

| Value | Meaning |
|---|---|
| `started` | ensure running |
| `stopped` | ensure not running |
| `restarted` | restart it |
| `reloaded` | reload config without full restart |

**`file` (files/directories)**

| Value | Meaning |
|---|---|
| `directory` | ensure it's a directory |
| `file` | ensure it's a regular file |
| `touch` | create empty file if missing |
| `absent` | ensure removed |
| `link` | ensure it's a symlink |

```yaml
- name: Create a directory
  file:
    path: /opt/myapp
    state: directory
```

---

## Core Building Blocks

```
Inventory → Hosts/Groups → Playbook → Play → Tasks → Modules → Managed Hosts
```

| Concept | What it is |
|---|---|
| **Inventory** | List of machines Ansible manages (static file or dynamic — AWS, GCP, etc.) |
| **Host** | One managed machine |
| **Group** | A named set of hosts, so you can target them together |
| **Playbook** | A YAML file containing one or more **plays** |
| **Play** | Maps a set of **tasks** to a set of **hosts** |
| **Task** | One action to perform (e.g. install a package) |
| **Module** | The actual unit of work a task calls (`apt`, `copy`, `service`, etc.) |
| **Variables** | Make playbooks reusable (`{{ package_name }}`) |
| **Facts** | Auto-collected info about a host (`ansible_hostname`, `ansible_distribution`...) |
| **Handlers** | Run **only when notified** by a task — e.g. restart a service after config changes |
| **Roles** | Standard folder structure for organizing reusable automation |
| **Templates** | Config files with variables, rendered before copying to a host |

### Example inventory
```ini
[web]
server1.example.com
server2.example.com

[database]
db1.example.com
```

### Example playbook
```yaml
---
- name: Configure Web Servers
  hosts: web
  become: yes

  vars:
    package_name: nginx

  tasks:
    - name: Install package
      apt:
        name: "{{ package_name }}"
        state: present

    - name: Start service
      service:
        name: nginx
        state: started
```

### Handlers example
```yaml
tasks:
  - name: Copy config
    copy:
      src: nginx.conf
      dest: /etc/nginx/nginx.conf
    notify: Restart nginx

handlers:
  - name: Restart nginx
    service:
      name: nginx
      state: restarted
```
Handler only fires if the task above it actually changed something.

---

## Ad-hoc Commands (quick one-liners, no playbook)

```bash
ansible all -m ping                                   # check connectivity
ansible web -m apt -a "name=nginx state=present"       # install nginx
ansible web -m service -a "name=nginx state=restarted" # restart service
```

---

## Commonly Used Modules

| Module | Purpose |
|---|---|
| `apt` / `yum` / `dnf` | Package management (Debian / RHEL / Fedora) |
| `copy` | Copy files as-is |
| `template` | Copy files with variables rendered |
| `file` | Manage files/directories |
| `service` | Start/stop/restart services |
| `user` / `group` | Manage users/groups |
| `git` | Clone/update repos |
| `command` | Run a command (no shell features) |
| `shell` | Run shell commands (pipes, redirects allowed) |
| `cron` | Manage cron jobs |
| `ping` | Test connectivity |

---

## Must-Remember Takeaways

1. Ansible is **agentless** — SSH/WinRM only, no software on targets.
2. It's **declarative** — you say *what* you want, not *how* to get there.
3. **Idempotency** = safe to rerun; nothing changes if already in desired state.
4. `state` is the parameter that expresses the desired end state — its valid values change per module.
5. Structure: **Inventory → Playbook → Play → Task → Module**.
6. **Handlers** only run when notified (usually after a change) — great for restarts.
7. **Roles** = reusable, organized automation for bigger projects.