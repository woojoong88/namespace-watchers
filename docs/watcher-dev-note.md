# Dev note for namespace watchers controller module

## How to build Docker image
```bash
$ IMG=docker.io/woojoong/namespace-watcher-watcher:0.0.1 make docker-build
$ docker push docker.io/woojoong/namespace-watcher-watcher:0.0.1
```

## How to deploy to Kubernetes
```bash
$ helm install namespace-watcher-watcher . -n namespace-watcher-watcher --create-namespace
```

## How to delete from Kubernetes
```bash
$ helm uninstall -n namespace-watcher-watcher namespace-watcher-watcher
```