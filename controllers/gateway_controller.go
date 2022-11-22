// Copyright 2022 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"reflect"
	"time"

	// "github.com/google/uuid"

	cons "github.com/linkall-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	vanusv1 "github.com/linkall-labs/vanus-operator/api/v1"
)

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gateway object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := log.Log.WithName("Gateway")
	logger.Info("Reconciling Gateway.")

	// Fetch the Controller instance
	gateway := &vanusv1.Gateway{}
	err := r.Get(ctx, req.NamespacedName, gateway)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Controller resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Controller.")
		return reconcile.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	gatewayDeployment := r.getDeploymentForGateway(gateway)

	// Set Console instance as the owner and controller
	if err := controllerutil.SetControllerReference(gateway, gatewayDeployment, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: gatewayDeployment.Name, Namespace: gatewayDeployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating Vanus Gateway Deployment", "Namespace", gatewayDeployment, "Name", gatewayDeployment.Name)
		err = r.Create(context.TODO(), gatewayDeployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Support console deployment scaling
	if !reflect.DeepEqual(gateway.Spec.Replicas, found.Spec.Replicas) {
		found.Spec.Replicas = gateway.Spec.Replicas
		err = r.Update(context.TODO(), found)
		if err != nil {
			logger.Error(err, "Failed to update console CR ", "Namespace", found.Namespace, "Name", found.Name)
		} else {
			logger.Info("Successfully updated console CR ", "Namespace", found.Namespace, "Name", found.Name)
		}
	}

	// TODO: update console if name server address changes

	// CR already exists - don't requeue
	logger.Info("Skip reconcile: Vanus Gateway Deployment already exists", "Namespace", found.Namespace, "Name", found.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1.Gateway{}).
		Complete(r)
}

// newDeploymentForCR returns a deployment pod with modifying the ENV
func (r *GatewayReconciler) getDeploymentForGateway(gateway *vanusv1.Gateway) *appsv1.Deployment {
	ls := labelsForGateway(gateway.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gateway.Name,
			Namespace: gateway.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: gateway.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: gateway.Spec.ServiceAccountName,
					Containers: []corev1.Container{{
						Resources:       gateway.Spec.Resources,
						Image:           gateway.Spec.Image,
						Name:            cons.GatewayContainerName,
						ImagePullPolicy: gateway.Spec.ImagePullPolicy,
						Env:             getEnvForGateway(gateway),
						Ports: []corev1.ContainerPort{{
							Name:          "grpc",
							ContainerPort: 2148,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.ConfigMountPath,
							Name:      "config-gateway",
						}},
					}},
					Volumes: getVolumesForGateway(gateway),
				},
			},
		},
	}
	// Set Controller instance as the owner and controller
	controllerutil.SetControllerReference(gateway, dep, r.Scheme)

	return dep
}

// labelsForController returns the labels for selecting the resources
// belonging to the given controller CR name.
func labelsForGateway(name string) map[string]string {
	return map[string]string{"app": name}
}

func getEnvForGateway(gateway *vanusv1.Gateway) []corev1.EnvVar {
	envs := []corev1.EnvVar{{
		Name:      cons.EnvControllerPodIP,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
	}, {
		Name:      cons.EnvControllerPodName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  cons.EnvControllerLogLevel,
		Value: "DEBUG",
	}}
	return envs
}

func getVolumesForGateway(gateway *vanusv1.Gateway) []corev1.Volume {
	volumes := []corev1.Volume{{
		Name: "config-gateway",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "config-gateway",
				},
			}},
	}}
	return volumes
}
