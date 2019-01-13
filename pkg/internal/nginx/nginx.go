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

package nginx

import (
	"fmt"

	"github.com/cooleo/slugify"
	"k8s.io/apimachinery/pkg/labels"

	drupalv1beta1 "github.com/sylus/drupal-operator/pkg/apis/drupal/v1beta1"
)

// Nginx embeds drupalv1beta1.Droplet and adds utility functions
type Nginx struct {
	*drupalv1beta1.Droplet
}

type component struct {
	name       string // eg. web, database, cache
	objNameFmt string
	objName    string
}

var (
	// NginxSecret component
	NginxSecret = component{name: "web", objNameFmt: "%s-nginx"}
	// NginxConfigMap component
	NginxConfigMap = component{name: "web", objNameFmt: "%s"}
	// NginxDeployment component
	NginxDeployment = component{name: "web", objNameFmt: "%s"}
	// NginxCron component
	NginxCron = component{name: "cron", objNameFmt: "%s-nginx-cron"}
	// NginxDBUpgrade component
	NginxDBUpgrade = component{name: "upgrade", objNameFmt: "%s-upgrade"}
	// NginxService component
	NginxService = component{name: "web", objNameFmt: "%s"}
	// NginxIngress component
	NginxIngress = component{name: "web", objNameFmt: "%s"}
	// NginxCodePVC component
	NginxCodePVC = component{name: "code", objNameFmt: "%s-code"}
	// NginxMediaPVC component
	NginxMediaPVC = component{name: "media", objNameFmt: "%s-media"}
)

// New wraps a drupalv1beta1.Droplet into a Nginx object
func New(obj *drupalv1beta1.Droplet) *Nginx {
	return &Nginx{obj}
}

// Unwrap returns the wrapped drupalv1beta1.Droplet object
func (o *Nginx) Unwrap() *drupalv1beta1.Droplet {
	return o.Droplet
}

// Labels returns default label set for drupalv1beta1.Nginx
func (o *Nginx) Labels() labels.Set {
	partOf := "drupal"
	if o.ObjectMeta.Labels != nil && len(o.ObjectMeta.Labels["app.kubernetes.io/part-of"]) > 0 {
		partOf = o.ObjectMeta.Labels["app.kubernetes.io/part-of"]
	}

	version := defaultTag
	if len(o.Spec.Nginx.Tag) != 0 {
		version = o.Spec.Nginx.Tag
	}

	labels := labels.Set{
		"app.kubernetes.io/name":     "nginx",
		"app.kubernetes.io/instance": o.ObjectMeta.Name,
		"app.kubernetes.io/version":  version,
		"app.kubernetes.io/part-of":  partOf,
	}

	return labels
}

// ComponentLabels returns labels for a label set for a drupalv1beta1.Nginx component
func (o *Nginx) ComponentLabels(component component) labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = component.name

	if component == NginxDBUpgrade {
		l["drupal.sylus.ca/upgrade-for"] = o.ImageTagVersion()
	}

	return l
}

// ComponentName returns the object name for a component
func (o *Nginx) ComponentName(component component) string {
	name := component.objName
	if len(component.objNameFmt) > 0 {
		name = fmt.Sprintf(component.objNameFmt, o.ObjectMeta.Name)
	}

	if component == NginxDBUpgrade {
		name = fmt.Sprintf("%s-for-%s", name, o.ImageTagVersion())
	}

	return name
}

// ImageTagVersion returns the version from the image tag in a format suitable
// fro kubernetes object names and labels
func (o *Nginx) ImageTagVersion() string {
	return slugify.Slugify(o.Spec.Nginx.Tag)
}

// PodLabels return labels to apply to web pods
func (o *Nginx) PodLabels() labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = "nginx"
	return l
}

// JobPodLabels return labels to apply to cli job pods
func (o *Nginx) JobPodLabels() labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = "nginx-cli"
	return l
}
