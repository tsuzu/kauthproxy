# kubectl-oidc-port-forward [![CircleCI](https://circleci.com/gh/int128/kubectl-oidc-port-forward.svg?style=shield)](https://circleci.com/gh/int128/kubectl-oidc-port-forward)

This is a kubectl plugin of the reverse proxy with the `authorization` header.
It provides OpenID Connect authentication for Kubernetes Dashboard.

**Status:** Proof of concept. Not for production.


## Getting Started

```sh
go build -o kubectl-oidc_port_forward
```

You need to setup Kubernetes OpenID Connect authentication and
install [kubelogin](https://github.com/int128/kubelogin).

```sh
export KUBECONFIG=.kubeconfig

# Login and update the kubeconfig
kubectl oidc-login

# Forward local port to the Kubernetes Dashboard service
kubectl oidc-port-forward svc/kubernetes-dashboard 8888:https/443
```

Open http://localhost:8888 and then the Kubernetes Dashboard should appear.


### How it works

oidc-port-forward starts a reverse proxy and executes the `kubectl port-forward` command.

```
+---------------------------+
| Browser                   |
+---------------------------+
  ↓ http/8888
+---------------------------+
| kubectl oidc-port-forward | This adds an authorization header.
+---------------------------+
  ↓ https/*
+---------------------------+
| kubectl port-forward      | This forwards requests to the service or pod.
+---------------------------+
  ↓ https/443
+---------------------------+
| svc/kubernetes-dashboard  |
+---------------------------+
```
