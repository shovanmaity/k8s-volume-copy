Rsync Populator is a volume populator that helps to create a volume from any rsync source. Volume Copy control loop, it makes use of Rsync Populator to rename a volume. When a `VolumeCopy` CR is created it sets up rsync source on the source PVC and creates a RsyncPopulator and a new PVC pointing to that volume populator as a data source. 

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
3. Before copying volume should not be used by any application.

Here are the steps to copy a pvc -
1. Install volume populator and volume copy controller.
   ```bash
   kubectl create ns volume-copy
   ```
   ```bash
   kubectl apply -f config/crd/demo.io_rsyncpopulators.yaml
   kubectl apply -f config/crd/demo.io_volumecopies.yaml
   kubectl apply -f yaml/populator/rsync/deploy.yaml
   kubectl apply -f yaml/volume-copy/deploy.yaml
   ```
2. Create a volume and install a demo app to write some data on that volume.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/volume-copy/app/pvc.yaml
   kubectl apply -f yaml/volume-copy/app/pod.yaml
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
   kubectl delete -f yaml/volume-copy/app/pod.yaml
   ```
4. Create a  `VolumeCopy` cr. It has old and new pvc details in the spec.
   ```bash
   kubectl apply -f yaml/volume-copy/cr.yaml
   ```
   ```yaml
   apiVersion: demo.io/v1
   kind: VolumeCopy
   metadata:
     name: volume-copy
   spec:
     sourceNamespace: default
     sourcePVC: my-pvc
     destinationPVC: my-pvc-dash
     destinationSC: my-sc
   ```
5. Wait for volume copy comes to `WaitingForConsumer` or `Completed` state.
   ```bash
   kubectl get volumecopy.demo.io/volume-copy -o=jsonpath="{.status.state}{'\n'}"
   ```
6. Create a new pod and check the older data is present or not in the new PVC.
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
   kubectl delete -f yaml/volume-copy/cr.yaml
   ```
   ```bash
   kubectl delete -f yaml/volume-copy/app/pod-d.yaml
   ```
   ```bash
   kubectl delete -f yaml/populator/rsync/deploy.yaml
   kubectl delete -f yaml/volume-copy/deploy.yaml
   kubectl delete -f config/crd/demo.io_rsyncpopulators.yaml
   kubectl delete -f config/crd/demo.io_volumecopies.yaml
   ```
   ```bash
   kubectl delete pvc my-pvc-dash
   kubectl delete ns volume-copy
   ```
