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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/common"
	"github.com/sylus/drupal-operator/pkg/internal/drupal"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

const (
	drupalHTTPPort = 9000
)

// NewServiceSyncer returns a new sync.Interface for reconciling web Service
func NewServiceSyncer(droplet *drupal.Drupal, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(drupal.DrupalDeployment)

	obj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.Name,
			Namespace: droplet.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Service", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.Service)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		selector := droplet.PodLabels()
		if !labels.Equals(selector, out.Spec.Selector) {
			if out.ObjectMeta.CreationTimestamp.IsZero() {
				out.Spec.Selector = selector
			} else {
				return fmt.Errorf("service selector is immutable")
			}
		}

		if len(out.Spec.Ports) != 1 {
			out.Spec.Ports = make([]corev1.ServicePort, 1)
		}

		out.Spec.Ports[0].Name = "http"
		out.Spec.Ports[0].Port = int32(9000)
		out.Spec.Ports[0].TargetPort = intstr.FromInt(drupalHTTPPort)

		return nil
	})
}
