hostname: ${hostname}
fqdn: ${hostname}
manage_etc_hosts: false

write_files:
  - path: /etc/hosts
    content: |
      127.0.0.1 localhost
      127.0.1.1 ${hostname}
      %{ for name, ip in cluster_hosts ~}
      ${ip} ${name}
      %{ endfor ~}
  - path: /etc/modules-load.d/k8s.conf
    content: |
      overlay
      br_netfilter
  - path: /etc/sysctl.d/k8s.conf
    content: |
      net.bridge.bridge-nf-call-iptables = 1
      net.bridge.bridge-nf-call-ip6tables = 1
      net.ipv4.ip_forward = 1

runcmd:
  - swapoff -a
  - sed -i '/ swap / s/^/#/' /etc/fstab
  - modprobe overlay
  - modprobe br_netfilter
  - sysctl --system
