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

package drupal

import (
	"fmt"

	"github.com/cooleo/slugify"
	"k8s.io/apimachinery/pkg/labels"

	drupalv1beta1 "github.com/sylus/drupal-operator/pkg/apis/drupal/v1beta1"
)

// Drupal embeds drupalv1beta1.Droplet and adds utility functions
type Drupal struct {
	*drupalv1beta1.Droplet
}

type component struct {
	name       string // eg. web, database, cache
	objNameFmt string
	objName    string
}

var (
	// DrupalSecret component
	DrupalSecret = component{name: "web", objNameFmt: "%s-drupal"}
	// DrupalConfigMap component
	DrupalConfigMap = component{name: "web", objNameFmt: "%s"}
	// DrupalDeployment component
	DrupalDeployment = component{name: "web", objNameFmt: "%s"}
	// DrupalCron component
	DrupalCron = component{name: "cron", objNameFmt: "%s-drupal-cron"}
	// DrupalDBUpgrade component
	DrupalDBUpgrade = component{name: "upgrade", objNameFmt: "%s-upgrade"}
	// DrupalService component
	DrupalService = component{name: "web", objNameFmt: "%s"}
	// DrupalIngress component
	DrupalIngress = component{name: "web", objNameFmt: "%s"}
	// DrupalCodePVC component
	DrupalCodePVC = component{name: "code", objNameFmt: "%s-code"}
	// DrupalMediaPVC component
	DrupalMediaPVC = component{name: "media", objNameFmt: "%s-media"}
)

// New wraps a drupalv1beta1.Droplet into a Drupal object
func New(obj *drupalv1beta1.Droplet) *Drupal {
	return &Drupal{obj}
}

// Unwrap returns the wrapped drupalv1beta1.Droplet object
func (o *Drupal) Unwrap() *drupalv1beta1.Droplet {
	return o.Droplet
}

// Labels returns default label set for drupalv1beta1.Drupal
func (o *Drupal) Labels() labels.Set {
	partOf := "drupal"
	if o.ObjectMeta.Labels != nil && len(o.ObjectMeta.Labels["app.kubernetes.io/part-of"]) > 0 {
		partOf = o.ObjectMeta.Labels["app.kubernetes.io/part-of"]
	}

	version := defaultTag
	if len(o.Spec.Drupal.Tag) != 0 {
		version = o.Spec.Drupal.Tag
	}

	labels := labels.Set{
		"app.kubernetes.io/name":     "drupal",
		"app.kubernetes.io/instance": o.ObjectMeta.Name,
		"app.kubernetes.io/version":  version,
		"app.kubernetes.io/part-of":  partOf,
	}

	return labels
}

// ComponentLabels returns labels for a label set for a drupalv1beta1.Drupal component
func (o *Drupal) ComponentLabels(component component) labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = component.name

	if component == DrupalDBUpgrade {
		l["drupal.sylus.ca/upgrade-for"] = o.ImageTagVersion()
	}

	return l
}

// ComponentName returns the object name for a component
func (o *Drupal) ComponentName(component component) string {
	name := component.objName
	if len(component.objNameFmt) > 0 {
		name = fmt.Sprintf(component.objNameFmt, o.ObjectMeta.Name)
	}

	if component == DrupalDBUpgrade {
		name = fmt.Sprintf("%s-for-%s", name, o.ImageTagVersion())
	}

	return name
}

// ImageTagVersion returns the version from the image tag in a format suitable
// for kubernetes object names and labels
func (o *Drupal) ImageTagVersion() string {
	return slugify.Slugify(o.Spec.Drupal.Tag)
}

// PodLabels return labels to apply to web pods
func (o *Drupal) PodLabels() labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = "drupal"
	return l
}

// JobPodLabels return labels to apply to cli job pods
func (o *Drupal) JobPodLabels() labels.Set {
	l := o.Labels()
	l["app.kubernetes.io/component"] = "drupal-cli"
	return l
}

// DataBaseBackend returns the type of database to leverage
func (o *Drupal) DataBaseBackend() (string, string) {
	if o.Spec.Drupal.DatabaseBackEnd == "postgres" {
		return "pgsql", "5432"
	}

	return "msql", "3306"
}
