[control_plane]
%{ for name, ip in control_plane ~}
${name} ansible_host=${bastion_ip} private_ip=${ip}
%{ endfor ~}

[workers]
%{ for name, ip in workers ~}
${name} ansible_host=${ip} private_ip=${ip} ansible_ssh_common_args='-o ProxyCommand="ssh -i ~/.ssh/id_ed25519 -o StrictHostKeyChecking=no -W %h:%p -q ubuntu@${bastion_ip}"'
%{ endfor ~}

[kubernetes:children]
control_plane
workers
