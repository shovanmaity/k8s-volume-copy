PV Populator is a `Volume Populator` that helps to rename a PVC. It does a patch on `PersistentVolume` of older `Claim` with new `Claim`.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Default storageclass should be configured for this demo.
3. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
4. Before rename volume should not be used by any application.

Here are the steps to rename a pvc using PV Populator.
1. Install volume populator.
   ```bash
   kubectl apply -f config/crd/demo.io_persistentvolumepopulators.yaml
   kubectl apply -f yaml/populator/pv/deploy.yaml
   ```
2. Create a volume and install a demo app to write some data on that volume.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/populator/pv/app/pvc.yaml
   kubectl apply -f yaml/populator/pv/app/pod.yaml
   ```
   ```bash
   shovan@probot:~$ kubectl exec -it demo sh
   / #
   / # cd /data/
   /data # ls -l
   total 16
   drwx------    2 root     root         16384 Jun 14 16:17 lost+found
   /data # echo "hello!" > file
   /data # cat file
   hello!
   ```
3. After writing some data delete the pod.
   ```bash
   kubectl delete -f yaml/populator/pv/app/pod.yaml
   ```
4. Create a pvpopulator cr. It has old pvc name in the spec.
   ```bash
   kubectl apply -f yaml/populator/pv/cr.yaml
   ```
   ```yaml
   apiVersion: demo.io/v1
   kind: PersistentVolumePopulator
   metadata:
     name: pv-populator
   spec:
     pvcName: my-pvc
   ```
5. Create a new pvc pointing to the claim-cpopulator.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/populator/pv/app/pvc-d.yaml
   ```
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: my-pvc-dash
   spec:
    #storageClassName: my-sc
     dataSource:
       apiGroup: demo.io
       kind: PersistentVolumePopulator
       name: pv-populator
     accessModes:
     - ReadWriteOnce
     volumeMode: Filesystem
     resources:
       requests:
         storage: 2Gi
   ```
6. Create a new pod and check the older data is present or not.
   ```bash
   kubectl apply -f yaml/populator/pv/app/pod-d.yaml
   ```
   ```bash
   shovan@probot:~$ kubectl exec -it demo sh
   / #
   / # cd /data/
   /data # ls -lrth
   total 20K
   drwx------    2 root     root       16.0K Jun 14 16:30 lost+found
   -rw-r--r--    1 root     root           7 Jun 14 16:31 file
   /data # cat file
   hello!
   ```
7. Cleanup the resources.
   ```bash
   kubectl delete -f yaml/populator/pv/app/pod-d.yaml
   kubectl delete -f yaml/populator/pv/app/pvc-d.yaml
   ```
   ```bash
   kubectl delete -f yaml/populator/pv/cr.yaml
   ```
   ```bash
   kubectl delete -f config/crd/demo.io_persistentvolumepopulators.yaml
   kubectl delete -f yaml/populator/pv/deploy.yaml
   ```
