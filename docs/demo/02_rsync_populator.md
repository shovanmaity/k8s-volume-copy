Rsync Populator is a volume populator that helps to create volume from any rsync source. Rsync client is used as volume populator plugin. `RsyncPopulator` CR contains the information of source location and how to access the source.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Namespace `volume-copy` is reserved for volume populator. Don't create any application or pvc in that namespace.
3. Before rename volume should not be used by any application.

Here are the steps of how to use rsync populator -
1. Install volume populator
   ```bash
   kubectl create ns volume-copy
   ```
   ```bash
   kubectl apply -f config/crd/demo.io_rsyncpopulators.yaml
   kubectl apply -f yaml/populator/rsync/deploy.yaml
   ```
2. Create a volume and install a demo app to write some data on that volume.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/populator/rsync/app/pvc.yaml
   kubectl apply -f yaml/populator/rsync/app/pod.yaml
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
   kubectl delete -f yaml/populator/rsync/app/pod.yaml
   ```
4. Setup rsync source, install rsync daemon pod, configmap and service it will mount source volume.
   ```bash
   kubectl apply -f yaml/server/rsync/deploy.yaml
   ```
5. Create a rsyncpopulator cr. It has all the details of rsync source.
   ```bash
   kubectl apply -f yaml/populator/rsync/cr.yaml
   ```
   ```yaml
   apiVersion: demo.io/v1
   kind: RsyncPopulator
   metadata:
     name: rsync-populator
   spec:
     username: user
     password: pass
     url: rsync-daemon.default:873
     path: /data
   ```
6. Create a new pvc pointing to rsync-populator.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f yaml/populator/rsync/app/pvc-d.yaml
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
       kind: RsyncPopulator
       name: rsync-populator
     accessModes:
     - ReadWriteOnce
     volumeMode: Filesystem
     resources:
       requests:
         storage: 2Gi
   ```
7. Create a new pod and check the older data is present or not in the new PVC.
   ```bash
   kubectl apply -f yaml/populator/rsync/app/pod-d.yaml
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
8. Cleanup the resources.
   ```bash
   kubectl delete -f yaml/populator/rsync/app/pod-d.yaml
   kubectl delete -f yaml/populator/rsync/app/pvc-d.yaml
   ```
   ```bash
   kubectl delete -f yaml/populator/rsync/cr.yaml
   ```
   ```bash
   kubectl delete -f yaml/server/rsync/deploy.yaml
   kubectl delete -f yaml/populator/rsync/app/pvc.yaml

   ```
   ```bash
   kubectl delete -f config/crd/demo.io_rsyncpopulators.yaml
   kubectl delete -f yaml/populator/rsync/deploy.yaml
   ```
   ```bash
   kubectl delete ns volume-copy
   ```
