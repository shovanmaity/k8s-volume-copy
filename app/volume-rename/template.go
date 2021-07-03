package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	internalv1 "github.com/shovanmaity/k8s-volume-copy/client/apis/demo.io/v1"
)

type templateConfig struct {
	pvcName string
	newName string
}

// templateFromVolumeRename returns templatConfig object created from VolumeRename CR.
// templateConfig used to get populator and pvc object
func templateFromVolumeRename(cr internalv1.VolumeRename) (*templateConfig, error) {
	tc := &templateConfig{
		pvcName: cr.Spec.OldPVC,
		newName: cr.Spec.NewPVC,
	}
	return tc, nil
}

// getPersistentVolumePopulatorTemplate returns PersistentVolumePopulator object. It
// is a volume populator and it helps to create a new volume using this as a source.
func (tc *templateConfig) getPersistentVolumePopulatorTemplate() internalv1.PersistentVolumePopulator {
	populator := internalv1.PersistentVolumePopulator{
		TypeMeta: metav1.TypeMeta{
			Kind:       pvpKind,
			APIVersion: groupDemoIO + "/" + versionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: populatorName + "-" + tc.pvcName,
			Labels: map[string]string{
				nameLabel:      populatorName + "-" + tc.pvcName,
				roleLabel:      componentName,
				createdByLabel: componentName,
				managedByLabel: componentName,
			},
		},
		Spec: internalv1.PersistentVolumePopulatorSpec{
			PVCName: tc.pvcName,
		},
	}
	return populator
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
			Name:        tc.newName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: pvc.Spec,
	}

	// Set Populator details
	pvcDash.Spec.DataSource = &corev1.TypedLocalObjectReference{
		Kind: pvpKind,
		APIGroup: func() *string {
			name := groupDemoIO
			return &name
		}(),
		Name: populatorName + "-" + tc.pvcName,
	}

	// remove volume name
	pvcDash.Spec.VolumeName = ""
	return pvcDash
}
