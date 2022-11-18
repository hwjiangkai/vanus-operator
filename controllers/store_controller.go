/*
Copyright 2022.

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
	"reflect"
	"strconv"
	"time"

	// "github.com/google/uuid"

	cons "github.com/linkall-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	vanusv1 "github.com/linkall-labs/vanus-operator/api/v1"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=stores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=stores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=stores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Store object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := log.Log.WithName("Store")
	logger.Info("Reconciling Store.")

	// Fetch the Store instance
	store := &vanusv1.Store{}
	err := r.Get(ctx, req.NamespacedName, store)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Store resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Store.")
		return reconcile.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	sts := r.getStatefulSetForStore(store)
	// Check if the statefulSet already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Store StatefulSet.", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		err = r.Create(ctx, sts)
		if err != nil {
			logger.Error(err, "Failed to create new Store StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		}
	} else if err != nil {
		logger.Error(err, "Failed to get Store StatefulSet.")
	}

	// List the pods for this store's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForStore(store.Name))
	listOps := &client.ListOptions{
		Namespace:     store.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.List(ctx, podList, listOps)
	if err != nil {
		logger.Error(err, "Failed to list pods.", "Store.Namespace", store.Namespace, "Store.Name", store.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNamesForStore(podList.Items)
	logger.Info("store.Status.Nodes length = " + strconv.Itoa(len(store.Status.Nodes)))
	logger.Info("podNames length = " + strconv.Itoa(len(podNames)))
	// Ensure every pod is in running phase
	for _, pod := range podList.Items {
		if !reflect.DeepEqual(pod.Status.Phase, corev1.PodRunning) {
			logger.Info("pod " + pod.Name + " phase is " + string(pod.Status.Phase) + ", wait for a moment...")
		}
	}

	// Update status.Size if needed
	if store.Spec.Replicas != store.Status.Size {
		logger.Info("store.Status.Size = " + strconv.Itoa(int(store.Status.Size)))
		logger.Info("store.Spec.Size = " + strconv.Itoa(int(store.Spec.Replicas)))
		store.Status.Size = store.Spec.Replicas
		err = r.Status().Update(ctx, store)
		if err != nil {
			logger.Error(err, "Failed to update Store Size status.")
		}
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, store.Status.Nodes) {
		store.Status.Nodes = podNames
		err = r.Status().Update(ctx, store)
		if err != nil {
			logger.Error(err, "Failed to update Store Nodes status.")
		}
	}

	return ctrl.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

// SetupWithManager sets up the store with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1.Store{}).
		Complete(r)
}

// returns a store StatefulSet object
func (r *StoreReconciler) getStatefulSetForStore(store *vanusv1.Store) *appsv1.StatefulSet {
	ls := labelsForStore(store.Name)
	an := annotationsForStore()
	// After CustomResourceDefinition version upgraded from v1beta1 to v1
	// `store.spec.VolumeClaimTemplates.metadata` declared in yaml will not be stored by kubernetes.
	// Here is a temporary repair method: to generate a random name
	// if strings.EqualFold(store.Spec.VolumeClaimTemplates[0].Name, "") {
	// 	// store.Spec.VolumeClaimTemplates[0].Name = uuid.New().String()
	// 	store.Spec.VolumeClaimTemplates[0].Name = "data"
	// }

	var replica = store.Spec.Replicas
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      store.Name,
			Namespace: store.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      ls,
					Annotations: an,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: store.Spec.ServiceAccountName,
					// TODO(jiangkai): Supported in the next version
					// Affinity:           store.Spec.Affinity,
					// Tolerations:        store.Spec.Tolerations,
					// NodeSelector:       store.Spec.NodeSelector,
					// PriorityClassName:  store.Spec.PriorityClassName,
					// ImagePullSecrets:   store.Spec.ImagePullSecrets,
					Containers: []corev1.Container{{
						Resources: store.Spec.Resources,
						Image:     store.Spec.Image,
						Name:      cons.StoreContainerName,
						// SecurityContext: getContainerSecurityContext(store),
						ImagePullPolicy: store.Spec.ImagePullPolicy,
						Env:             getENVForStore(store),
						// TODO(jiangkai): use const
						Ports: []corev1.ContainerPort{{
							Name:          "grpc",
							ContainerPort: 11811,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.ConfigMountPath,
							Name:      "config-store",
						}, {
							MountPath: cons.StoreMountPath,
							Name:      "data",
						}},
						Command: []string{"/bin/sh", "-c", "NODE_ID=${HOSTNAME##*-} /vanus/bin/store"},
					}},
					Volumes: getVolumesForStore(store),
					// SecurityContext: getPodSecurityContext(store),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplatesForStore(store),
		},
	}
	// Set Store instance as the owner and store
	controllerutil.SetControllerReference(store, dep, r.Scheme)

	return dep
}

func getENVForStore(store *vanusv1.Store) []corev1.EnvVar {
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

func getVolumeClaimTemplatesForStore(store *vanusv1.Store) []corev1.PersistentVolumeClaim {
	ls := labelsForStore(store.Name)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse("1Gi")
	pvcs := []corev1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{
			Labels: ls,
			Name:   "data",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: requests,
			},
		},
	}}
	return pvcs
}

// TODO(jiangkai): Supported in the next version
// func getPodSecurityContext(store *vanusv1.Store) *corev1.PodSecurityContext {
// 	var securityContext = corev1.PodSecurityContext{}
// 	if store.Spec.PodSecurityContext != nil {
// 		securityContext = *store.Spec.PodSecurityContext
// 	}
// 	return &securityContext
// }

// TODO(jiangkai): Supported in the next version
// func getContainerSecurityContext(store *vanusv1.Store) *corev1.SecurityContext {
// 	var securityContext = corev1.SecurityContext{}
// 	if store.Spec.ContainerSecurityContext != nil {
// 		securityContext = *store.Spec.ContainerSecurityContext
// 	}
// 	return &securityContext
// }

func getVolumesForStore(store *vanusv1.Store) []corev1.Volume {
	volumes := []corev1.Volume{{
		Name: "config-store",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "config-store",
				},
			}},
	}}
	return volumes
}

// labelsForStore returns the labels for selecting the resources
// belonging to the given store CR name.
func labelsForStore(name string) map[string]string {
	return map[string]string{"app": name}
}

// labelsForStore returns the labels for selecting the resources
// belonging to the given store CR name.
func annotationsForStore() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": "2112"}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNamesForStore(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
