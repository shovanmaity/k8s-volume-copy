package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	internalv1 "github.com/shovanmaity/k8s-volume-copy/client/apis/demo.io/v1"
)

type templateConfig struct {
	sourcePVCName           string
	sourcePVCNamespace      string
	destinationPVCName      string
	destinationPVCNamespace string
	destinationSCName       string
	imageName               string
	rsyncPassword           string
}

func templateFromVolumeCopy(cr internalv1.VolumeCopy) (*templateConfig, error) {
	tc := &templateConfig{
		sourcePVCName:           cr.Spec.SourcePVC,
		sourcePVCNamespace:      cr.Spec.SourceNamespace,
		destinationPVCName:      cr.Spec.DestinationPVC,
		destinationPVCNamespace: cr.Namespace,
		destinationSCName:       cr.Spec.DestinationSC,
	}
	return tc, nil
}

// getPVCDashTemplate returns dash pvc object using old pvc object
// this is how new pvc object is created
// pick provisioner name and node name annotation
// pick all labels
// add created by label
// add datasource so that it can works with volume populator
// drop spec.volumeName
func (tc *templateConfig) getPVCDashTemplate(pvc corev1.PersistentVolumeClaim) corev1.PersistentVolumeClaim {
	// annotaions for pvc dash
	annotations := make(map[string]string)

	// labels for pvc dash
	labels := pvc.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[createdByLabel] = componentName

	pvcDash := corev1.PersistentVolumeClaim{
		TypeMeta: pvc.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        tc.destinationPVCName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: pvc.Spec,
	}

	// Set Populator details
	pvcDash.Spec.DataSource = &corev1.TypedLocalObjectReference{
		Kind: rsyncKind,
		APIGroup: func() *string {
			name := groupDemoIO
			return &name
		}(),
		Name: "rsync-daemon-" + tc.sourcePVCName,
	}

	// remove volume name
	pvcDash.Spec.VolumeName = ""
	pvcDash.Spec.StorageClassName = func() *string {
		name := tc.destinationSCName
		return &name
	}()
	return pvcDash
}

func (tc *templateConfig) getPodTemplate() corev1.Pod {
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rsync-daemon-" + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				nameLabel:      "rsync-daemon-" + tc.sourcePVCName,
				appLabel:       "rsync-daemon-" + tc.sourcePVCName,
				roleLabel:      "rsync-daemon",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "rsync-daemon",
					Image:           tc.imageName,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  "RSYNC_PASSWORD",
							Value: tc.rsyncPassword,
						},
					},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 873,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: "/data",
						},
						{
							Name:      "config",
							MountPath: "/etc/rsyncd.con",
							SubPath:   "rsyncd.con",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: tc.sourcePVCName,
						},
					},
				},
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "rsync-daemon-" + tc.sourcePVCName,
							},
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	return pod
}

func (tc *templateConfig) getCmTemplate() corev1.ConfigMap {
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rsync-daemon-" + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				nameLabel:      "rsync-daemon-" + tc.sourcePVCName,
				roleLabel:      "rsync-daemon",
			},
		},
		Data: map[string]string{
			"rsyncd.conf": rsyncdconfig,
		},
	}
	return cm
}

func (tc *templateConfig) getSvcTemplate() corev1.Service {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rsync-daemon-" + tc.sourcePVCName,
			Labels: map[string]string{
				createdByLabel: componentName,
				managedByLabel: componentName,
				nameLabel:      "rsync-daemon-" + tc.sourcePVCName,
				roleLabel:      "rsync-daemon",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "rsync-daemon",
					Port:     873,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				nameLabel: "rsync-daemon-" + tc.sourcePVCName,
				appLabel:  "rsync-daemon-" + tc.sourcePVCName,
				roleLabel: "rsync-daemon",
			},
		},
	}
	return svc
}

func (tc *templateConfig) getRsyncPopulatorTemplate() internalv1.RsyncPopulator {
	populator := internalv1.RsyncPopulator{
		TypeMeta: metav1.TypeMeta{
			Kind:       rsyncKind,
			APIVersion: groupDemoIO + "/" + versionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rsync-daemon-" + tc.sourcePVCName,
			Labels: map[string]string{
				nameLabel:      "rsync-daemon-" + tc.sourcePVCName,
				roleLabel:      populatorName,
				createdByLabel: componentName,
				managedByLabel: componentName,
			},
		},
		Spec: internalv1.RsyncPopulatorSpec{
			Username: "user",
			Password: tc.rsyncPassword,
			Path:     "/data",
			URL:      "rsync-daemon-" + tc.sourcePVCName + "." + tc.sourcePVCNamespace + ":873",
		},
	}
	return populator
}

var rsyncdconfig = `
# /etc/rsyncd.conf

# Minimal configuration file for rsync daemon
# See rsync(1) and rsyncd.conf(5) man pages for help

# This line is required by the /etc/init.d/rsyncd script
pid file = /var/run/rsyncd.pid

uid = 0
gid = 0
use chroot = yes
reverse lookup = no
[data]
    hosts deny = *
    hosts allow = 0.0.0.0/0
    read only = false
    path = /data
    auth users = , user:rw
    secrets file = /etc/rsyncd.secrets
    timeout = 600
    transfer logging = true
`
