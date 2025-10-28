# Image URL to use all building/pushing image targets
VERSION ?= 0.1.0
# If VERSION file exists, override VERSION with its value (format: "VERSION = X.Y.Z")
ifneq ($(wildcard VERSION),)
VERSION := $(shell sed -n 's/^VERSION *= *//p' VERSION)
endif
IMG ?= ghcr.io/appthrust/capt:v$(VERSION)
# Default image used by `make setup` for local dev. Override with SETUP_IMG=<image>.
SETUP_IMG ?= ghcr.io/appthrust/capt:dev
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen
	mkdir -p config/clusterapi/controlplane/bases
	mkdir -p config/rbac
	$(CONTROLLER_GEN) rbac:roleName=manager-role-controlplane paths="./internal/controller/controlplane/..." output:stdout > config/rbac/controlplane-role.yaml
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true webhook paths="./api/controlplane/..." output:crd:artifacts:config=config/clusterapi/controlplane/bases
	mkdir -p config/clusterapi/infrastructure/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role-infrastructure paths="./internal/controller" output:stdout > config/rbac/infrastructure-role.yaml
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true webhook paths="./api/..." output:crd:artifacts:config=config/clusterapi/infrastructure/bases

.PHONY: clusterapi-manifests
clusterapi-manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	mkdir -p config/clusterapi/controlplane/bases
	mkdir -p config/rbac
	$(CONTROLLER_GEN) rbac:roleName=manager-role-controlplane paths="./internal/controller/controlplane/..." output:stdout > config/rbac/controlplane-role.yaml
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true webhook paths="./api/controlplane/..." output:crd:artifacts:config=config/clusterapi/controlplane/bases
	mkdir -p config/clusterapi/infrastructure/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role-infrastructure paths="./internal/controller" output:stdout > config/rbac/infrastructure-role.yaml
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true webhook paths="./api/..." output:crd:artifacts:config=config/clusterapi/infrastructure/bases

.PHONY: clusterctl-setup
clusterctl-setup: clusterapi-manifests kustomize $(KUSTOMIZE_PREREQ) ## Build components and create local config for clusterctl testing.
	# Build kustomize manifests with proper image reference
	mkdir -p capt/infrastructure-capt/v0.0.0
	mkdir -p capt/control-plane-capt/v0.0.0
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	# Add controller args patch for infrastructure components
	cd config/manager && $(KUSTOMIZE) edit add patch --path infrastructure-args-patch.yaml
	$(KUSTOMIZE) build config/clusterapi/infrastructure > capt/infrastructure-capt/v0.0.0/infrastructure-components.yaml
	cd config/manager && $(KUSTOMIZE) edit remove patch --path infrastructure-args-patch.yaml
	# Add controller args patch for control plane components
	cd config/manager && $(KUSTOMIZE) edit add patch --path controlplane-args-patch.yaml
	$(KUSTOMIZE) build config/clusterapi/controlplane > capt/control-plane-capt/v0.0.0/control-plane-components.yaml
	cd config/manager && $(KUSTOMIZE) edit remove patch --path controlplane-args-patch.yaml
	git restore config/manager/kustomization.yaml
	# Copy metadata files
	cp hack/capi/metadata.yaml capt/infrastructure-capt/v0.0.0/metadata.yaml
	cp hack/capi/metadata.yaml capt/control-plane-capt/v0.0.0/metadata.yaml
	# Create config file with current directory
	sed -e 's#%pwd%#'`pwd`'#g' ./hack/capi/config.yaml > capi-local-config.yaml
	@echo "CAPI components ready for clusterctl testing"
	@echo "Use: clusterctl init --infrastructure capt --control-plane capt --config capi-local-config.yaml"

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: clusterapi-manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

# Utilize Kind or modify the e2e tests to load the image locally, enabling compatibility with other vendors.
.PHONY: test-e2e  # Run the e2e tests against a Kind k8s instance that is spun up.
test-e2e:
	go test ./test/e2e/ -v -ginkgo.v

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: clusterapi-manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: clusterapi-manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
# PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name capt-builder
	$(CONTAINER_TOOL) buildx use capt-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm capt-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: clusterapi-manifests $(KUSTOMIZE_PREREQ) ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE_BUILD) config/clusterapi | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: clusterapi-manifests $(KUSTOMIZE_PREREQ) ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE_BUILD) config/clusterapi | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: clusterapi-manifests $(KUSTOMIZE_PREREQ) ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE_BUILD) config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: $(KUSTOMIZE_PREREQ) ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE_BUILD) config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Setup

.PHONY: setup-v1beta1
setup-v1beta1: setup-capi-crds setup-cert-manager setup-capi setup-crossplane setup-provider-terraform setup-webhook-certs ## Legacy v1beta1 bootstrap (pre-installs v1beta1 CRDs). Prefer `make setup`.

# Ideal, one-shot bootstrap that avoids pre-installing legacy v1beta1 CRDs
# Flow:
#  1) kind cluster (fresh)
#  2) cert-manager
#  3) clusterctl init (CAPI core + kubeadm)
#  4) apply CAPT control-plane/infrastructure CRDs
#  5) install Crossplane and provider-terraform
#  6) deploy CAPT (manager + webhooks + certs)
#  7) wait for CA injection (fallback to manual inject if needed)
.PHONY: setup
KIND_NAME ?= capt
setup: ## One-shot bootstrap (v1beta2-compatible): kind, cert-manager, clusterctl, Crossplane, CAPT, CA
	$(MAKE) kind-recreate
	$(MAKE) setup-cert-manager
	$(MAKE) setup-capi
	$(MAKE) setup-crossplane
	$(MAKE) setup-provider-terraform
	# Build local manager image and load into kind
	$(MAKE) docker-build IMG=$(SETUP_IMG)
	- kind load docker-image $(SETUP_IMG) --name $(KIND_NAME) || { \
		$(CONTAINER_TOOL) save $(SETUP_IMG) | kind load image-archive /dev/stdin --name $(KIND_NAME); \
	}
	# Deploy CAPT using the locally built image
	$(MAKE) deploy IMG=$(SETUP_IMG)
	# Wait for CA injection in webhooks & CRDs
	$(MAKE) wait-ca

.PHONY: setup-ideal
setup-ideal: setup ## Alias of `setup`

.PHONY: kind-recreate
kind-recreate: ## Recreate kind cluster named $(KIND_NAME)
	- kind delete cluster --name $(KIND_NAME)
	kind create cluster --name $(KIND_NAME)

.PHONY: wait-ca
wait-ca: ## Wait for cainjector to inject caBundle (verification only; no manual patch)
	@echo "Waiting for webhook TLS secret..."
	@i=0; until kubectl -n capt-system get secret webhook-server-cert >/dev/null 2>&1; do i=$$((i+1)); if [ $$i -gt 120 ]; then echo "timeout waiting for webhook-server-cert"; exit 1; fi; sleep 2; done
	@echo "Waiting for cainjector to inject caBundle into admission webhooks..."
	@i=0; until [ $$(kubectl get mutatingwebhookconfiguration capt-mutating-webhook-configuration -o jsonpath='{.webhooks[*].clientConfig.caBundle}' | wc -c) -gt 1 ]; do i=$$((i+1)); if [ $$i -gt 60 ]; then echo "error: cainjector did not inject caBundle in time"; exit 1; fi; sleep 2; done
	@echo "Waiting for CRD conversion caBundle..."
	@for crd in \
	  captclusters.infrastructure.cluster.x-k8s.io \
	  captclustertemplates.infrastructure.cluster.x-k8s.io \
	  captmachines.infrastructure.cluster.x-k8s.io \
	  captmachinesets.infrastructure.cluster.x-k8s.io \
	  captmachinetemplates.infrastructure.cluster.x-k8s.io \
	  captmachinedeployments.infrastructure.cluster.x-k8s.io \
	  workspacetemplates.infrastructure.cluster.x-k8s.io \
	  workspacetemplateapplies.infrastructure.cluster.x-k8s.io \
	  captcontrolplanes.controlplane.cluster.x-k8s.io \
	  captcontrolplanetemplates.controlplane.cluster.x-k8s.io; do \
	  i=0; \
	  until [ $$(kubectl get crd $$crd -o jsonpath='{.spec.conversion.webhook.clientConfig.caBundle}' | wc -c) -gt 1 ]; do \
	    i=$$((i+1)); \
	    if [ $$i -gt 60 ]; then echo "error: cainjector did not inject caBundle to CRD $$crd in time"; exit 1; fi; \
	    sleep 2; \
	  done; \
	done
	@echo "CA injection ensured."

.PHONY: all-in-one-noop
all-in-one-noop: setup run-noop-e2e ## Bootstrap and run no-op ClusterClass e2e

.PHONY: run-noop-e2e
run-noop-e2e: ## Run the no-op ClusterClass e2e script
	./test/e2e/scripts/no-op-clusterclass.sh

.PHONY: setup-capi-crds
setup-capi-crds: ## Install only Cluster API Core CRDs (avoid conflicting tf.upbound.io CRDs).
	# Ensure no conflicting tf.upbound.io CRDs from third_party are present
	$(KUBECTL) delete crd workspaces.tf.upbound.io --ignore-not-found
	# Apply only the CAPI Cluster CRD
	$(KUBECTL) apply -f third_party/cluster-api/config/crd/bases/cluster.x-k8s.io_clusters.yaml

.PHONY: setup-capi
CAPT_NAMESPACE ?= capt-system

setup-capi: clusterctl-setup clusterctl ## Install CAPI via clusterctl with kubeadm bootstrap and CAPT providers.
	@echo "Initializing Cluster API with kubeadm bootstrap and CAPT providers..."
	# Cleanup stale clusterctl inventory to avoid decode errors
	-$(KUBECTL) -n $(CAPT_NAMESPACE) delete configmap clusterctl --ignore-not-found
	-$(KUBECTL) -n capi-system delete configmap clusterctl --ignore-not-found
	export CLUSTER_TOPOLOGY=true && \
	$(CLUSTERCTL_BIN) init --core cluster-api --bootstrap kubeadm --target-namespace $(CAPT_NAMESPACE) --config capi-local-config.yaml
	@echo "✓ clusterctl init completed"
	# Apply infrastructure & control-plane components (installed outside clusterctl to avoid provider inventory issues)
	$(KUBECTL) apply -f capt/infrastructure-capt/v0.0.0/infrastructure-components.yaml
	$(KUBECTL) apply -f capt/control-plane-capt/v0.0.0/control-plane-components.yaml
	# Wait CRDs for ClusterClass topology & kubeadm bootstrap
	$(KUBECTL) wait --for=condition=Established crd/clusterclasses.cluster.x-k8s.io --timeout=120s || true
	$(KUBECTL) wait --for=condition=Established crd/kubeadmconfigtemplates.bootstrap.cluster.x-k8s.io --timeout=120s || true

.PHONY: setup-cert-manager
setup-cert-manager: ## Install cert-manager (required by Crossplane webhooks) and wait for CRDs.
	$(KUBECTL) apply -f https://github.com/cert-manager/cert-manager/releases/download/$(CERT_MANAGER_VERSION)/cert-manager.yaml
	$(KUBECTL) wait --for=condition=Established crd/certificates.cert-manager.io --timeout=300s || true

.PHONY: setup-crossplane
setup-crossplane: helm3 ## Install Crossplane via Helm and wait for CRDs (auto-installs helm locally if missing).
	@HELM_CMD=helm; \
	if ! command -v helm >/dev/null 2>&1; then HELM_CMD=$(HELM_BIN); fi; \
	$$HELM_CMD repo add crossplane-stable https://charts.crossplane.io/stable >/dev/null 2>&1 || true; \
	$$HELM_CMD repo update >/dev/null; \
	$$HELM_CMD upgrade --install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace --version $(CROSSPLANE_VERSION); \
	echo "Waiting for Crossplane CRDs to be created..."; \
	i=0; until $(KUBECTL) get crd/providers.pkg.crossplane.io >/dev/null 2>&1; do i=$$((i+1)); if [ $$i -gt 120 ]; then echo "timeout waiting for providers.pkg.crossplane.io"; exit 1; fi; sleep 2; done; \
	$(KUBECTL) wait --for=condition=Established crd/providers.pkg.crossplane.io --timeout=300s; \
	i=0; until $(KUBECTL) get crd/deploymentruntimeconfigs.pkg.crossplane.io >/dev/null 2>&1; do i=$$((i+1)); if [ $$i -gt 120 ]; then echo "warn: deploymentruntimeconfigs.pkg.crossplane.io not found (older versions)"; break; fi; sleep 2; done; \
	if $(KUBECTL) get crd/deploymentruntimeconfigs.pkg.crossplane.io >/dev/null 2>&1; then $(KUBECTL) wait --for=condition=Established crd/deploymentruntimeconfigs.pkg.crossplane.io --timeout=300s; fi; \
	$(KUBECTL) -n crossplane-system rollout status deploy/crossplane --timeout=300s || true; \
	$(KUBECTL) -n crossplane-system rollout status deploy/crossplane-rbac-manager --timeout=300s || true

.PHONY: setup-provider-terraform
setup-provider-terraform: ## Install Upbound provider-terraform and wait for tf.upbound.io CRDs (via kustomize in config/crossplane/terraform).
	$(KUSTOMIZE_BUILD) config/crossplane/terraform | $(KUBECTL) apply -f -
	# Poll until ProviderConfig CRD appears, then wait for Established
	echo "Waiting for tf.upbound.io ProviderConfig CRD..."; \
	i=0; until $(KUBECTL) get crd/providerconfigs.tf.upbound.io >/dev/null 2>&1; do i=$$((i+1)); if [ $$i -gt 180 ]; then echo "timeout waiting for providerconfigs.tf.upbound.io"; exit 1; fi; sleep 2; done; \
	$(KUBECTL) wait --for=condition=Established crd/providerconfigs.tf.upbound.io --timeout=600s || true
	# Ensure Workspaces CRD is Established (may already exist)
	if $(KUBECTL) get crd/workspaces.tf.upbound.io >/dev/null 2>&1; then $(KUBECTL) wait --for=condition=Established crd/workspaces.tf.upbound.io --timeout=600s || true; fi

.PHONY: setup-webhook-certs
setup-webhook-certs: ## Generate self-signed webhook certs for local 'make run'.
	mkdir -p /tmp/k8s-webhook-server/serving-certs
	if [ ! -f /tmp/k8s-webhook-server/serving-certs/tls.crt ] || [ ! -f /tmp/k8s-webhook-server/serving-certs/tls.key ]; then \
		if ! command -v openssl >/dev/null 2>&1; then \
			echo "openssl is required to generate self-signed certs"; exit 1; \
		fi; \
		openssl req -x509 -nodes -newkey rsa:2048 \
			-keyout /tmp/k8s-webhook-server/serving-certs/tls.key \
			-out /tmp/k8s-webhook-server/serving-certs/tls.crt \
			-subj "/CN=localhost" -days 365; \
	else \
		echo "Webhook certs already present in /tmp/k8s-webhook-server/serving-certs"; \
	fi

##@ Kind
.PHONY: kind-capt
kind-capt: clusterapi-manifests clusterctl-setup docker-build clusterctl ## Setup complete kind environment with CAPI and CAPT
	@echo "Setting up kind cluster with CAPI and CAPT..."
	kind create cluster --name capt
	@echo "✓ Kind cluster created"
	@echo "Deploying CAPT..."
	$(CONTAINER_TOOL) save ${IMG} | kind load image-archive /dev/stdin --name capt
	$(MAKE) setup-capi
	@echo "✓ CAPT deployed"
	@echo "Setup complete!"

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
HELMIFY ?= $(LOCALBIN)/helmify
HELM_BIN ?= $(LOCALBIN)/helm

## Tool Versions
KUSTOMIZE_VERSION ?= v5.4.3
CONTROLLER_TOOLS_VERSION ?= v0.16.1
ENVTEST_VERSION ?= release-0.19
GOLANGCI_LINT_VERSION ?= v1.63.4
CERT_MANAGER_VERSION ?= v1.16.1
CROSSPLANE_VERSION ?= v1.16.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

# Optional: use kubectl kustomize instead of local binary for 'build' only.
# Usage: make USE_KUBECTL_KUSTOMIZE=1 install
USE_KUBECTL_KUSTOMIZE ?= 0
ifeq ($(USE_KUBECTL_KUSTOMIZE),1)
KUSTOMIZE_BUILD := kubectl kustomize
KUSTOMIZE_PREREQ :=
else
KUSTOMIZE_WORKS := $(shell $(KUSTOMIZE) version >/dev/null 2>&1 && echo yes || echo no)
ifeq ($(KUSTOMIZE_WORKS),yes)
KUSTOMIZE_BUILD := $(KUSTOMIZE) build
KUSTOMIZE_PREREQ := kustomize
else
KUSTOMIZE_BUILD := kubectl kustomize
KUSTOMIZE_PREREQ :=
endif
endif

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: helmify
helmify: $(HELMIFY) ## Download helmify locally if necessary. Used by 'helm' target to generate charts/capt.
$(HELMIFY): $(LOCALBIN)
	test -s $(LOCALBIN)/helmify || GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@latest

.PHONY: helm3
helm3: $(HELM_BIN) ## Download helm locally if necessary.
$(HELM_BIN): $(LOCALBIN)
	@test -s $(HELM_BIN) || GOBIN=$(LOCALBIN) go install helm.sh/helm/v3/cmd/helm@v3.15.2

.PHONY: helm
helm: clusterapi-manifests $(KUSTOMIZE_PREREQ) helmify ## Generate helm charts
	@echo "Generating Helm chart in charts/capt"
	@mkdir -p charts/capt
	$(KUSTOMIZE_BUILD) config/default | $(HELMIFY) charts/capt

.PHONY: clean
clean: ## Clean up generated files
	rm -f capt-image-bundle.tar
	rm -rf capt/
	rm -f capi-local-config.yaml

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: clusterctl
clusterctl: $(LOCALBIN) ## Download clusterctl locally (prebuilt release with GitVersion)
	@set -e; \
	url=https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CLUSTERCTL_VERSION)/clusterctl-linux-$(CLUSTERCTL_ARCH); \
	echo "Downloading $$url"; \
	curl -fsSL $$url -o $(LOCALBIN)/clusterctl.tmp; \
	chmod +x $(LOCALBIN)/clusterctl.tmp; \
	mv $(LOCALBIN)/clusterctl.tmp $(LOCALBIN)/clusterctl; \
	$(LOCALBIN)/clusterctl version || true

CLUSTERCTL_BIN ?= $(LOCALBIN)/clusterctl

# clusterctl prebuilt binary (ensures GitVersion is embedded)
CLUSTERCTL_VERSION ?= v1.11.2
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
  CLUSTERCTL_ARCH := amd64
else ifeq ($(UNAME_M),aarch64)
  CLUSTERCTL_ARCH := arm64
else
  CLUSTERCTL_ARCH := amd64
endif
