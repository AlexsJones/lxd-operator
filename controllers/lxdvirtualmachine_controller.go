/*
Copyright 2022 AlexsJones.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	componentsv1alpha1 "lxd-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LXDVirtualMachineReconciler reconciles a LXDVirtualMachine object
type LXDVirtualMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=components.machinery.canonical,resources=lxdvirtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=components.machinery.canonical,resources=lxdvirtualmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=components.machinery.canonical,resources=lxdvirtualmachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LXDVirtualMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile

func (r *LXDVirtualMachineReconciler) checkFinalizers(cnf *componentsv1alpha1.LXDVirtualMachine,
	ctx context.Context) (ctrl.Result, error) {
	finalizerName := "machinery.canonical/finalizer"
	// Check to see if the Cluster has a finalizer
	if cnf.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !ctrlutil.ContainsFinalizer(cnf, finalizerName) {
			ctrlutil.AddFinalizer(cnf, finalizerName)
			if err := r.Update(ctx, cnf); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Check to see if the Cluster is under deletion
		if ctrlutil.ContainsFinalizer(cnf, finalizerName) {

			// TODO: Delete and other bits
			// remove our finalizer from the Cluster and update it.
			ctrlutil.RemoveFinalizer(cnf, finalizerName)
			if err := r.Update(ctx, cnf); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *LXDVirtualMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cnf := &componentsv1alpha1.LXDVirtualMachine{}
	err := r.Get(ctx, req.NamespacedName, cnf)
	if err != nil {
		if errors.IsNotFound(err) {

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err

	}
	// Finalizers ----------------------------------------------------------------------
	if res, err := r.checkFinalizers(cnf, ctx); err != nil {
		return res, err
	}
	hostPath := "Directory"
	// Create pod ----------------------------------------------------------------------
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cnf.Name,
			Namespace: cnf.Namespace,
			Labels:    map[string]string{"lxdvirtualmachine": cnf.Name},
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "kvm",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/dev/kvm",
						},
					},
				},
				{
					Name: "snap",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/snap/lxd/common/lxd",
							Type: (*corev1.HostPathType)(&hostPath),
						},
					},
				},
			},
			InitContainers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "snap",
							MountPath: "/var/lib/lxd",
						},
						{
							Name:      "kvm",
							MountPath: "/dev/kvm",
						},
					},
					Name:  "lxd-init",
					Image: "tibbar/talisman:latest",
					Command: []string{
						"/bin/sh", "-c", fmt.Sprintf("sleep 10 && ./lxc launch images:%s $(POD_NAME) --vm || true", cnf.Spec.Image),
					},
				},
			},
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "snap",
							MountPath: "/var/lib/lxd",
						},
						{
							Name:      "kvm",
							MountPath: "/dev/kvm",
						},
					},
					Name:            "lxd-vm",
					Image:           "tibbar/talisman:latest",
					ImagePullPolicy: corev1.PullAlways,
					Command: []string{
						"/bin/bash",
						"-c",
						"sleep 10 && ./lxc exec $(POD_NAME) bash",
					},
				},
			},
		},
	}

	found := &corev1.Pod{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), pod)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := ctrlutil.SetOwnerReference(cnf, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LXDVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&componentsv1alpha1.LXDVirtualMachine{}).
		Complete(r)
}
