/*
Copyright 2021.

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	netdevv1alpha1 "github.com/zshi-redhat/network-device-operator/api/v1alpha1"
)

// NetDevicePoolReconciler reconciles a NetDevicePool object
type NetDevicePoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=netdev.k8s.cncf.io,resources=netdevicepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netdev.k8s.cncf.io,resources=netdevicepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netdev.k8s.cncf.io,resources=netdevicepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NetDevicePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *NetDevicePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("netdevicepool", req.NamespacedName)
	logger.Info("Reconciling")

	pList := &netdevv1alpha1.NetDevicePoolList{}
	err := r.List(context.TODO(), pList, &client.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	logger.Info("NetDevicePoolList", pList)

	nodeList := &corev1.NodeList{}
	nodeMatchLabels := &client.MatchingLabels{
		"node-role.kubernetes.io/worker": "",
		"beta.kubernetes.io/os":          "linux",
	}
	err = r.List(context.TODO(), nodeList, nodeMatchLabels)
	if err != nil {
		logger.Error(err, "Failed to list node")
		return ctrl.Result{}, err
	}

	err = r.syncNodeDeviceConfigMaps(nodeList)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetDevicePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netdevv1alpha1.NetDevicePool{}).
		Complete(r)
}

func (r *NetDevicePoolReconciler) syncNodeDeviceConfigMaps(nodeList *corev1.NodeList) error {
	logger := r.Log.WithValues("syncNodeDeviceConfigMap")
	logger.Info("start")

	found := &corev1.ConfigMap{}
	for _, node := range nodeList.Items {
		name := node.GetName() + "-netdev-states"
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				state := &corev1.ConfigMap{}
				state.Name = name
				state.Namespace = namespace
				err = r.Create(context.TODO(), state)
				if err != nil {
					return fmt.Errorf("Failed to create network device node state configmap: %v", err)
				}
				logger.Info("created network device node state configmap", state.Namespace, state.Name)
			} else {
				return fmt.Errorf("Failed to get network device node state configmap: %s/%s", namespace, name)
			}
		} else {
			logger.Info("update network device node state configmap")
			newState := found.DeepCopy()
			err = r.Update(context.TODO(), newState)
			if err != nil {
				return fmt.Errorf("Failed to update network device node state configmap: %s/%s", namespace, name)
			}
		}
	}
	return nil
}
