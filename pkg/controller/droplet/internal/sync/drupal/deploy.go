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

package sync

import (
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/imdario/mergo"

	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/common"
	"github.com/sylus/drupal-operator/pkg/internal/drupal"
	"github.com/sylus/drupal-operator/pkg/util/mergo/transformers"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

// NewDeploymentSyncer returns a new sync.Interface for reconciling web Deployment
func NewDeploymentSyncer(droplet *drupal.Drupal, secret *corev1.Secret, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(drupal.DrupalDeployment)

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.ComponentName(drupal.DrupalDeployment),
			Namespace: droplet.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*appsv1.Deployment)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		template := droplet.PodTemplateSpec()

		if len(template.Annotations) == 0 {
			template.Annotations = make(map[string]string)
		}
		template.Annotations["drupal.sylus.ca/secretVersion"] = secret.ResourceVersion

		out.Spec.Template.ObjectMeta = template.ObjectMeta

		selector := metav1.SetAsLabelSelector(droplet.PodLabels())
		if !reflect.DeepEqual(selector, out.Spec.Selector) {
			if out.ObjectMeta.CreationTimestamp.IsZero() {
				out.Spec.Selector = selector
			} else {
				return fmt.Errorf("deployment selector is immutable")
			}
		}

		err := mergo.Merge(&out.Spec.Template.Spec, template.Spec, mergo.WithTransformers(transformers.PodSpec))
		if err != nil {
			return err
		}

		if droplet.Spec.Replicas != nil {
			out.Spec.Replicas = droplet.Spec.Replicas
		}

		return nil
	})
}
