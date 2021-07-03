[PV Populator](https://github.com/shovanmaity/k8s-volume-copy/blob/main/docs/demo/01_pv_populator.md) is a `Volume Populator` that helps to rename a PVC. It does a patch on `PersistentVolume` of older `Claim` with new `Claim`. Volume Rename control loop use this to rename a volume. We can rename a pvc using PV Populator there we need to create populator and pvc. Volume Rename takes care of these 2 steps. We just need to create `VolumeRename` CR.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Default storageclass should be configured for this demo.
3. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
4. Before rename volume should not be used by any application.

Here are the steps to rename a pvc.
1. Install volume populator and volume rename controller.
   ```bash
   kubectl apply -f config/crd/demo.io_persistentvolumepopulators.yaml
   kubectl apply -f config/crd/demo.io_volumerenames.yaml
   kubectl apply -f yaml/populator/pv/deploy.yaml
   kubectl apply -f yaml/volume-rename/deploy.yaml
   ```
2. Create a volume and install a demo app to write some data on that volume.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/volume-rename/app/pvc.yaml
   kubectl apply -f yaml/volume-rename/app/pod.yaml
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
   kubectl delete -f yaml/volume-rename/app/pod.yaml
   ```
4. Create a pvpopulator cr. It has old pvc name in the spec.
   ```bash
   kubectl apply -f yaml/volume-rename/cr.yaml
   ```
   ```yaml
   apiVersion: demo.io/v1
   kind: VolumeRename
   metadata:
     name: volume-rename
   spec:
     oldPVC: my-pvc
     newPVC: my-pvc-dash
   ```
5. Wait for volume rename comes to `Completed` state.
   ```bash
   kubectl get volumerename.demo.io/volume-rename -o=jsonpath="{.status.state}{'\n'}"
   ```
6. Create a new pod and check the older data is present or not.
   ```bash
   kubectl apply -f yaml/volume-rename/app/pod-d.yaml
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
   kubectl delete -f yaml/volume-rename/cr.yaml
   ```
   ```bash
   kubectl delete -f yaml/volume-rename/app/pod-d.yaml
   ```
   ```bash
   kubectl delete -f yaml/populator/pv/deploy.yaml
   kubectl delete -f yaml/volume-rename/deploy.yaml
   kubectl delete -f config/crd/demo.io_persistentvolumepopulators.yaml
   kubectl delete -f config/crd/demo.io_volumerenames.yaml
   ```
