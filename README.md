# Gitea API Helpers

A set of helpers to interact and configure [Gitea](https://gitea.io/en-us/) using [Gitea API](https://docs.gitea.io/en-us/api-usage/). The helper also has set of Kubernetes jobs that could be used to configure the Gitea using Kubernetes jobs

> __NOTE__: This is intended only for Demo purpose and currently developed to be used with Drone CI

## Required tools

- [Kustomize](https://kustomize.io/)
- [envsusbst](https://www.man7.org/linux/man-pages/man1/envsubst.1.html)

All linux distributions adds **envsubst** via [gettext](https://www.gnu.org/software/gettext/) package. On macOS it can be installed using [Homebrew](https://brew.sh/) like `brew install gettext`.

## Clone the Sources

```shell
git clone https://github.com/kameshsampath/gitea-api-helper && \
  cd "$(basename "$_" .git)"
export GITEA_HELPER_HOME="${PWD}"
```

## Build and Test locally

The following section details on how to build and test the helper locally.

### Create Kubernetes Cluster

```shell
$GITEA_HELPER_HOME/bin/kind.sh
```

### Deploy Gitea

```shell
helm repo add gitea-charts https://dl.gitea.io/charts/
helm repo update
helm upgrade \
  --install gitea gitea-charts/gitea \
  --values $GITEA_HELPER_HOME/helm_vars/gitea/values.yaml \
  --wait
```

Gitea service can be accessed using the url <http://gitea-127.0.0.1.sslip.io:30950/>

The default credentials is `demo/demo@123`

### Setup Workshop

### Kubernetes Cluster

Assuming you have,

- Kubernetes Cluster with cluster-admin privileges
- Gitea deployed and running

Create a `kustomization` file like

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: drone
resources:
- https://raw.githubusercontent.com/kameshsampath/drone-tutorial-gitea-helper/master/manifests/ha/install.yaml
## add your overrides
```

__TIP__: This can be useful if you want to override the `workshop.yaml` to suit your settings

Then do,

```shell
kubectl apply -k <your kustomize dir>
```

#### Locally

Create workshop config file like,

```yaml
# The Gitea Configuration
giteaAdminUserName: demo
giteaAdminUserPassword: demo@123
giteaURL: http://gitea-127.0.0.1.sslip.io:30950/
users:
  # the lower bound of user e.g. user-01
  from: 1
  # the upper bound of user e.g. user-10
  to: 2
  # create Gitea oAuth app for user
  oAuthAppName: demo-oauth
  # oAuth redirect URL
  oAuthRedirectURI: https://drone-127.0.0.1.sslip.io:30980
  # add oAuth App ClientID and ClientSecret to Kubernetes Secret
  addKubernetesSecret: true
  # The Namespace where to create the secret, the secret will 
  # use the format demo-oauth-<username>-secret
  secretNamespace: default
  repos:
    - https://github.com/kameshsampath/jar-stack
```

Run the command,

```shell
go run cmd/main.go setup-workshop --workshop-file <path to the workshop config> -k <path to kubeconfig>
```

__TODO__: Release of binaries and kubernetes jobs to do this w/o manually running the command

## Clean up

```shell
 kind delete cluster --name=gitea-dev
```
