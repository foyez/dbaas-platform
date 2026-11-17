# Linux users, groups and permissions

!!! abstract "Where it's used in this project"
    SSH key permissions and a non-root user for the OpenStack VMs; `sudo` for Ansible's `become`; and inside Kubernetes the same ideas reappear as `securityContext`: the UI runs as non-root and the API drops every Linux capability ([`gitops/api/base/deployment.yaml`](https://github.com/foyez/dbaas-platform/blob/main/gitops/api/base/deployment.yaml)). Overview: [Linux fundamentals](linux-fundamentals.md).

## Key ideas

### Users and groups

Every process and every file belongs to a **user** and a **group**. That's how Linux keeps one program from reading or breaking another's data.

| Concept | Meaning | Stored in |
|---|---|---|
| UID | Numeric user ID; the kernel only sees numbers | `/etc/passwd` |
| GID | Numeric group ID | `/etc/group` |
| Primary group | Default group for new files | `/etc/passwd` |
| Supplementary groups | Extra memberships (`sudo`, `docker`, `adm`) | `/etc/group` |
| Password hash | Only readable by root | `/etc/shadow` |

```bash
id                     # uid=1000(ubuntu) gid=1000(ubuntu) groups=1000(ubuntu),27(sudo),4(adm)
id deploy
getent passwd deploy   # deploy:x:1001:1001:CI/CD Deploy User:/home/deploy:/bin/bash
```

| User | UID | Purpose |
|---|---|---|
| `root` | 0 | Superuser; permission checks don't apply |
| system users | < 1000 | Run services: `www-data`, `postgres`, `sshd` |
| `nobody` | 65534 | Minimal-privilege identity |
| regular users | ≥ 1000 | People |

**Least privilege:** nginx runs its workers as `www-data`, PostgreSQL as `postgres`. If one is exploited, the attacker has that user's rights, not root's: no reading `/etc/shadow`, no installing software system-wide.

### Managing users and groups

```bash
sudo useradd -m -s /bin/bash -c "CI/CD Deploy User" deploy   # -m home, -s shell
sudo useradd -r -s /usr/sbin/nologin myapp                   # system user for a service, no login
sudo passwd deploy
sudo usermod -aG developers deploy     # -a APPEND; without it -G replaces all groups
sudo usermod -L deploy                 # lock; -U unlock
sudo gpasswd -d deploy developers      # remove from a group
sudo userdel -r deploy                 # delete with home directory
sudo groupadd developers
```

`useradd` is the low-level, scriptable tool; `adduser` is Debian's interactive wrapper. Group changes only apply to **new** login sessions, so log out and back in (or `newgrp`).

### sudo

`sudo` runs a command as root (or another user, `sudo -u postgres psql`) after checking `/etc/sudoers` and `/etc/sudoers.d/`. Members of the `sudo` group (Ubuntu) or `wheel` (RHEL) may run anything. Rules can be much narrower:

```text
# /etc/sudoers.d/deploy  — edit with: sudo visudo -f /etc/sudoers.d/deploy
deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart myapp
```

`visudo` checks the syntax before saving; a broken sudoers file can lock everyone out of sudo. Every sudo use is logged to `/var/log/auth.log`.

### Reading permissions

```text
-rwxr-x---  1 deploy deployers 8421504 Mar 28 10:00 server
│└┬┘└┬┘└┬┘
│ │  │  └── others: ---
│ │  └───── group (deployers): r-x
│ └──────── user/owner (deploy): rwx
└────────── type: - file, d directory, l symlink
```

| Bit | On a file | On a directory |
|---|---|---|
| `r` (4) | Read contents | List names (`ls`) |
| `w` (2) | Modify contents | Create, delete, rename entries inside |
| `x` (1) | Execute | Enter it and reach files inside (`cd`, open by path) |

The kernel checks **owner first, then group, then others**, and uses the first that matches. Deleting a file needs `w` on the *directory*, not on the file.

**Numeric mode:** add up 4 + 2 + 1 per class, owner-group-others: `rwx r-x r-- = 754`.

| Mode | Symbolic | Typical use |
|---|---|---|
| `755` | `rwxr-xr-x` | Programs, scripts, public directories |
| `644` | `rw-r--r--` | Config files, web content |
| `640` | `rw-r-----` | Secrets readable by a service's group |
| `600` | `rw-------` | SSH private keys, kubeconfigs, `.env` files |
| `700` | `rwx------` | `~/.ssh/`, private directories |
| `775` / `664` | `rwxrwxr-x` / `rw-rw-r--` | Shared team directories and files |

### Changing permissions and ownership

```bash
chmod 600 ~/.ssh/id_ed25519      # numeric: set everything at once
chmod u+x deploy.sh              # symbolic: change one bit, keep the rest
chmod g-w,o-r secrets.env
chmod -R 755 /var/www/app        # recursive (careful: makes every file executable too)
sudo chown deploy:deployers server
sudo chown -R deploy: /home/deploy/.ssh     # owner and their primary group
sudo chgrp deployers /var/www/app
```

Use numeric in scripts (explicit, predictable) and symbolic for one-bit tweaks. For directory trees, `find dir -type d -exec chmod 755 {} +` and `-type f ... 644` avoid making files executable.

**umask** sets the default for new files: with the common `022`, files are created `644` and directories `755`. `umask 077` makes new files private (`600`).

**SSH is strict:** a private key readable by group or others is refused with `Permissions 0644 for 'id_ed25519' are too open`. `~/.ssh` must be `700`, keys and `authorized_keys` `600`, and the home directory must not be group-writable.

### Special bits

| Bit | Numeric | Shown as | On files | On directories |
|---|---|---|---|---|
| **setuid** | `4xxx` | `s` in owner's `x` | Runs as the file's **owner** (`/usr/bin/passwd` runs as root to edit `/etc/shadow`) | (ignored) |
| **setgid** | `2xxx` | `s` in group's `x` | Runs with the file's group | New files **inherit the directory's group**: ideal for shared team dirs |
| **sticky** | `1xxx` | `t` in others' `x` | (ignored) | Only a file's owner may delete it, even with `w` on the dir: `/tmp` is `1777` |

```bash
sudo chmod 2775 /srv/shared        # team directory: group inherited, group-writable
sudo chmod 1777 /srv/uploads       # anyone writes, only owners delete
find / -perm -4000 -type f 2>/dev/null    # audit setuid binaries
```

setuid is ignored on scripts and is a classic privilege-escalation target, so avoid creating setuid binaries.

### Capabilities: splitting up root

Root's power is actually a set of **capabilities**: `CAP_NET_BIND_SERVICE` (bind ports below 1024), `CAP_NET_ADMIN` (change network config), `CAP_SYS_ADMIN` (almost everything), `CAP_CHOWN`, and more. A process can hold just the ones it needs. That's how CoreDNS listens on port 53 as a non-root user in the Very Hard Way course, and what `capabilities: drop: [ALL]` removes in a Pod spec.

### The same ideas in containers

| Linux | Kubernetes `securityContext` |
|---|---|
| Don't run services as root | `runAsNonRoot: true`, `runAsUser: 10001` |
| Files owned by the service's group | `fsGroup`: volumes are made group-owned and group-writable |
| No setuid escalation | `allowPrivilegeEscalation: false` (sets the kernel's `no_new_privs`) |
| Only the capabilities needed | `capabilities: {drop: [ALL]}` |
| Read-only system directories | `readOnlyRootFilesystem: true` |

UIDs are just numbers: a container's UID 1000 is the host's UID 1000 unless user namespaces remap it. That's one reason running containers as root is risky.

## What I learned building it

- **`usermod -G` without `-a` silently removes a user from every other group**, including `sudo`. Always `-aG`.
- **SSH key errors are usually permissions.** `Permissions ... are too open` and silent key rejection both come from loose modes on the key, `~/.ssh`, or `authorized_keys`.
- **Kubernetes security settings are Linux settings with YAML names.** Understanding UIDs, groups and capabilities made `securityContext` obvious, including why the API drops all capabilities and mounts a read-only root filesystem.
- **Mounted volumes and non-root containers clash** until the files' group matches; `fsGroup` fixes it without making things world-writable.

## Review questions

??? question "What does `chmod 640 secrets.env` with owner `root` and group `myapp` allow?"
    Root can read and write; members of `myapp` (the service user) can read; everyone else has no access.

??? question "Why does deleting a file depend on the directory's permissions?"
    Deleting removes an entry from the directory, which is a change to the directory. It needs `w` (and `x`) on the directory; the file's own bits don't matter, unless the sticky bit is set.

??? question "What's the difference between `usermod -G docker alice` and `usermod -aG docker alice`?"
    Without `-a`, `docker` becomes alice's only supplementary group, removing all others. With `-a` it's appended.

??? question "Why is `/tmp` mode `1777`?"
    Everyone must be able to create files there (`777`), but the sticky bit (`1`) stops users from deleting or renaming each other's files.

??? question "How can `passwd` change `/etc/shadow` when run by a normal user?"
    It's owned by root and has the setuid bit, so it runs with root's effective UID.

??? question "How do you let a non-root process listen on port 443?"
    Give it `CAP_NET_BIND_SERVICE` (`setcap` on the binary, `AmbientCapabilities=` in a systemd unit, or a capability in the Pod spec), or have it listen on a high port behind a proxy or Service.

??? question "Map `runAsNonRoot`, `fsGroup` and `capabilities.drop: [ALL]` to Linux concepts."
    Not running as UID 0; making volume files owned by a supplementary group the process belongs to; removing all of root's capabilities from the process.
