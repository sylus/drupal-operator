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
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sylus/drupal-operator/pkg/controller/droplet/internal/sync/common"
	"github.com/sylus/drupal-operator/pkg/internal/nginx"
	"github.com/sylus/drupal-operator/pkg/util/syncer"
)

// NewIngressSyncer returns a new sync.Interface for reconciling Nginx Ingress
func NewIngressSyncer(droplet *nginx.Nginx, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := droplet.ComponentLabels(nginx.NginxIngress)

	obj := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      droplet.ComponentName(nginx.NginxIngress),
			Namespace: droplet.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Ingress", droplet.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*extv1beta1.Ingress)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), common.ControllerLabels)

		if len(out.ObjectMeta.Annotations) == 0 {
			out.ObjectMeta.Annotations = make(map[string]string)
		}
		for k, v := range droplet.Spec.IngressAnnotations {
			out.ObjectMeta.Annotations[k] = v
		}

		bk := extv1beta1.IngressBackend{
			ServiceName: droplet.ComponentName(nginx.NginxService),
			ServicePort: intstr.FromString("http"),
		}
		bkpaths := []extv1beta1.HTTPIngressPath{
			{
				Path:    "/",
				Backend: bk,
			},
		}

		rules := []extv1beta1.IngressRule{}
		for _, d := range droplet.Spec.Domains {
			rules = append(rules, extv1beta1.IngressRule{
				Host: string(d),
				IngressRuleValue: extv1beta1.IngressRuleValue{
					HTTP: &extv1beta1.HTTPIngressRuleValue{
						Paths: bkpaths,
					},
				},
			})
		}
		out.Spec.Rules = rules

		if len(droplet.Spec.TLSSecretRef) > 0 {
			tls := extv1beta1.IngressTLS{
				SecretName: string(droplet.Spec.TLSSecretRef),
			}
			for _, d := range droplet.Spec.Domains {
				tls.Hosts = append(tls.Hosts, string(d))
			}
			out.Spec.TLS = []extv1beta1.IngressTLS{tls}
		} else {
			out.Spec.TLS = nil
		}

		return nil
	})
}
