# How to create Kubernetes with KubeSpray

## Update inventory.ini file in build/kubespray-configs

Format: <hostname> ansible_host=<node-ip> ansible_user=<node-user>

### example:
```text
[all]
node1 ansible_host=128.105.144.109 ansible_user=woojoong

[kube_control_plane]
node1

[etcd]
node1

[kube_node]
node1

[calico_rr]

[k8s_cluster:children]
kube_control_plane
kube_node
calico_rr

```

## Run Makefile target

```bash
$ make deploy-kubernetes
```