# Create and delete namespaces

## create and then delete 100 namespaces
```bash
$ for i in {1..100}; do kubectl create ns test-$i & ; done
$ for i in {1..100}; do kubectl delete ns test-$i & ; done
```

## What if we want to remove finalizers?
```bash
$ for i in {1..100}; do kubectl patch ns test-$i -p '{"metadata": {"finalizers": null}}' & ; done
```