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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/common"
	"github.com/sylus/drupal-operator/pkg/internal/drupal"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

// NewCodePVCSyncer returns a new sync.Interface for reconciling codePVC
func NewCodePVCSyncer(droplet *drupal.Drupal, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(drupal.DrupalCodePVC)

	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.ComponentName(drupal.DrupalCodePVC),
			Namespace: droplet.Namespace,
		},
	}
	return syncer.NewObjectSyncer("CodePVC", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.PersistentVolumeClaim)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		if droplet.Spec.Drupal.CodeVolumeSpec == nil || droplet.Spec.Drupal.CodeVolumeSpec.PersistentVolumeClaim == nil {
			return fmt.Errorf(".spec.code.persistentVolumeClaim is not defined")
		}

		// PVC spec is immutable
		if !reflect.DeepEqual(out.Spec, corev1.PersistentVolumeClaimSpec{}) {
			return nil
		}

		out.Spec = *droplet.Spec.Drupal.CodeVolumeSpec.PersistentVolumeClaim

		return nil
	})
}
