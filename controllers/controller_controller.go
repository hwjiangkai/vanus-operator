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
	"github.com/linkall-labs/vanus-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	vanusv1 "github.com/linkall-labs/vanus-operator/api/v1"
)

// ControllerReconciler reconciles a Controller object
type ControllerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=controllers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=controllers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=controllers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Controller object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := log.Log.WithName("Controller")
	logger.Info("Reconciling Controller.")

	// Fetch the Controller instance
	controller := &vanusv1.Controller{}
	err := r.Get(ctx, req.NamespacedName, controller)
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

	//create headless svc
	// headlessSvc := &corev1.Service{}
	// err = r.Get(ctx, types.NamespacedName{Name: utils.BuildHeadlessSvcResourceName(req.Name), Namespace: req.Namespace}, headlessSvc)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		// create;
	// 		consoleSvc := r.generateHeadlessSvc(controller)
	// 		err = r.Create(ctx, consoleSvc)
	// 		if err != nil {
	// 			logger.Error(err, "Failed to create controller headless svc")
	// 			return reconcile.Result{}, err
	// 		} else {
	// 			logger.Info("Successfully create controller headless svc")
	// 		}
	// 	} else {
	// 		return reconcile.Result{}, err
	// 	}
	// }

	sts := r.getStatefulSetForController(controller)
	// Check if the statefulSet already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Controller StatefulSet.", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		err = r.Create(ctx, sts)
		if err != nil {
			logger.Error(err, "Failed to create new Controller StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		}
	} else if err != nil {
		logger.Error(err, "Failed to get Controller StatefulSet.")
	}

	// List the pods for this controller's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForController(controller.Name))
	listOps := &client.ListOptions{
		Namespace:     controller.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.List(ctx, podList, listOps)
	if err != nil {
		logger.Error(err, "Failed to list pods.", "Controller.Namespace", controller.Namespace, "Controller.Name", controller.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNamesForController(podList.Items)
	logger.Info("controller.Status.Nodes length = " + strconv.Itoa(len(controller.Status.Nodes)))
	logger.Info("podNames length = " + strconv.Itoa(len(podNames)))
	// Ensure every pod is in running phase
	for _, pod := range podList.Items {
		if !reflect.DeepEqual(pod.Status.Phase, corev1.PodRunning) {
			logger.Info("pod " + pod.Name + " phase is " + string(pod.Status.Phase) + ", wait for a moment...")
		}
	}

	// Update status.Size if needed
	if controller.Spec.Replicas != controller.Status.Size {
		logger.Info("controller.Status.Size = " + strconv.Itoa(int(controller.Status.Size)))
		logger.Info("controller.Spec.Size = " + strconv.Itoa(int(controller.Spec.Replicas)))
		controller.Status.Size = controller.Spec.Replicas
		err = r.Status().Update(ctx, controller)
		if err != nil {
			logger.Error(err, "Failed to update Controller Size status.")
		}
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, controller.Status.Nodes) {
		controller.Status.Nodes = podNames
		err = r.Status().Update(ctx, controller)
		if err != nil {
			logger.Error(err, "Failed to update Controller Nodes status.")
		}
	}

	//create svc
	controllerSvc := &corev1.Service{}
	controllerSvcName := utils.BuildSvcResourceName(req.Name)
	err = r.Get(ctx, types.NamespacedName{Name: controllerSvcName, Namespace: req.Namespace}, controllerSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			// create;
			svcToCreate := r.generateSvcForController(controller)
			err = r.Create(ctx, svcToCreate)
			if err != nil {
				logger.Error(err, "Failed to create controller svc")
				return reconcile.Result{}, err
			} else {
				logger.Info("Successfully create controller svc")
			}
		} else {
			return reconcile.Result{}, err
		}
	}
	// todo: only here use, why?
	// maybe use to wait controller ready for others
	cons.ControllerAccessPoint = controllerSvcName + ":2048"

	return ctrl.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1.Controller{}).
		Complete(r)
}

// returns a controller StatefulSet object
func (r *ControllerReconciler) getStatefulSetForController(controller *vanusv1.Controller) *appsv1.StatefulSet {
	ls := labelsForController(controller.Name)

	// After CustomResourceDefinition version upgraded from v1beta1 to v1
	// `controller.spec.VolumeClaimTemplates.metadata` declared in yaml will not be stored by kubernetes.
	// Here is a temporary repair method: to generate a random name
	// if strings.EqualFold(controller.Spec.VolumeClaimTemplates[0].Name, "") {
	// 	// controller.Spec.VolumeClaimTemplates[0].Name = uuid.New().String()
	// 	controller.Spec.VolumeClaimTemplates[0].Name = "data"
	// }

	var replica = controller.Spec.Replicas
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controller.Name,
			Namespace: controller.Namespace,
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
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: controller.Spec.ServiceAccountName,
					// TODO(jiangkai): Supported in the next version
					// Affinity:           controller.Spec.Affinity,
					// Tolerations:        controller.Spec.Tolerations,
					// NodeSelector:       controller.Spec.NodeSelector,
					// PriorityClassName:  controller.Spec.PriorityClassName,
					// ImagePullSecrets:   controller.Spec.ImagePullSecrets,
					Containers: []corev1.Container{{
						Resources: controller.Spec.Resources,
						Image:     controller.Spec.Image,
						Name:      cons.ControllerContainerName,
						// SecurityContext: getContainerSecurityContext(controller),
						ImagePullPolicy: controller.Spec.ImagePullPolicy,
						Env:             getEnvForController(controller),
						// TODO(jiangkai): use const
						Ports: []corev1.ContainerPort{{
							Name:          "grpc",
							ContainerPort: 2048,
						}, {
							Name:          "etcd-client",
							ContainerPort: 2379,
						}, {
							Name:          "etcd-peer",
							ContainerPort: 2380,
						}, {
							Name:          "metrics",
							ContainerPort: 2112,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.ConfigMountPath,
							Name:      "config-controller",
						}, {
							MountPath: cons.StoreMountPath,
							Name:      "data",
						}},
						Command: []string{"/bin/sh", "-c", "NODE_ID=${HOSTNAME##*-} /vanus/bin/controller"},
					}},
					Volumes: getVolumesForController(controller),
					// SecurityContext: getPodSecurityContext(controller),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplatesForController(controller),
		},
	}
	// Set Controller instance as the owner and controller
	controllerutil.SetControllerReference(controller, dep, r.Scheme)

	return dep
}

func getEnvForController(controller *vanusv1.Controller) []corev1.EnvVar {
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

func getVolumeClaimTemplatesForController(controller *vanusv1.Controller) []corev1.PersistentVolumeClaim {
	ls := labelsForController(controller.Name)
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
// func getPodSecurityContext(controller *vanusv1.Controller) *corev1.PodSecurityContext {
// 	var securityContext = corev1.PodSecurityContext{}
// 	if controller.Spec.PodSecurityContext != nil {
// 		securityContext = *controller.Spec.PodSecurityContext
// 	}
// 	return &securityContext
// }

// TODO(jiangkai): Supported in the next version
// func getContainerSecurityContext(controller *vanusv1.Controller) *corev1.SecurityContext {
// 	var securityContext = corev1.SecurityContext{}
// 	if controller.Spec.ContainerSecurityContext != nil {
// 		securityContext = *controller.Spec.ContainerSecurityContext
// 	}
// 	return &securityContext
// }

func getVolumesForController(controller *vanusv1.Controller) []corev1.Volume {
	volumes := []corev1.Volume{{
		Name: "config-controller",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "config-controller",
				},
			}},
	}}
	return volumes
}

// labelsForController returns the labels for selecting the resources
// belonging to the given controller CR name.
func labelsForController(name string) map[string]string {
	return map[string]string{"app": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNamesForController(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *ControllerReconciler) generateSvcForController(cr *vanusv1.Controller) *corev1.Service {
	controllerSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cr.Namespace,
			Name:       cr.Name,
			Labels:     labelsForController(cr.Name),
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			Selector: labelsForController(cr.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       cr.Name,
					Port:       2048,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(2048),
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, controllerSvc, r.Scheme)
	return controllerSvc
}
