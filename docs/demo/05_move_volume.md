In this demo we will move PVC from one storageclass to another storageclass. In these yaml storage class `hp-node-1` and `hp-node-2` are hard coded, if you want you can change it. In this demo we will use `VolumeRename` and `VolumeCopy`. Flow is `scale down` -> `rename volume(s)` -> `copy volume(s)` -> `scale up`.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Storageclass should be configured for this demo.
3. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
4. Before moving volume should not be used by any application.

Here are steps -
1. Install PV and Rsync volume populator and volume rename and copy controller.
   ```bash
   kubectl create ns volume-copy
   ```
   ```bash
   kubectl apply -f config/crd/demo.io_persistentvolumepopulators.yaml
   kubectl apply -f config/crd/demo.io_rsyncpopulators.yaml
   kubectl apply -f config/crd/demo.io_volumerenames.yaml
   kubectl apply -f config/crd/demo.io_volumecopies.yaml
   kubectl apply -f yaml/populator/pv/deploy.yaml
   kubectl apply -f yaml/populator/rsync/deploy.yaml
   kubectl apply -f yaml/volume-rename/deploy.yaml
   kubectl apply -f yaml/volume-copy/deploy.yaml
   ```
2. Deploy minio app with `hp-node-1` storageclass.
   ```bash
   kubectl apply -f yaml/volume-move/app.yaml
   ```
3. After adding some file in minio scale it down.
   ```bash
   kubectl scale sts minio --replicas=0
   ```
4. Apply all volume rename CR and wait for it to come to `Completed` state.
   ```bash
   kubectl apply -f yaml/volume-move/rename/1.yaml
   kubectl apply -f yaml/volume-move/rename/2.yaml
   kubectl apply -f yaml/volume-move/rename/3.yaml
   kubectl apply -f yaml/volume-move/rename/4.yaml
   ```
   ```bash
   kubectl get volumerename.demo.io -o=jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.namespace}{"\t"}{.status.state}{"\n"}{end}'
   ```
5. Apply all volume copy CR and wait for it to come to `Completed` state. Please change the storageclass if you want. Wait for volume copy CR to come to  `WaitingForConsumer` or `Completed` state.
   ```bash
   kubectl apply -f yaml/volume-move/copy/1.yaml
   kubectl apply -f yaml/volume-move/copy/2.yaml
   kubectl apply -f yaml/volume-move/copy/3.yaml
   kubectl apply -f yaml/volume-move/copy/4.yaml
   ```
   ```bash
   kubectl get volumecopy.demo.io -o=jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.namespace}{"\t"}{.status.state}{"\n"}{end}'
   ```
6. Scale up the minio app and check the data.
   ```bash
   kubectl scale sts minio --replicas=4
   ```
7. Cleanup the resources.
   ```bash
      kubectl delete -f config/crd/demo.io_persistentvolumepopulators.yaml
      kubectl delete -f config/crd/demo.io_rsyncpopulators.yaml
      kubectl delete -f config/crd/demo.io_volumerenames.yaml
      kubectl delete -f config/crd/demo.io_volumecopies.yaml
      kubectl delete -f yaml/populator/pv/deploy.yaml
      kubectl delete -f yaml/populator/rsync/deploy.yaml
      kubectl delete -f yaml/volume-rename/deploy.yaml
      kubectl delete -f yaml/volume-copy/deploy.yaml
   ```
   ```bash
   kubectl delete ns volume-copy
   ```