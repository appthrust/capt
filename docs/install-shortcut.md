```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.1/cert-manager.yaml
```

```
clusterctl init --addon helm
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

## Create Cluster

```bash
kubectl apply -f config/samples/demo-cluster6/cluster.yaml
```
