/*

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

package droplet

import (
	"context"

	drupalv1beta1 "github.com/sylus/drupal-operator/pkg/apis/drupal/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	syncDrupal "github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/drupal"
	syncNginx "github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/nginx"
	"github.com/sylus/drupal-operator/pkg/internal/drupal"
	"github.com/sylus/drupal-operator/pkg/internal/nginx"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

var log = logf.Log.WithName("controller")

/**
 * USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
 * business logic.  Delete these comments after modifying this file.*
 */
const controllerName = "drupal-controller"

// Add creates a new Droplet Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDroplet{Client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetRecorder(controllerName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Droplet
	err = c.Watch(&source.Kind{Type: &drupalv1beta1.Droplet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	subresources := []runtime.Object{
		&appsv1.Deployment{},
		&batchv1beta1.CronJob{},
		&corev1.ConfigMap{},
		&corev1.PersistentVolumeClaim{},
		&corev1.Service{},
		&corev1.Secret{},
		&extv1beta1.Ingress{},
	}

	for _, subresource := range subresources {
		err = c.Watch(&source.Kind{Type: subresource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &drupalv1beta1.Droplet{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDroplet{}

// ReconcileDroplet reconciles a Droplet object
type ReconcileDroplet struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Droplet object and makes changes based on the state read
// and what is in the Droplet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=,resources=configmaps;secrets;services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs;jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=drupal.sylus.ca,resources=droplets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=drupal.sylus.ca,resources=droplets/status,verbs=get;update;patch
func (r *ReconcileDroplet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	droplet := drupal.New(&drupalv1beta1.Droplet{})
	err := r.Get(context.TODO(), request.NamespacedName, droplet.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	r.scheme.Default(droplet.Unwrap())
	droplet.SetDefaults()

	nginx := nginx.New(&drupalv1beta1.Droplet{})
	err = r.Get(context.TODO(), request.NamespacedName, nginx.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	r.scheme.Default(nginx.Unwrap())
	nginx.SetDefaults()

	secretSyncer := syncDrupal.NewSecretSyncer(droplet, r.Client, r.scheme)
	syncers := []syncer.Interface{
		secretSyncer,

		syncDrupal.NewConfigMapSyncer(droplet, r.Client, r.scheme),
		syncDrupal.NewDeploymentSyncer(droplet, secretSyncer.GetObject().(*corev1.Secret), r.Client, r.scheme),
		syncDrupal.NewServiceSyncer(droplet, r.Client, r.scheme),
		syncDrupal.NewDrupalCronSyncer(droplet, r.Client, r.scheme),

		syncNginx.NewConfigMapSyncer(nginx, r.Client, r.scheme),
		syncNginx.NewDeploymentSyncer(nginx, r.Client, r.scheme),
		syncNginx.NewServiceSyncer(nginx, r.Client, r.scheme),
		syncNginx.NewIngressSyncer(nginx, r.Client, r.scheme),
	}

	if droplet.Spec.CodeVolumeSpec != nil && droplet.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, syncDrupal.NewCodePVCSyncer(droplet, r.Client, r.scheme))
	}

	if droplet.Spec.MediaVolumeSpec != nil && droplet.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, syncDrupal.NewMediaPVCSyncer(droplet, r.Client, r.scheme))
	}

	return reconcile.Result{}, r.sync(syncers)
}

func (r *ReconcileDroplet) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.recorder); err != nil {
			log.Error(err, "unable to reconcile with object ")
			return err
		}
	}
	return nil
}
