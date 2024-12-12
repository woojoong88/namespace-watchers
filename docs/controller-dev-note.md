# Dev note for namespace watchers controller module

## How to scaffold with operator-sdk

```bash
$ operator-sdk init --repo github.com/woojoong88/namespace-watchers/modules/controller
$ operator-sdk create api --group core --version v1 --kind Pod 
```

## How to build Docker image
```bash
$ IMG=docker.io/woojoong/namespace-watcher-controller:0.0.1 make docker-build
$ docker push docker.io/woojoong/namespace-watcher-controller:0.0.1
```

## How to deploy to Kubernetes
```bash
$ helm install namespace-watcher-controller . -n namespace-watcher-controller --create-namespace
```