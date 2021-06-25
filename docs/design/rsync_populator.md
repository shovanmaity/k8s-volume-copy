Rsync Populator is a `Volume Populator` that helps to create volume using rsync source. In this project we are creating rsync daemon on source volume and then populator will use that source to create a volume.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Namespace `kvm-a1b2c3d4e5` is reserved for volume populator. Don't create any application or pvc in that namespace.

Here are the steps to create a volume using rsync source.
1. Install volume populator
   ```bash
   kubectl apply -f yaml/populator/rsync/crd.yaml
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
4. Install rsync daemon as a source, it will mount source volume.
   ```bash
   kubectl apply -f yaml/server/rsync/deploy.yaml
   ```
5. Create a rsyncpopulator cr. It has all the details of source rsyncd.
   ```bash
   kubectl apply -f yaml/populator/rsync/cr.yaml
   ```
   ```yaml
   apiVersion: kvm.io/v1
   kind: RsyncPopulator
   metadata:
     name: rsync-populator
   spec:
     username: user
     password: pass
     service: rsyncd.default
     path: /data
   ```
6. Create a new pvc pointing to the rsync-populator.
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
       apiGroup: kvm.io
       kind: RsyncPopulator
       name: rsync-populator
     accessModes:
     - ReadWriteOnce
     volumeMode: Filesystem
     resources:
       requests:
         storage: 2Gi
   ```
7. Create a new pod and check the older data is present or not.
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
   kubectl delete -f yaml/populator/rsync/crd.yaml
   kubectl delete -f yaml/populator/rsync/deploy.yaml
   ```
