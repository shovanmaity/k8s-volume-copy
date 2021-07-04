PV Populator is a volume populator that does a patch on PV of older PVC with new PVC. Here source of the volume is the PV of any older PVC. Both the PVC(new and source) should be from same storageclass. It is like mentioning PVC as a data source for a PVC. There one clone volume will be created and here PV of the mentioned PVC will be patched with new PVC. In this way we will be able to rename a PVC and it will refer to the PV of older PVC.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
3. Before rename volume should not be used by any application.

Here are the steps of how to use persistent volume populator -
1. Install volume populator.
   ```bash
   kubectl create ns volume-copy
   ```
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
3. After writing some data delete the Pod.
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
5. Create a new pvc pointing to pv-cpopulator.
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
6. Create a new Pod and check the older data is present or not in the new PVC.
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
   ```bash
   kubectl delete ns volume-copy
   ```
