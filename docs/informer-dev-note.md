# Dev note for namespace watchers controller module

## How to build Docker image
```bash
$ IMG=docker.io/woojoong/namespace-watcher-informer:0.0.1 make docker-build
$ docker push docker.io/woojoong/namespace-watcher-informer:0.0.1
```

## How to deploy to Kubernetes
```bash
$ helm install namespace-watcher-informer . -n namespace-watcher-informer --create-namespace
```

## How to delete from Kubernetes
```bash
$ helm uninstall -n namespace-watcher-informer namespace-watcher-informer
```