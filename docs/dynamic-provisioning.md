## 使用步骤

1. 创建一个 StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: luscsi.luskit.io
provisioner: luscsi.luskit.io
volumeBindingMode: Immediate
parameters:
  mgsAddress: 172.30.1.11@o2ib
  fsName: lstore
  subdir: /test1
```

2. 创建一个 PersistentVolumeClaim

```yaml
```
