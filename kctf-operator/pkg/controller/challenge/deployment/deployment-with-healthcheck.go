package deployment

import (
	kctfv1 "github.com/google/kctf/pkg/apis/kctf/v1"
	utils "github.com/google/kctf/pkg/controller/challenge/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment with Healthcheck
func withHealthcheck(challenge *kctfv1.Challenge) *appsv1.Deployment {
	dep := deployment(challenge)

	idx_challenge := utils.IndexOfContainer("challenge", dep.Spec.Template.Spec.Containers)
	idx_healthcheck := utils.IndexOfContainer("healthcheck", dep.Spec.Template.Spec.Containers)

	// Get the container with the challenge and add healthcheck configurations
	dep.Spec.Template.Spec.Containers[idx_challenge].LivenessProbe = &corev1.Probe{
		FailureThreshold: 2,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(45281),
			},
		},
		InitialDelaySeconds: 45,
		TimeoutSeconds:      3,
		PeriodSeconds:       30,
	}

	dep.Spec.Template.Spec.Containers[idx_challenge].ReadinessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(45281),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      3,
		PeriodSeconds:       5,
	}

	if idx_healthcheck == -1 {
		healthcheckContainer := corev1.Container{
			Name: "healthcheck",
		}
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, healthcheckContainer)
		idx_healthcheck = len(dep.Spec.Template.Spec.Containers) - 1
	}

	dep.Spec.Template.Spec.Containers[idx_healthcheck].Image = challenge.Spec.Healthcheck.Image
	dep.Spec.Template.Spec.Containers[idx_healthcheck].Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(1000, resource.DecimalSI),
		},
		Requests: corev1.ResourceList{
			"cpu": *resource.NewMilliQuantity(50, resource.DecimalSI),
		},
	}

	dep.Spec.Template.Spec.Containers[idx_healthcheck].VolumeMounts = []corev1.VolumeMount{{
		Name:      "pow-bypass",
		ReadOnly:  true,
		MountPath: "/pow-bypass",
	}}

	healthcheckVolume := corev1.Volume{
		Name: "pow-bypass",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "pow-bypass",
			},
		},
	}

	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, healthcheckVolume)

	return dep
}
