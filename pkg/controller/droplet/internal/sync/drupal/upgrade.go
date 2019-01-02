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
	batchv1 "k8s.io/api/batch/v1"
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

// NewDBUpgradeJobSyncer returns a new sync.Interface for reconciling database upgrade Job
func NewDBUpgradeJobSyncer(droplet *drupal.Drupal, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(drupal.DrupalDBUpgrade)

	obj := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.ComponentName(drupal.DrupalDBUpgrade),
			Namespace: droplet.Namespace,
		},
	}

	var (
		backoffLimit          int32
		activeDeadlineSeconds int64 = 10
	)

	return syncer.NewObjectSyncer("DBUpgradeJob", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*batchv1.Job)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		if !out.CreationTimestamp.IsZero() {
			return nil
		}

		out.Spec.BackoffLimit = &backoffLimit
		out.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		cmd := []string{"/bin/sh", "-c", "drush updatedb && drush cr all"}
		template := droplet.JobPodTemplateSpec(cmd...)

		out.Spec.Template.ObjectMeta = template.ObjectMeta

		err := mergo.Merge(&out.Spec.Template.Spec, template.Spec, mergo.WithTransformers(transformers.PodSpec))
		if err != nil {
			return err
		}

		return nil
	})
}
