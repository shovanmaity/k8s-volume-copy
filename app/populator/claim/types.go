package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeClaimPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeClaimPopulatorSpec   `json:"spec"`
	Status VolumeClaimPopulatorStatus `json:"status"`
}

type VolumeClaimPopulatorSpec struct {
	Name string `json:"name"`
}

type VolumeClaimPopulatorStatus struct {
}
