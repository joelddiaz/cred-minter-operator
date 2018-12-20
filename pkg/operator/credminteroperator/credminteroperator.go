/*
Copyright 2018 The Kubernetes Authors.

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

package credminteroperator

import (
	"context"
	"fmt"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"

	////apiextensionsclientv1beta1 "github.com/kubernetes/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	credminterv1alpha1 "github.com/openshift/cred-minter-operator/pkg/apis/credminter/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apistuff "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	////apistuff "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/types/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	kubeclient "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"

	////credminterv1 "github.com/openshift/cred-minter-operator/pkg/apis/credminter/v1alpha1"
	"github.com/openshift/cred-minter-operator/pkg/operator/assets"
	////"github.com/openshift/library-go/pkg/operator/v1helpers"
)

const (
	operatorNamespace            = "openshift-cred-minter-operator"
	credMinterNamespace          = "openshift-cred-minter"
	credMinterCustomResourceYAML = "config/cred-minter-yaml/crds/credminter_v1beta1_credentialsrequest.yaml"
	clusterRoleYAML              = "config/cred-minter-yaml/rbac/rbac_role.yaml"
	clusterRoleBindingYAML       = "config/cred-minter-yaml/rbac/rbac_role_binding.yaml"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CredMinterOperatorConfig Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this credminter.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	log.Info("Adding cred-minter operator to manager")
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCredMinterOperatorConfig{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// CONFIG INSTANCE?
	_, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	/*
		v1helpers.EnsureOperatorConfigExists(
			dynamicClient,
			assets.MustAsset("config/operator-config.yaml"),
			schema.GroupVersionResource{
				Group:    credminterv1.SchemeGroupVersion.Group,
				Version:  "v1alpha1",
				Resource: "credminteroperatorconfigs",
			},
		)
	*/

	credminterConfigReconciler := r.(*ReconcileCredMinterOperatorConfig)

	credminterConfigReconciler.imagePullSpec = os.Getenv("IMAGE")
	if len(credminterConfigReconciler.imagePullSpec) == 0 {
		log.Warn("no IMAGE specified, using bleeding edge")
		credminterConfigReconciler.imagePullSpec = "quay.io/dgoodwin/cred-minter:latest"
	}

	credminterConfigReconciler.kubeClient, err = kubeclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("error creating kubeClient: %v", err)
	}

	credminterConfigReconciler.eventRecorder = setupEventRecorder(credminterConfigReconciler.kubeClient)

	credminterConfigReconciler.apiExtClient, err = apistuff.NewForConfig(mgr.GetConfig())
	////credminterConfigReconciler.apiExtClient, err = apiextensionsclientv1beta1.NewForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("error creating apiExtensionClient: %v", err)
	}

	// Create a new controller
	c, err := controller.New("credminteroperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to CredMinterOperatorConfig
	err = c.Watch(&source.Kind{Type: &credminterv1alpha1.CredMinterOperatorConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &credminterv1alpha1.CredMinterOperatorConfig{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCredMinterOperatorConfig{}

// ReconcileCredMinterOperatorConfig reconciles a CredMinterOperatorConfig object
type ReconcileCredMinterOperatorConfig struct {
	client.Client
	scheme       *runtime.Scheme
	kubeClient   kubeclient.Interface
	apiExtClient apistuff.ApiextensionsV1beta1Interface
	////apiExtClient  apiextensionsclientv1beta1.ApiextensionsV1beta1Interface
	eventRecorder events.Recorder
	imagePullSpec string
}

// Reconcile reads that state of the cluster for a CredMinterOperatorConfig object and makes changes based on the state read
// and what is in the CredMinterOperatorConfig.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=credminter.operator.openshift.io,resources=credminteroperatorconfigs,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileCredMinterOperatorConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the CredMinterOperatorConfig instance
	instance := &credminterv1alpha1.CredMinterOperatorConfig{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err = r.setupPreReqs(); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Name + "-deployment",
			////Namespace: instance.Namespace,
			Namespace: credMinterNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: r.imagePullSpec,
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCredMinterOperatorConfig) setupPreReqs() error {
	// Create namespace for cred-minter to live in
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: credMinterNamespace,
		},
	}
	_, _, err := resourceapply.ApplyNamespace(r.kubeClient.CoreV1(), r.eventRecorder, namespace)
	if err != nil {
		log.Errorf("error creating namespace: %v", err)
		return err
	}

	// CRD for cred-minter
	crd, _ := assets.Asset(credMinterCustomResourceYAML)
	crdObj := resourceread.ReadCustomResourceDefinitionV1Beta1OrDie(crd)

	_, _, err = resourceapply.ApplyCustomResourceDefinition(r.apiExtClient, r.eventRecorder, crdObj)
	if err != nil {
		log.Errorf("failed to apply CRD: %v", err)
		return err
	}

	// ClusterRole with permissions for cred-minter
	clusterRole, _ := assets.Asset(clusterRoleYAML)
	clusterRoleObj := resourceread.ReadClusterRoleV1OrDie(clusterRole)

	rbacClient := r.kubeClient.RbacV1()
	_, _, err = resourceapply.ApplyClusterRole(rbacClient, r.eventRecorder, clusterRoleObj)
	if err != nil {
		log.Errorf("failed to apply cluster role: %v", err)
		return err
	}

	// Bind the above role to the serviceaccount in the cred-minter namespace
	clusterRoleBinding, _ := assets.Asset(clusterRoleBindingYAML)
	clusterRoleBindingObj := resourceread.ReadClusterRoleBindingV1OrDie(clusterRoleBinding)
	log.Printf("SUBJECT: %+v", clusterRoleBindingObj)

	_, _, err = resourceapply.ApplyClusterRoleBinding(rbacClient, r.eventRecorder, clusterRoleBindingObj)
	if err != nil {
		log.Errorf("failed to apply cluster role binding: %v", err)
		return err
	}

	return nil
}

func setupEventRecorder(kubeClient kubeclient.Interface) events.Recorder {
	controllerRef, err := events.GetControllerReferenceForCurrentPod(kubeClient, operatorNamespace, nil)
	if err != nil {
		log.WithError(err).Warning("Cannot determine pod name for event recorder. Using logger.")
		return getLogRecorder()
	}

	eventsClient := kubeClient.CoreV1().Events(controllerRef.Namespace)
	return events.NewRecorder(eventsClient, operatorNamespace, controllerRef)
}

type logRecorder struct{}

func (logRecorder) Event(reason, message string) {
	log.WithField("reason", reason).Info(message)
}

func (logRecorder) Eventf(reason, messageFmt string, args ...interface{}) {
	log.WithField("reason", reason).Infof(messageFmt, args...)
}

func (logRecorder) Warning(reason, message string) {
	log.WithField("reason", reason).Warning(message)
}

func (logRecorder) Warningf(reason, messageFmt string, args ...interface{}) {
	log.WithField("reason", reason).Warningf(messageFmt, args...)
}

func getLogRecorder() events.Recorder {
	return &logRecorder{}
}
