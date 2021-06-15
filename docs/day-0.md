Here are the steps to move a pvc from one storageclass to another storageclass. As only core components are used this needs some manual steps.

NOTE -
1. `AnyVolumeDataSource` feature gate should be enabled in the kubernetes cluster.
2. Destination storageclass should pointing to a CSI driver.
3. Both the storageclass should be configured.
4. Namespace `volume-migration` is reserved for volume populator. Don't create any application or pvc in that namespace.

Here are the steps to migrate a pvc from one storageclass to another storageclass.
1. Install volume populator
   ```bash
   kubectl apply -f package/populator/kubernetes/crd.yaml
   kubectl apply -f package/populator/kubernetes/deploy.yaml
   ```
2. Create a volume and install a demo app to write some data on that volume.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f package/populator/kubernetes/app/pvc.yaml
   kubectl apply -f package/populator/kubernetes/app/pod.yaml
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
3. After writing some data delete the pod
   ```bash
   kubectl delete -f package/populator/kubernetes/app/pod.yaml
   ```
4. Install rsyncd, it will mount source volume.
   ```bash
   kubectl apply -f package/server/kubernetes/deploy.yaml
   ```
5. Create a rsyncpopulator cr. It has all the details of source rsyncd.
   ```bash
   kubectl apply -f package/populator/kubernetes/cr.yaml
   ```
   ```yaml
   apiVersion: example.io/v1
   kind: RsyncPopulator
   metadata:
     name: rsync-populator-1
   spec:
     username: user
     password: pass
     service: rsyncd.default
     path: /data
   ```
6. Create a new pvc pointing to the rsyncpopulator.
   ```bash
   # Please edit the storageclass accordingly
   kubectl apply -f package/populator/kubernetes/app/pvcd.yaml
   ```
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: my-pvc-dash
   spec:
    #storageClassName: my-sc
     dataSource:
       apiGroup: example.io
       kind: RsyncPopulator
       name: rsync-populator-1
     accessModes:
     - ReadWriteOnce
     volumeMode: Filesystem
     resources:
       requests:
         storage: 2Gi
   ```
7. Create a new pod to check the older data is present or not.
   ```bash
   kubectl apply -f package/populator/kubernetes/app/podd.yaml
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
