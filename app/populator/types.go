package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RsyncPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RsyncPopulatorSpec   `json:"spec"`
	Status RsyncPopulatorStatus `json:"status"`
}

type RsyncPopulatorSpec struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
	Service  string `json:"service"`
}

type RsyncPopulatorStatus struct {
}
