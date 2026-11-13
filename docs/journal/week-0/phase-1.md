# Phase 1 — Foundation

> Core Idea first, flow diagrams, key files/commands, then memory tricks. Each section is meant to be re-readable in 5 minutes before an interview.

## Table of Contents
1. [Users & Groups](#1-users--groups)
2. [Command Line & Linux Basics](#2-command-line--linux-basics)
3. [systemd](#3-systemd)
4. [SSH](#4-ssh)
5. [journald](#5-journald)
6. [rsyslog](#6-rsyslog)
7. [firewalld](#7-firewalld)
8. [nftables](#8-nftables)
9. [Package Management](#9-package-management)
10. [Git](#10-git)

---

# 1. Users & Groups

## Core Idea ⭐

> **Humans use names → Linux kernel uses numbers.**

The kernel **never** checks usernames or group names during a permission check — it only compares **UIDs and GIDs**. `/etc/passwd` and `/etc/group` are just lookup tables mapping names to those numbers.

---

## Every User Has

```text
kazi
 ├── UID = 1000
 ├── Primary GID = 1000
 └── Supplementary Groups = docker, sudo, libvirt
```

## Permission Check Flow

```text
Process wants to access a file
      ↓
Kernel compares:
  Process UID  ↔ File Owner UID
  Process GIDs ↔ File Group GID
  Permission bits (rwx)
      ↓
1000 == 1000 ✔   (numbers, never "kazi == kazi")
      ↓
Allow / Deny
```

## Key Files

```text
/etc/passwd   username → UID/GID, home dir, shell   (NOT passwords)
/etc/group    group name → GID, member list
/etc/shadow   password hashes + aging — root-only readable
```

<details>
<summary><strong>/etc/passwd anatomy</strong></summary>

```
kazi:x:1000:1000:Kazi:/home/kazi:/bin/bash
 │    │  │    │    │      │         └─ login shell
 │    │  │    │    │      └─ home dir
 │    │  │    │    └─ GECOS (full name/comment)
 │    │  │    └─ primary GID
 │    │  └─ UID
 │    └─ password placeholder (real hash lives in /etc/shadow)
 └─ username
```

UID ranges: `0` = root · `1–999` = system/service accounts (`nova`, `neutron`, `libvirt-qemu`) · `1000+` = human users.

</details>

<details>
<summary><strong>Core commands</strong></summary>

```bash
useradd -m -s /bin/bash -G docker,sudo kazi   # create, home dir, shell, secondary groups
usermod -aG libvirt kazi                       # APPEND to secondary groups — -G alone wipes them
userdel -r kazi                                # delete user + home dir

groupadd deployers
gpasswd -a kazi deployers                      # add to group
gpasswd -d kazi deployers                      # remove from group

id kazi        # UID, GID, all group memberships
groups kazi
whoami
```

**The `-aG` vs `-G` mistake is the #1 way people accidentally lock themselves out of `docker`/`sudo`/`libvirt` groups.** Always append.

</details>

<details>
<summary><strong>Password aging & sudo</strong></summary>

```bash
chage -l kazi           # list aging info
passwd -l kazi           # lock account (prepends ! to hash)
passwd -u kazi           # unlock

su -                      # full login shell as target user (needs THEIR password)
sudo -u nova whoami       # run as a specific service user (common when debugging OpenStack)
sudo -l                   # what am I allowed to run
visudo                    # edit /etc/sudoers safely — validates syntax before saving
```

</details>

## Login Flow

```text
Enter username/password
      ↓
/etc/passwd → resolve username to UID/GID
      ↓
/etc/shadow → verify password hash
      ↓
Shell starts with that UID/GID
      ↓
Every child process inherits UID/GID
```

## Easy Memory Flow

```text
Username
   ↓ /etc/passwd
UID + GID
   ↓
Shell → Processes
   ↓
Kernel permission check
   ↓
Compare NUMBERS (UID/GID), not names
   ↓
Access Granted / Denied
```

## One-Line Memory Tricks

* **Linux identifies users by UID, not username — the kernel never sees names.**
* **Every process runs with a UID and GID, inherited from its parent shell.**
* **`/etc/passwd` → identity mapping. `/etc/group` → group mapping. `/etc/shadow` → password hashes.**
* **`usermod -aG` appends; `usermod -G` (no `-a`) silently wipes other group memberships.**
* **Every OpenStack service runs as its own system user (`nova`, `neutron`, `cinder`) — "Permission denied" in a service log is almost always a UID/group mismatch, not a bug.**

---

# 2. Command Line & Linux Basics

## Core Idea ⭐

> **Everything is a file, and every tool speaks text in / text out.**

That's *why* the shell is the primary interface for automation — text can be piped, scripted, and run identically a thousand times. A GUI can't be composed like that.

---

## The Filesystem Is One Tree

```text
/
├── etc      → config files (almost everything you edit lives here)
├── var/log  → logs
├── home     → user directories
├── usr/bin  → installed programs
├── tmp      → scratch space, wiped on reboot
```

No `C:\`, `D:\` — everything, including other disks, mounts *under* `/`.

---

## Permissions — Owner / Group / Other

```text
-rwxr-xr--
│└┬┘└┬┘└┬┘
│ │  │  └─ other: r-- (read only)
│ │  └──── group: r-x (read+execute)
│ └─────── owner: rwx (read+write+execute)
└───────── file type (- = file, d = directory)
```

```bash
chmod 644 file        # owner rw, group r, other r
chmod +x script.sh    # add execute for everyone
chown user:group file # change owner + group
```

**Numeric shortcut:** r=4, w=2, x=1 → sum per group. `644` = rw-r--r--, `755` = rwxr-xr-x.

---

## Processes & Pipes

```text
command1  |  command2  |  command3
 (stdout)     (stdin→stdout)  (stdin)
```

```bash
ps aux              # list all running processes
kill <pid>           # ask nicely to stop
kill -9 <pid>        # force stop
command &            # run in background
jobs                 # list background jobs
grep pattern | wc -l # pipe: count matching lines
```

<details>
<summary><strong>Why background/foreground matters</strong></summary>

A foreground command blocks your terminal until it exits. Long-lived services shouldn't be run this way by hand — that's exactly the gap **systemd** fills (§2): supervising processes so a human doesn't have to keep a terminal open.

</details>

---

## Easy Memory Flow

```text
You type a command
      ↓
Shell parses it
      ↓
Shell forks a process
      ↓
Kernel schedules it, enforces permissions (UID/GID from Users & Groups sheet)
      ↓
stdout/stderr returned as text
      ↓
(optionally) piped into the next command
```

## One-Line Memory Tricks

* **Permissions are checked per file, per owner/group/other, using UID/GID — not usernames.**
* **`chmod` changes *what's allowed*; `chown` changes *who* it applies to.**
* **Pipes compose small tools into pipelines — the core Unix philosophy.**
* **The shell is scriptable because it's text in/text out — that's the whole reason CLI-first tooling dominates cloud engineering.**

---

# 3. systemd

## Core Idea ⭐

> **systemd is PID 1 — it starts everything else, in the right order, and restarts what dies.**

Without a supervisor, "my service crashed at 3am" means "down until someone notices." systemd exists to make that not true.

---

## Unit Lifecycle

```text
System boots
      ↓
systemd (PID 1) starts
      ↓
Reads target (multi-user.target, etc.)
      ↓
Starts units in dependency order (After=, Requires=, Wants=)
      ↓
Unit crashes?  → Restart= policy decides if/how it comes back
```

## Unit File Anatomy

```ini
[Unit]
After=network-online.target      # ordering only
Requires=postgresql.service      # hard dependency

[Service]
Type=simple                      # simple | forking | notify | oneshot
ExecStart=/usr/bin/myapp
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

`After` ≠ `Requires`: **After is ordering, Requires is dependency.** Most real units need both.

<details>
<summary><strong>Core commands</strong></summary>

```bash
systemctl start|stop|restart|reload <unit>
systemctl status <unit> -l --no-pager
systemctl enable --now <unit>          # enable at boot + start now
systemctl daemon-reload                # ALWAYS after editing a unit file
systemctl list-units --state=failed    # find everything broken
systemctl cat <unit>                   # print effective unit file
journalctl -u <unit> -f                # follow its logs (see §4)
```

</details>

## Easy Memory Flow

```text
Unit file declared
      ↓
daemon-reload (teach systemd about the change)
      ↓
enable (symlink into target's .wants/ — boot-time)
      ↓
start (run now)
      ↓
crash → Restart= policy → systemd relaunches it
```

## One-Line Memory Tricks

* **systemd = PID 1 = supervisor of everything else.**
* **`After` is ordering. `Requires`/`Wants` is dependency. They're independent.**
* **Forgetting `daemon-reload` after editing a unit file is the #1 "my change didn't apply" mistake.**
* **`systemctl edit` creates an override, not a raw edit — survives package upgrades.**

---

# 4. SSH

## Core Idea ⭐

> **Two independent trust decisions: you trust the server (host key), the server trusts you (your key).**

Both use asymmetric crypto, but they're separate checks — mixing them up is a common source of confusion.

---

## Connection Flow

```text
Client connects
      ↓
Server presents host key → client checks ~/.ssh/known_hosts
      ↓ (first time: "are you sure?" prompt)
Client proves identity:
      - has matching private key for a pubkey in server's authorized_keys?
      ↓
Session established, encrypted
```

## Key Files

```text
~/.ssh/id_ed25519       (private key, chmod 600 — NEVER share)
~/.ssh/id_ed25519.pub   (public key, chmod 644 — safe to share)
~/.ssh/authorized_keys  (on the SERVER — which pubkeys may log in)
~/.ssh/known_hosts      (on the CLIENT — which host keys are trusted)
~/.ssh/config           (client shortcuts: Host aliases, ProxyJump, etc.)
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
ssh-keygen -t ed25519 -C "you@host"
ssh-copy-id user@host                    # push your pubkey to their authorized_keys
ssh -vvv user@host                       # verbose — see exactly which auth step failed
ssh -J bastion user@internal-host        # jump host, one-off
ssh -L 8080:localhost:80 user@host       # local port forward
sshd -t                                  # test server config BEFORE reloading
sudo journalctl -u sshd -f               # server-side, live
```

</details>

## Easy Memory Flow

```text
Generate keypair
      ↓
Public key → server's authorized_keys
      ↓
Private key stays on your machine, chmod 600
      ↓
Connect: server challenges, you prove possession of the private key
      ↓
No password ever crosses the wire
```

## One-Line Memory Tricks

* **Private key never leaves your machine — that's the whole security model.**
* **Wrong permissions on `~/.ssh` or the key files = silent, confusing refusal.**
* **`known_hosts` = you trust them. `authorized_keys` = they trust you. Different files, different directions.**
* **`ssh -vvv` before you guess — it tells you exactly which auth method failed.**

---

# 5. journald

## Core Idea ⭐

> **Structured, indexed logs, owned by systemd — filter by field, not by grepping text.**

Every log line is key=value (`_SYSTEMD_UNIT`, `PRIORITY`, `MESSAGE`...), which is what makes precise filtering fast.

---

## Flow

```text
App calls syslog() / writes to stdout under systemd
      ↓
journald captures it, tags with structured fields
      ↓
Stored (RAM-only by default, or persistent if configured)
      ↓
journalctl queries by unit / time / priority / any field
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
journalctl -u nova-compute -f              # follow one unit's logs
journalctl -u nova-compute -p err          # priority filter (emerg..debug)
journalctl -u nova-compute --since "10 min ago"
journalctl -b                              # this boot only
journalctl -b -1                           # previous boot
journalctl -k                              # kernel messages (replaces dmesg)
journalctl --disk-usage
```

</details>

## Persistence Gotcha ⚠️

```text
Storage=volatile (default on many distros) → RAM only → LOST on reboot
Storage=persistent → /var/log/journal → survives reboot
```

```bash
sudo mkdir -p /var/log/journal
sudo systemd-tmpfiles --create --prefix /var/log/journal
# set Storage=persistent in /etc/systemd/journald.conf, then restart journald
```

## Easy Memory Flow

```text
Service logs something
      ↓
journald indexes it with structured fields
      ↓
journalctl -u <unit> -p err --since ... filters precisely
      ↓
(if Storage=volatile) reboot → gone, unless persistent
```

## One-Line Memory Tricks

* **journald logs are structured key=value, not flat text — that's why filtering is fast and precise.**
* **Volatile by default on many distros — a crash-then-reboot can silently erase the evidence.**
* **`-p err` filters priority; `-u` filters by unit; combine them.**

---

# 6. rsyslog

## Core Idea ⭐

> **journald is local and often volatile; rsyslog is how logs get shipped long-term or off-box.**

They're not competitors — rsyslog commonly reads *from* journald and writes to flat files or a remote aggregator.

---

## Flow

```text
App → syslog() → journald (indexes, local)
                     │
                     ▼ (ImJournal module, default on)
                 rsyslog
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
   /var/log/*.log        remote log server
   (flat files)           (ELK / Loki / Graylog)
```

## Config Basics

```text
# /etc/rsyslog.d/50-default.conf
facility.priority    action
auth,authpriv.*       /var/log/auth.log
*.info;mail.none       /var/log/syslog
*.*                    @@logserver.internal:514   # ship everything, TCP
```

Facilities: `kern, user, mail, daemon, auth, cron, local0-7`
Priorities (rising severity): `debug → info → notice → warning → err → crit → alert → emerg`

<details>
<summary><strong>Core commands</strong></summary>

```bash
sudo systemctl restart rsyslog
logger "test message from $(hostname)"   # generate a test line, verify routing
```

</details>

## Easy Memory Flow

```text
journald has it locally
      ↓
rsyslog reads it (or receives syslog() calls directly)
      ↓
Routes by facility.priority rule
      ↓
Writes to a flat file AND/OR forwards to a remote aggregator
```

## One-Line Memory Tricks

* **journald = local, structured, often volatile. rsyslog = routing + long-term + remote shipping.**
* **Multi-node systems centralize via rsyslog because grepping N individual journalctl sessions doesn't scale.**
* **`@@host:port` = TCP forward; single `@` = UDP (no delivery guarantee).**

---

# 7. firewalld

## Core Idea ⭐

> **Zone-based frontend over nftables — think in zones and services, not raw rules.**

You assign interfaces to zones (`public`, `internal`, `trusted`...) and each zone has an allowed services/ports policy.

---

## Flow

```text
Packet arrives on interface
      ↓
Interface belongs to a zone (public / internal / trusted / dmz...)
      ↓
Zone's allowed services/ports/rich-rules evaluated
      ↓
Accept / Reject / Drop
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
firewall-cmd --get-active-zones
firewall-cmd --zone=public --list-all

firewall-cmd --zone=public --add-port=8080/tcp              # runtime only
firewall-cmd --zone=public --add-port=8080/tcp --permanent  # persists, needs --reload
firewall-cmd --reload

firewall-cmd --zone=public --add-service=https --permanent
firewall-cmd --zone=internal --change-interface=eth1 --permanent
```

</details>

## The #1 Gotcha ⚠️

```text
--permanent alone           → saved, but NOT active until --reload
(no --permanent)            → active now, but LOST on reboot/reload
--permanent + --reload      → the only combination that's both active and durable
```

## Easy Memory Flow

```text
Interface → Zone → Rules (services/ports/rich-rules) → Verdict
                                  ↓
                    firewalld translates all of this into nftables underneath
```

## One-Line Memory Tricks

* **Zones = "what's allowed from this interface", not per-rule micromanagement.**
* **`--permanent` without `--reload` = silently not applied. Always pair them (or test runtime first, then make permanent).**
* **Rich rules exist for anything zones/services can't express (specific source IP + port combos).**

---

# 8. nftables

## Core Idea ⭐

> **Tables → chains → rules — the actual packet-filtering engine firewalld compiles down to.**

You reach for this directly when you need precision beyond zones, or aren't running firewalld at all.

---

## Flow

```text
Packet enters kernel netfilter hooks
      ↓
Matches a chain (input / forward / output) at a given hook + priority
      ↓
Rules evaluated top to bottom within the chain
      ↓
First matching verdict wins: accept / drop / continue
      ↓
Chain's default policy applies if nothing matched
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
nft list ruleset                     # see everything, live
nft add table inet filter
nft add chain inet filter input { type filter hook input priority 0 \; policy drop \; }
nft add rule inet filter input ct state established,related accept
nft add rule inet filter input tcp dport 22 accept

nft list ruleset -a                  # show handles (needed to delete)
nft delete rule inet filter input handle 4
nft flush ruleset                    # wipe everything — careful over SSH!
```

</details>

## Easy Memory Flow

```text
table (namespace, e.g. inet filter)
      ↓
chain (hook point: input/forward/output, + priority + default policy)
      ↓
rule (match conditions + verdict)
      ↓
first match wins, else default policy
```

## One-Line Memory Tricks

* **`inet` family = IPv4 + IPv6 in one table — real improvement over iptables needing two separate rule sets.**
* **Rules are deleted by handle, not by content — `nft list ruleset -a` first.**
* **`ct state established,related accept` early in the chain = don't re-evaluate already-tracked connections against every rule.**
* **firewalld is the friendly frontend; nftables is what's actually enforcing the packet decisions.**

---

# 9. Package Management

## Core Idea ⭐

> **A trusted, versioned, dependency-aware way to install software — not manual downloads.**

Repo metadata tells the manager what versions exist and what each package depends on, so installs/upgrades/removals stay consistent.

---

## Flow

```text
apt update / dnf check-update
      ↓
Fetches repo metadata (what's available, dependency graph)
      ↓
install <pkg>
      ↓
Resolve dependencies → download → verify signature → unpack → run post-install hooks
```

## APT vs DNF — Same Concepts, Different Syntax

| Concept | APT (Debian/Ubuntu) | DNF (RHEL family) |
|---|---|---|
| Refresh metadata | `apt update` | `dnf check-update` |
| Install | `apt install pkg` | `dnf install pkg` |
| Remove (keep config) | `apt remove pkg` | `dnf remove pkg` |
| Remove (+ config) | `apt purge pkg` | — |
| List installed | `dpkg -l` | `dnf list installed` |
| Which package owns this file | `dpkg -S /path` | `dnf provides /path` |
| Pin/hold a version | `apt-mark hold pkg` | `dnf versionlock add pkg` |
| Undo last transaction | — (no native undo) | `dnf history undo <id>` |

<details>
<summary><strong>Why installing a package doesn't start its service</strong></summary>

Shipping a `.service` file with a package does **not** enable/start it automatically on most distros — deliberate packaging policy. Always follow with:
```bash
sudo systemctl enable --now <service>
```
This is the classic "package installed fine, nothing's running" trap.

</details>

## Easy Memory Flow

```text
Repo metadata refreshed
      ↓
Dependency graph resolved
      ↓
Package downloaded + signature verified
      ↓
Installed, files registered (so it can be cleanly removed later)
      ↓
(you) still need to systemctl enable --now it, usually
```

## One-Line Memory Tricks

* **Package managers track *what's installed and why* — manual `./configure && make install` gives you none of that.**
* **`apt`/`dnf` resolve dependencies automatically; that's the entire value proposition over manual installs.**
* **Installing ≠ running — always check `systemctl status` after installing a service package.**
* **`dpkg -S` / `dnf provides` — reverse lookup, "which package owns this file" — great for debugging.**

---

# 10. Git

## Core Idea ⭐

> **Git tracks snapshots (commits), not diffs — and every object is content-addressed by hash.**

A commit is a full snapshot pointer, not a delta — diffs are *computed* between snapshots on demand, not stored directly.

---

## Flow

```text
Working directory (your files, edited)
      ↓ git add
Staging area / index (what WILL be in the next commit)
      ↓ git commit
Local repo (.git) — permanent snapshot, linked to previous commit
      ↓ git push
Remote repo (GitHub/GitLab) — shared history
```

## Branching & Merging

```text
main:     A---B---C-------F
                   \      /
feature:            D----E

C = branch point (git checkout -b feature)
D, E = commits on feature branch, isolated from main
F = merge commit, brings feature's changes into main
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
git init / git clone <url>
git status                          # what's staged/unstaged/untracked
git add <file>                      # stage
git commit -m "message"             # snapshot
git push / git pull                 # sync with remote
git branch feature-x
git checkout feature-x              # (or: git switch feature-x)
git merge feature-x                 # bring branch into current one
git log --oneline --graph           # visual history
git diff                            # unstaged changes
git diff --staged                   # staged changes
```

</details>

<details>
<summary><strong>Merge conflicts — what's actually happening</strong></summary>

Git can auto-merge changes to *different* parts of a file. A conflict means two branches changed the **same lines** — Git can't guess which is correct, so it marks both versions inline (`<<<<<<<`, `=======`, `>>>>>>>`) and waits for a human to pick/combine, then `git add` the resolved file and commit.

</details>

## Why This Matters for GitOps

```text
Git commit = desired state, reviewed and versioned
      ↓
GitOps agent (e.g. ArgoCD) watches the repo
      ↓
Detects drift between Git and the live cluster
      ↓
Reconciles live state to match Git — automatically
```
This is the *same* reconcile-loop pattern as a Kubernetes Operator (§11 of the prerequisites guide) — just watching Git instead of a Custom Resource.

## Easy Memory Flow

```text
Edit file
      ↓
git add (stage)
      ↓
git commit (permanent snapshot, local)
      ↓
git push (sync to remote)
      ↓
Others git pull to get your changes
      ↓
Conflicts only happen when the SAME lines changed on both sides
```

## One-Line Memory Tricks

* **Commits are snapshots, not diffs — Git computes diffs on demand for display.**
* **Staging area = "what will be in the next commit," separate from the working directory.**
* **A merge conflict = same lines changed on both branches; anything else auto-merges.**
* **GitOps only works because Git already gives you version history + review — it's the substrate, not a separate invention.**