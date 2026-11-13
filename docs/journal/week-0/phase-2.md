# Phase 2 — Systems Layer

> Core Idea first, flow diagrams, key files/commands, then memory tricks. Same format as Phase 1 — each section re-readable in 5 minutes before an interview.

## Table of Contents
1. [L2 vs L3](#1-l2-vs-l3)
2. [Routing](#2-routing)
3. [NAT](#3-nat)
4. [VLANs](#4-vlans)
5. [VXLAN](#5-vxlan)
6. [Linux Bridges](#6-linux-bridges)
7. [Open vSwitch](#7-open-vswitch)
8. [KVM & QEMU](#8-kvm--qemu)
9. [libvirt](#9-libvirt)
10. [Disk Images: qcow2 vs raw](#10-disk-images-qcow2-vs-raw)
11. [cloud-init](#11-cloud-init)

---

# 1. L2 vs L3

## Core Idea ⭐

> **L2 moves frames by MAC address within one broadcast domain. L3 moves packets by IP address across different ones — and only a router bridges the two.**

A switch never looks at an IP address to forward a frame; a router never forwards without consulting a routing table.

---

## Flow

```text
Same L2 segment (VLAN 10):
  VM1 (MAC aa:aa, 10.0.1.10)  <--ARP + switching-->  VM2 (MAC bb:bb, 10.0.1.11)
  No router involved. No routing table consulted.

Different L2 segments:
  VM1 (VLAN 10) ---> Router (L3) ---> VM3 (VLAN 20)
  MAC changes at every hop; IP stays the same end-to-end.
```

## ARP — the glue

```text
Host A wants to send to 10.0.1.11 (same segment)
      ↓
Broadcast: "who has 10.0.1.11?"
      ↓
Owner replies with its MAC
      ↓
Cached in ip neigh (ARP table)
```

```bash
ip neigh show                     # local ARP cache
arping -I eth0 10.0.1.11          # manually trigger ARP resolution
```

## Easy Memory Flow

```text
Same segment?  → L2, MAC addresses, switch, ARP, no router
Different segment? → L3, IP addresses, router, routing table
```

## One-Line Memory Tricks

* **Broadcast domain = L2 segment. ARP/DHCP broadcasts never cross a router.**
* **Switches learn MACs; routers consult routing tables. Different jobs, different layers.**
* **VLANs exist specifically to carve one wire into multiple L2 broadcast domains.**

---

# 2. Routing

## Core Idea ⭐

> **Every host — not just dedicated routers — consults a routing table for every outbound packet: longest-prefix-match wins.**

---

## Flow

```text
Packet needs to leave
      ↓
Kernel checks routing table for best-matching destination
      ↓
More specific route wins (10.0.2.0/24 beats 10.0.0.0/16 for a 10.0.2.x destination)
      ↓
Forward via matched interface/next-hop, or via "default" if nothing else matches
```

## Reading a Table

```text
ip route show
default via 192.168.1.1 dev eth0                          # gateway of last resort
10.0.1.0/24 dev eth1 proto kernel scope link src 10.0.1.5  # directly connected, no gateway
10.0.2.0/24 via 10.0.1.1 dev eth1                          # reach via this next-hop
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
ip route add 10.0.2.0/24 via 10.0.1.1 dev eth1
ip route del 10.0.2.0/24
ip route get 10.0.2.5              # "which route would actually be used" — best first debugging step

sysctl -w net.ipv4.ip_forward=1    # turn a Linux host into an actual router
```

</details>

## The #1 Gotcha ⚠️

```text
Perfect routing tables + perfect firewall rules
      ↓
STILL won't route between two interfaces
      ↓
unless net.ipv4.ip_forward=1
```

## Easy Memory Flow

```text
Packet → routing table lookup → longest prefix match wins → forward via that interface/next-hop
                                                                      ↓
                                                    (only works between interfaces if ip_forward=1)
```

## One-Line Memory Tricks

* **Longest prefix match always wins — the most specific route, not the first one listed.**
* **`ip route get <dest>` is the fastest way to stop guessing and see the real answer.**
* **`ip_forward=0` (default on most hosts) silently blocks routing between interfaces — check this before debugging anything fancier.**
* **Static routing = you declare it. Dynamic (BGP/OSPF) = routers exchange it automatically — out of scope for host-level admin, but datacenter fabrics run on it.**

---

# 3. NAT

## Core Idea ⭐

> **NAT rewrites source and/or destination addresses as packets cross a boundary — SNAT shares one public IP outbound, DNAT exposes one internal service inbound.**

---

## SNAT / Masquerade — Outbound Sharing

```text
Private hosts (10.0.1.0/24) ---> [masquerade rewrites src IP] ---> Internet
                                   using the gateway's own outbound IP
```

```
table ip nat {
    chain postrouting {
        type nat hook postrouting priority 100;
        ip saddr 10.0.1.0/24 oif "eth0" masquerade
    }
}
```

`masquerade` auto-detects the current outbound IP (handles DHCP-assigned/dynamic IPs); plain `snat to <ip>` is for a fixed IP.

## DNAT — Inbound / Floating IPs

```text
Internet ---> [public IP:2222] ---> DNAT rewrites dst ---> 10.0.1.10:22 (private VM)
```

```
table ip nat {
    chain prerouting {
        type nat hook prerouting priority -100;
        tcp dport 2222 dnat to 10.0.1.10:22
    }
}
```

**This is exactly what a Neutron floating IP is** — a public IP DNATed to an instance's private IP, with a matching SNAT rule for return traffic.

<details>
<summary><strong>conntrack — how the kernel remembers translations</strong></summary>

```bash
conntrack -L                       # list tracked connections + NAT state
conntrack -L -p tcp --dport 22
```
Return traffic for a NAT'd connection gets un-translated automatically because the kernel tracks it — this is also why `ct state established,related accept` is standard in firewall rules: trust already-tracked traffic instead of re-evaluating it.

</details>

## Easy Memory Flow

```text
Outbound sharing?  → SNAT/masquerade, rewrite SOURCE
Inbound exposure?  → DNAT, rewrite DESTINATION
Either way          → needs ip_forward=1 to actually work
```

## One-Line Memory Tricks

* **SNAT changes where a packet appears to come FROM. DNAT changes where it's actually going TO.**
* **Masquerade = SNAT that auto-adapts to a dynamic outbound IP — the right choice for a gateway.**
* **A floating IP is DNAT + matching SNAT, nothing more exotic than that.**
* **NAT without `ip_forward=1` does nothing — the packet needs to be routed through, not just rewritten.**

---

# 4. VLANs

## Core Idea ⭐

> **802.1Q tags one physical link into multiple isolated broadcast domains — up to 4094 usable IDs.**

Access ports = untagged, one VLAN. Trunk ports = tagged, multiple VLANs over one link (switch-to-switch, switch-to-hypervisor).

---

## Tag Anatomy

```text
Normal frame:   [Dest MAC][Src MAC][EtherType][Payload]
802.1Q tagged:  [Dest MAC][Src MAC][0x8100][VLAN ID - 12 bit][EtherType][Payload]
```
12-bit ID = 4096 values, 0 and 4095 reserved → 4094 usable.

<details>
<summary><strong>Core commands</strong></summary>

```bash
ip link add link eth0 name eth0.10 type vlan id 10
ip link set eth0.10 up
ip addr add 10.0.10.5/24 dev eth0.10
ip -d link show eth0.10             # -d shows the VLAN id
```
Physical switch port must be configured as a **trunk** allowing VLAN 10, or nothing arrives.

</details>

## Easy Memory Flow

```text
Tag inserted into frame → switch/host reads tag → routes to correct broadcast domain
                                                          ↓
                                         Access port = 1 VLAN, untagged
                                         Trunk port = many VLANs, tagged
```

## One-Line Memory Tricks

* **4094 usable VLAN IDs — that ceiling is the entire reason VXLAN exists.**
* **VLANs don't cross an L3 boundary — can't stretch one across routed networks without extra tricks.**
* **Trunk ports carry multiple tagged VLANs; access ports carry exactly one, untagged.**

---

# 5. VXLAN

## Core Idea ⭐

> **"L2 over L3 over UDP" — wraps a whole Ethernet frame inside a UDP packet, letting an L2 segment span any routed IP network. ~16 million VNIs vs VLAN's 4094.**

---

## Encapsulation

```text
[Outer IP][Outer UDP, dport 4789][VXLAN header: 24-bit VNI][Original Ethernet frame]
                                                                      ↑
                                                    this is the tenant's actual L2 traffic
```

Any two hosts that can route IP to each other can now share an L2 segment — no physical switch config needed, unlike VLAN trunking.

## VTEP — the endpoint doing the work

```text
Hypervisor A (VTEP) --- encap/decap ---> IP network ---> Hypervisor B (VTEP) --- decap ---> VM
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
ip link add vxlan10 type vxlan id 10 dstport 4789 remote 10.0.0.20 dev eth0
ip link set vxlan10 up
ip addr add 172.16.0.1/24 dev vxlan10
```
`id 10` = VNI. `remote` = the other VTEP's IP. (OVS usually programs this for you in a real OpenStack deployment.)

</details>

## VLAN vs VXLAN

| | VLAN | VXLAN |
|---|---|---|
| ID space | 4094 | ~16 million |
| Crosses L3? | No | Yes |
| Needs physical switch config | Yes (trunk ports) | No — just IP reachability |
| Typical OpenStack use | Provider networks | Tenant (self-service) overlay networks |

## Easy Memory Flow

```text
Original Ethernet frame
      ↓ wrapped inside
UDP packet (dport 4789) between two VTEP IPs
      ↓ travels across
Any ordinary routed IP network
      ↓ unwrapped at
Destination VTEP → delivered as if it never left L2
```

## One-Line Memory Tricks

* **VXLAN = Ethernet frame as UDP payload. The "overlay" is really just tunneling.**
* **VNI (24-bit) is VXLAN's VLAN-ID equivalent — vastly bigger ID space.**
* **MTU matters: VXLAN adds ~50 bytes of overhead. Undersized physical MTU = silent fragmentation/drops — a classic "tenant network randomly broken" root cause.**

---

# 6. Linux Bridges

## Core Idea ⭐

> **A software L2 switch built into the kernel — any interface you attach behaves as if plugged into the same physical switch.**

---

## Flow

```text
ip link add br0 type bridge
      ↓
ip link set eth1 master br0     (physical NIC joins)
ip link set tap0 master br0     (VM's virtual NIC joins)
      ↓
br0 now forwards frames between them like a real switch,
learning MAC → port mappings as traffic flows
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
ip link add name br0 type bridge
ip link set br0 up
ip link set eth1 master br0

bridge link show                 # port membership
bridge fdb show br br0           # forwarding database (learned MAC -> port)
ip link set br0 type bridge stp_state 1   # enable spanning tree if redundant paths exist
```

</details>

## Why a VM's Networking Actually Works

```text
KVM VM's virtual NIC ---> creates a tap device on the host ---> tap device joins a bridge
                                                                        ↓
                                                    same L2 segment as whatever else is on that bridge
```

## Easy Memory Flow

```text
Create bridge → attach interfaces (master br0) → bridge learns MACs → forwards like hardware switch
```

## One-Line Memory Tricks

* **A Linux bridge is a real (software) L2 switch — same MAC-learning behavior as hardware.**
* **This is literally how a KVM VM gets network connectivity: its tap device joins a bridge alongside the host's uplink.**
* **STP matters once there's more than one path between segments — otherwise optional for a simple lab.**
* **Linux bridges are simple and kernel-native; OVS is the programmable superset — same mental model, far more capability.**

---

# 7. Open vSwitch

## Core Idea ⭐

> **A programmable virtual switch with OpenFlow flow tables, built-in VXLAN/GRE tunnel support, and a central config database (OVSDB) — the actual dataplane Neutron automates.**

---

## Components

```text
ovs-vsctl    → configure topology (bridges, ports) — talks to OVSDB
ovs-ofctl    → inspect/program OpenFlow flow tables directly
ovs-vswitchd → the dataplane daemon actually forwarding traffic
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
ovs-vsctl add-br br-int
ovs-vsctl add-port br-int tap0
ovs-vsctl show                      # full topology
ovs-vsctl list-ports br-int
ovs-vsctl del-port br-int eth1

ovs-ofctl show br-int               # ports as OpenFlow sees them
ovs-ofctl dump-flows br-int         # every match-action rule currently programmed
ovs-ofctl add-flow br-int "priority=100,in_port=1,actions=output:2"
```

</details>

## VXLAN Port, Native to OVS

```bash
ovs-vsctl add-port br-tun vxlan0 -- set interface vxlan0 type=vxlan options:remote_ip=10.0.0.20 options:key=10
```

## The br-int / br-tun / br-ex Pattern

```text
┌──────────┐  patch port  ┌──────────┐  vxlan port  ┌────────────────────┐
│  br-int   │──────────────│  br-tun   │───────────────│ other hypervisors  │
│ (VM taps, │               │ (encap/   │               │  via VTEP IPs      │
│  sec grps)│               │  decap)   │               └────────────────────┘
└──────────┘               └──────────┘
      │
 ┌──────────┐
 │  br-ex    │  external/provider network — floating IP traffic exits here
 └──────────┘
```

## Easy Memory Flow

```text
VM tap port attaches to br-int
      ↓
br-int patched to br-tun for cross-hypervisor (VXLAN) traffic
      ↓
br-ex bridges to the physical external network for floating IPs
      ↓
Every flow decision (allow, drop, VLAN-tag rewrite) is a rule in ovs-ofctl dump-flows
```

## One-Line Memory Tricks

* **"Software-defined" means exactly this: flow tables (match + priority + action) instead of pure MAC learning.**
* **Neutron's OVS agent is, at its core, a program writing `ovs-vsctl`/flow rules for you.**
* **`br-int` = where every VM's tap port lives. `br-tun` = tunnel encap/decap. `br-ex` = external/floating IP egress.**
* **Patch ports connect OVS bridges to each other in software — no physical cable needed.**

---

# 8. KVM & QEMU

## Core Idea ⭐

> **KVM turns the kernel into a hypervisor using hardware virtualization extensions (VT-x/AMD-V). QEMU emulates everything else (disks, NICs, BIOS) and hands CPU/memory work to KVM for near-native speed.**

QEMU *without* KVM is a full software emulator — dramatically slower, mainly useful for cross-architecture emulation.

---

## Flow

```text
Guest instruction
      ↓
Most instructions: executed directly on physical CPU via KVM (near-native speed)
      ↓
Device I/O (disk, NIC, BIOS): emulated by QEMU in userspace
```

<details>
<summary><strong>Checking support & running directly</strong></summary>

```bash
egrep -c '(vmx|svm)' /proc/cpuinfo   # >0 = CPU supports hardware virtualization
lsmod | grep kvm                     # kvm_intel or kvm_amd loaded?
ls -la /dev/kvm

qemu-system-x86_64 -enable-kvm -m 2048 -smp 2 \
  -drive file=vm1.qcow2,format=qcow2 \
  -netdev bridge,id=net0,br=br-int -device virtio-net-pci,netdev=net0 \
  -vnc :1
```
Every flag here is what libvirt XML expresses declaratively instead — libvirt exists so you don't hand-manage QEMU command lines.

</details>

## Easy Memory Flow

```text
KVM = kernel module, hardware-accelerated CPU/memory virtualization
QEMU = userspace program, emulates the rest of the "hardware"
Together = QEMU/KVM, the pairing almost everyone actually means by "KVM"
```

## One-Line Memory Tricks

* **KVM alone isn't a hypervisor you interact with directly — QEMU is the program that uses it.**
* **No KVM = full emulation = slow. This distinction explains most "why is my VM so slow" questions.**
* **Nested virtualization (KVM inside a VM) needs an explicit `nested=1` module parameter — not on by default.**

---

# 9. libvirt

## Core Idea ⭐

> **A management API/daemon on top of QEMU/KVM giving one stable, declarative (XML-based) interface for defining and controlling VMs — exactly the layer Nova talks to on every compute node.**

---

## Flow

```text
virsh command / Nova
      ↓
libvirtd (or virtqemud)
      ↓
Translates to QEMU invocation
      ↓
QEMU/KVM actually runs the VM
```

<details>
<summary><strong>Core commands</strong></summary>

```bash
virsh list --all                  # all domains (VMs), running or not
virsh start vm1 / virsh shutdown vm1 / virsh destroy vm1   # graceful vs hard stop
virsh dumpxml vm1                 # full domain definition — debug ground truth
virsh edit vm1                    # live-edit XML, validated before applying
virsh console vm1                 # attach to serial console
virsh domifaddr vm1               # get the VM's IP

virsh snapshot-create-as vm1 snap1
virsh snapshot-revert vm1 snap1

virt-install --name vm1 --memory 2048 --vcpus 2 \
  --disk path=vm1.qcow2 --network bridge=br-int \
  --os-variant ubuntu22.04 --import
```

</details>

## Domain XML — the Declarative Core

```xml
<domain type='kvm'>
  <name>vm1</name>
  <memory unit='MiB'>2048</memory>
  <vcpu>2</vcpu>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/var/lib/libvirt/images/vm1.qcow2'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='bridge'><source bridge='br-int'/></interface>
  </devices>
</domain>
```

## Easy Memory Flow

```text
Define desired VM (XML, or virt-install flags)
      ↓
libvirt talks to QEMU/KVM to make it real
      ↓
virsh = the CLI for start/stop/inspect/snapshot
      ↓
Nova generates XML like this on the fly for every instance it boots
```

## One-Line Memory Tricks

* **libvirt exists so nobody hand-manages QEMU command lines — declare the VM, let libvirt drive QEMU.**
* **`virsh dumpxml <instance-name>` on a real compute node is genuinely how you debug "why won't this instance boot."**
* **`virt-manager` is a GUI wrapper over the same libvirt operations — nothing it does bypasses virsh/XML.**

---

# 10. Disk Images: qcow2 vs raw

## Core Idea ⭐

> **raw = simplest, fastest, no extra features. qcow2 = thin-provisioned, snapshot-capable, supports backing files (cheap copy-on-write derived images).**

---

## Comparison

| | raw | qcow2 |
|---|---|---|
| Format | byte-for-byte disk image | QEMU's own copy-on-write format |
| Snapshots | No | Yes (internal) |
| Thin provisioning | Only if filesystem supports sparse files | Native |
| Backing files (derive from a base image) | No | Yes |
| Raw I/O performance | Best (no translation layer) | Slightly lower (COW/metadata overhead) |

<details>
<summary><strong>Core commands</strong></summary>

```bash
qemu-img create -f raw disk.img 20G
qemu-img create -f qcow2 disk.qcow2 20G
qemu-img info disk.qcow2                                        # format, virtual vs actual size, backing file
qemu-img create -f qcow2 -F qcow2 -b base.qcow2 derived.qcow2   # copy-on-write derived image
qemu-img convert -f raw -O qcow2 disk.img disk.qcow2            # format conversion
qemu-img resize disk.qcow2 +10G                                 # grow (guest still needs to grow its FS)
```

</details>

## Easy Memory Flow

```text
One golden qcow2 base image
      ↓  (backing file reference)
Many derived images, each storing ONLY its own diffs
      ↓
Cheap to create thousands of VMs without copying the full disk every time
```

## One-Line Memory Tricks

* **qcow2's snapshot + backing-file support is *why* "spin up from a golden image" is cheap and fast.**
* **raw wins on pure I/O throughput — some high-performance/Ceph-backed deployments skip qcow2 entirely for this reason.**
* **This maps straight to Glance: an image's `disk_format` metadata drives how Nova's libvirt driver provisions the disk.**

---

# 11. cloud-init

## Core Idea ⭐

> **First-boot instance customization driven by metadata the platform hands to the guest — not baked into the image. One golden image becomes thousands of differently-configured instances.**

---

## Datasources

```text
NoCloud (local lab):
  Seed ISO (cidata volume) attached as virtual CD-ROM
      ↓
  cloud-init reads user-data + meta-data from it on boot

OpenStack / EC2-style (real deployments):
  Guest queries http://169.254.169.254 (metadata service) over the network
      ↓
  Neutron's metadata agent proxies this back to Nova, which serves the instance's actual data
```

<details>
<summary><strong>Building a NoCloud seed for a lab VM</strong></summary>

```yaml
# user-data
#cloud-config
hostname: vm1
users:
  - name: kazi
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - ssh-ed25519 AAAA... kazi@laptop
packages: [curl, net-tools]
runcmd:
  - echo "cloud-init ran" > /tmp/marker
```
```yaml
# meta-data
instance-id: vm1
local-hostname: vm1
```
```bash
cloud-localds seed.iso user-data meta-data
# attach seed.iso as a CD-ROM device to the VM — cloud-init auto-detects the cidata label
```

</details>

<details>
<summary><strong>Debugging</strong></summary>

```bash
cloud-init status --long           # inside the guest
cloud-init analyze show            # timing breakdown of what ran and how long
cat /var/log/cloud-init-output.log
```

</details>

## Easy Memory Flow

```text
Golden image (no per-instance config baked in)
      ↓
Boot + datasource provides user-data/meta-data
      ↓
cloud-init sets hostname, injects SSH keys, installs packages, runs scripts
      ↓
Instance is fully configured — no rebuild of the image needed
```

## One-Line Memory Tricks

* **cloud-init separates "what the image is" from "what this specific instance becomes" — that separation is the entire point.**
* **NoCloud = local ISO datasource for labs. OpenStack metadata service = the real Nova/Neutron path via `169.254.169.254`.**
* **Nova doesn't interpret your cloud-config YAML — it just hands it to the metadata service for the guest's own cloud-init agent to consume.**