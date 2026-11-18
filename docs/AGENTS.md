# docs/ — Agent Guide

MkDocs Material site (`mkdocs.yml` at the repo root). Read this file instead of exploring `docs/`; open an existing page only when you need to edit it or match a specific section.

## Layout

```
docs/
  index.md, getting-started.md, api-reference.md (+ openapi.yaml)
  architecture/        index.md, workspace.dsl (Structurizr C4), diagrams/ (generated SVG; `make diagrams`)
  design-decisions/    index.md + one page per numbered decision
  components/          index.md + one page per deployable part (operator, api, frontend, auth, sdk, delivery, observability, infrastructure)
  notes/               learning notes, organised by topic (see map below)
  notes/images/        local images for notes
  assets/              zoom.js / zoom.css (click-to-zoom for images and Mermaid); don't touch
```

Every page must be listed in `nav:` in `mkdocs.yml`, or it's built but unreachable.

## Learning notes: topic map

Check here before writing, so a concept goes into its existing page instead of a duplicate.

| Nav group | Page | Owns |
|---|---|---|
| Linux | `linux-fundamentals.md` | Kernel/syscalls, distros, cloud-init, shell, pipes, FHS, everyday commands |
| | `linux-users-permissions.md` | Users/groups, sudo, rwx/numeric modes, umask, setuid/setgid/sticky, capabilities, securityContext mapping |
| | `linux-processes-services.md` | Processes, signals/graceful shutdown, nice, free/df/du, namespaces+cgroups+OOM, systemd, packages, cron |
| | `linux-networking-security.md` | ip/ss, bind addresses, connectivity debugging, netfilter/iptables/nftables, ufw/firewalld, SSH, hardening |
| | `linux-logs-troubleshooting.md` | journald/rsyslog, journalctl, grep/awk, dmesg, troubleshooting method, decision trees, tar/rsync backups |
| Infrastructure | `openstack.md`, `terraform.md`, `kubernetes-on-vms.md` | IaaS, Terraform state, Ansible + kubeadm node setup |
| Kubernetes | `kubernetes-concepts.md` | Overview/map: component table, API-server-as-hub, control loops, end-to-end flow, component identities |
| | `kubernetes-control-plane.md` | etcd, kube-apiserver (pipeline, REST, kubeconfig), scheduler, controller-manager, CCM, HA/leader election |
| | `kubernetes-nodes.md` | kubelet, CRI, containerd/OCI/runc, static Pods, pod sandbox, node tools, TLS bootstrapping |
| | `kubernetes-cluster-networking.md` | Networking model, CNI, overlays/Flannel, Services + kube-proxy, Service types, CoreDNS, NetworkPolicy |
| | `kubernetes-workloads.md` | Pods, labels, workload controllers, rollouts, ConfigMap/Secret (brief), requests/limits, affinity, taints |
| | `kubernetes-storage.md` | Volumes, PV/PVC/StorageClass, access modes, CSI, StatefulSet volumes |
| | `operators.md` | CRD/CR/operator, reconcile loop, controller-runtime, owner refs, finalizers |
| API | `api-design.md` | REST constraints, URLs, methods (safe/idempotent, PUT vs PATCH), status codes, error shape, versioning, pagination, clean architecture |
| | `api-hardening.md` | Idempotency keys, input validation, injection, SQL console limits, CORS, HTTPS rules, rate limiting, caching |
| | `connection-pooling.md`, `health-and-scaling.md`, `go-sdk.md` | |
| Networking | `networking.md` (TCP, DNS, TLS), `exposing-services.md` (Ingress, Gateway API, cert-manager) | |
| Security | `kubernetes-security.md` (RBAC, Secrets, hardened containers), `authentication.md` (OAuth/OIDC/ZITADEL) | |
| Frontend | `frontend.md` | |
| Operations | `gitops.md`, `observability.md` | |

When you add or split a page, update this table.

## Adding a learning note

1. Pick the page from the map. Add a section to it if it fits; create a new page only for a distinct topic (roughly 150–400 lines per page; split rather than grow past that).
2. New page: `docs/notes/<area>-<topic>.md`, kebab-case, prefixed by area (`linux-`, `kubernetes-`).
3. Register it in **both** `mkdocs.yml` `nav:` (under the right `Learning notes` group) and the topic table in `docs/notes/index.md` (`| | [Title](file.md) | one-line summary |`).
4. Link it from its overview page (`kubernetes-concepts.md` or `linux-fundamentals.md` abstract) and from neighbours where relevant.

### Page template

```markdown
# Title in sentence case

!!! abstract "Where it's used in this project"
    One or two sentences tying the topic to real code/config, with GitHub links
    (https://github.com/foyez/dbaas-platform/blob/main/<path>). End with: Overview: [..](..).

## Key ideas

### Subtopic
Short prose, then tables, then a Mermaid diagram or code block. Prefer tables over lists for comparisons.

## What I learned building it

- **Bold one-line lesson.** One or two sentences of explanation, project-specific where possible.

## Review questions

??? question "Question text?"
    Answer, indented four spaces.
```

Other page types: design decisions start `# N. Decision as a sentence`, then **Why:** / **Trade-off:** paragraphs; components describe what it does, how it's built, where the code lives.

## Writing rules

- First person ("I") in "What I learned"; neutral elsewhere. Short sentences, no filler, no emoji in headings.
- **Only claim project facts you've verified** with `grep`/`Read` (file paths, flags, values). If unsure, phrase it generically instead of as experience.
- Source material from courses: rewrite in your own words, keep the concepts, drop install walk-throughs and lab-specific steps. Correct inaccuracies rather than copying them.
- Link with relative paths (`kubernetes-nodes.md#anchor`, `../components/operator.md`). Anchors are GitHub-style lowercase slugs: `## Joining a node: TLS bootstrapping` → `#joining-a-node-tls-bootstrapping`.
- Admonitions available: `!!! abstract|tip|note|warning`, collapsible `??? question`.

## Images and diagrams

- **Mermaid** for anything you can draw: `flowchart`, `sequenceDiagram`. Quote labels containing `/`, `(`, `:` at the start, or other punctuation: `F["/var/log/*.log"]`, `A -->|"text (x)"| B`. Use `<br/>` for line breaks.
- **Official Kubernetes docs images** (CC BY 4.0): download into `docs/notes/images/` and credit:
  `<small>Image: [Kubernetes documentation](<page-url>), CC BY 4.0</small>`
- **Non-openly-licensed images** (e.g. iximiuz Labs course diagrams): hotlink, never copy into the repo, and credit:
  `<small>Image: [Kubernetes the (Very) Hard Way](https://labs.iximiuz.com/courses/kubernetes-the-very-hard-way-0cbfd997), iximiuz Labs</small>`
- Always write descriptive alt text: `![What the diagram shows](url)`.
- Architecture C4 diagrams come from `architecture/workspace.dsl`; edit the DSL and run `make diagrams`, don't hand-edit SVGs.

## Verify before finishing

```bash
make docs-build     # mkdocs --strict in Docker: fails on broken links / missing pages
make docs-check     # renders every Mermaid block with mermaid-cli
```

Docker (OrbStack) must be running. If it isn't, fall back to a throwaway venv in a temp dir:
`python3 -m venv $TMP/venv && $TMP/venv/bin/pip install mkdocs-material==<MKDOCS_MATERIAL_VERSION from Makefile> && NO_MKDOCS_2_WARNING=1 $TMP/venv/bin/mkdocs build --strict -d $TMP/site`.

`--strict` doesn't check `#anchor` links. Check them against the built HTML:

```bash
cd docs/notes && grep -ohE '\]\([a-z./-]+\.md#[a-z0-9-]+\)' <changed>.md | sort -u | while read l; do
  f=$(echo "$l" | sed -E 's/\]\(([^#]+)#.*/\1/'); a=$(echo "$l" | sed -E 's/.*#([^)]+)\)/\1/')
  grep -q "id=\"$a\"" "$SITE/notes/${f%.md}/index.html" || echo "MISSING $f#$a"; done
```

## Commits

Short conventional messages, scoped: `docs(notes): add <topic> note`, `docs(architecture): ...`. One commit per page or logical change; no co-author trailers. Only commit or push when asked.
