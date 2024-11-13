#!/bin/bash

kubectl get secrets eks-connection -n default -o jsonpath='{.data.kubeconfig}' | base64 -d
