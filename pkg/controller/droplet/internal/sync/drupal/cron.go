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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
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

// NewDrupalCronSyncer returns a new sync.Interface for reconciling droplet-cron CronJob
func NewDrupalCronSyncer(droplet *drupal.Drupal, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(drupal.DrupalCron)

	obj := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.ComponentName(drupal.DrupalCron),
			Namespace: droplet.Namespace,
		},
	}

	var (
		cronStartingDeadlineSeconds int64 = 10
		backoffLimit                int32
		activeDeadlineSeconds       int64 = 10
		successfulJobsHistoryLimit  int32 = 3
		failedJobsHistoryLimit      int32 = 1
	)

	return syncer.NewObjectSyncer("DrupalCron", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*batchv1beta1.CronJob)

		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		out.Spec.Schedule = "* * * * *"
		out.Spec.ConcurrencyPolicy = "Forbid"
		out.Spec.StartingDeadlineSeconds = &cronStartingDeadlineSeconds
		out.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit
		out.Spec.FailedJobsHistoryLimit = &failedJobsHistoryLimit

		out.Spec.JobTemplate.ObjectMeta.Labels = labels.Merge(objLabels, common.ControllerLabels)
		out.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
		out.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		cmd := []string{"crond", "-f", "-d", "8"}
		template := droplet.JobPodTemplateSpec(cmd...)

		out.Spec.JobTemplate.Spec.Template.ObjectMeta = template.ObjectMeta

		err := mergo.Merge(&out.Spec.JobTemplate.Spec.Template.Spec, template.Spec, mergo.WithTransformers(transformers.PodSpec))

		if err != nil {
			return err
		}

		return nil
	})
}
