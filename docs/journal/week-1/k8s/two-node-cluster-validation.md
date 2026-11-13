# Validate the Two Node Cluster

## 1. Basic cluster health

```bash
kubectl get nodes -o wide
```
Expect all nodes (`control-plane` + `worker-01`, etc.) listed as `Ready`. If a node shows `NotReady`, it's almost always the CNI (Flannel) not up yet — give it a minute, then re-check.

```bash
kubectl cluster-info
```
Confirms the API server and DNS are responding.

## 2. Control plane components

```bash
kubectl get pods -n kube-system -o wide
```
Look for: `etcd-*`, `kube-apiserver-*`, `kube-controller-manager-*`, `kube-scheduler-*`, `kube-proxy-*` (one per node), and `kube-flannel-ds-*` (one per node) — all should be `Running`, `1/1` or `2/2` ready.

## 3. Verify each node actually joined correctly

```bash
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.nodeInfo.kubeletVersion}{"\n"}{end}'
```
Confirms kubelet version matches `kubernetes_version` from `group_vars/all.yml` on every node — catches a partial/mismatched join.

## 4. On each node individually (via Ansible ad-hoc, no need to SSH one by one)

```bash
ansible kubernetes -m shell -a "systemctl is-active containerd kubelet"
```
Both should report `active` on every host in one shot.

```bash
ansible kubernetes -m shell -a "swapon --show"
```
Should return **empty** on every node — confirms the swap-disable task actually took.

## 5. Networking sanity check (pod-to-pod across nodes)

```bash
kubectl run test-a --image=busybox --restart=Never -- sleep 3600
kubectl run test-b --image=busybox --restart=Never -- sleep 3600
kubectl wait --for=condition=Ready pod/test-a pod/test-b --timeout=60s
kubectl exec test-a -- ping -c 3 $(kubectl get pod test-b -o jsonpath='{.status.podIP}')
```
If this fails, Flannel's VXLAN traffic is likely being blocked — check the `allow_flannel_vxlan` security group rule (UDP 8472) actually applied.

```bash
kubectl delete pod test-a test-b
```

## 6. DNS resolution inside the cluster

```bash
kubectl run dns-test --image=busybox --restart=Never -- nslookup kubernetes.default
kubectl logs dns-test
kubectl delete pod dns-test
```
Confirms CoreDNS is working — needed for any real workload to resolve services.

## 7. Confirm your Terraform-provisioned pieces landed correctly

```bash
ansible kubernetes -m shell -a "cat /etc/resolv.conf"      # should show 8.8.8.8/8.8.4.4 via DHCP, not a static override
ansible kubernetes -m shell -a "cat /etc/hosts"            # should list every cluster node by private IP
```

If everything above comes back clean, the cluster is genuinely healthy — not just "kubeadm exited 0."