# pve-cluster-autoscaler
Cluster autoscaler workload that scales up/down a cluster on proxmox to automatically respond to resource shortages in the cluster.

## Create a sample configuration
kubectl create configmap autoscaler-config \
    --from-literal=PM_USER=root@pam \
    --from-literal=PM_PASS=password \
    --from-literal=PM_API_URL=https://x.x.x.x:8006/api2/json \
    --from-literal=insecure=false \
    --from-literal=debug=false \
    --from-literal=templateName=template \
    --from-literal=nodeName=my-proxmox-node