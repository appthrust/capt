```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.1/cert-manager.yaml
```

```
clusterctl init --core cluster-api --bootstrap kubeadm --infrastructure capt --control-plane capt
```

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane \
  --namespace crossplane-system \
  --create-namespace crossplane-stable/crossplane 
```


```
kubectl create ns upbound-system
```

## crossplane provider terrafrom config

```bash
kubectl apply -f config/samples/crossplane-terraform-config/aws-creds-secret.yaml

kubectl apply -f config/samples/crossplane-terraform-config/aws-cli-nix-cm.yaml
kubectl apply -f config/samples/crossplane-terraform-config/aws-cli-config.yaml
kubectl apply -f config/samples/crossplane-terraform-config/aws-provider-config.yaml
kubectl apply -f config/samples/crossplane-terraform-config/provider.yaml
```

## Install capt

```
make install
```

## Workspace templates

```bash
kubectl apply -f config/samples/workspacetemplates/eks-controlplane-template.yaml
kubectl apply -f config/samples/workspacetemplates/eks-kubeconfig-template.yaml
kubectl apply -f config/samples/workspacetemplates/spot-role-check.yaml 
kubectl apply -f config/samples/workspacetemplates/spot-role-create.yaml 
kubectl apply -f config/samples/workspacetemplates/vpc.yaml
```

## ClusterClass (Topology) templates

```bash
kubectl apply -f templates/clusterclass/capt-clusterclass.yaml
kubectl apply -f config/samples/clustertopology/controlplanetemplate.yaml
```

## Create Cluster (Topology flavor)

```bash
export CLUSTER_NAME=demo
clusterctl generate cluster $CLUSTER_NAME --flavor topology --infrastructure capt --control-plane capt --target-namespace default > cluster.yaml
kubectl apply -f cluster.yaml
```
