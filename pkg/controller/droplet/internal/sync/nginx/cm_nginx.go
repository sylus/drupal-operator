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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/common"
	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/templates"
	"github.com/sylus/drupal-operator/pkg/internal/nginx"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

// Settings spec
type Settings struct {
	Domain     string
	Host       string
	s3BasePath string
	Resolver   string
}

// NewConfigMapSyncer returns a new sync.Interface for reconciling Nginx ConfigMap
func NewConfigMapSyncer(droplet *nginx.Nginx, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(nginx.NginxConfigMap)

	templateInput := Settings{
		Domain:     "drupal.innovation.cloud.statcan.ca",
		Host:       "mysite",
		s3BasePath: "stcdrupal.blob.core.windows.net/drupal-public",
		Resolver:   "10.0.0.10",
	}
	configMap := common.GenerateConfig(templateInput, templates.ConfigMapNginx)

	obj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", droplet.ComponentName(nginx.NginxConfigMap), "nginx"),
			Namespace: droplet.Namespace,
		},
		Data: map[string]string{
			"nginx.conf": *configMap,
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.ConfigMap)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		return nil
	})
}
