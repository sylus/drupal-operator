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

const (
	defaultTag           = "0.0.1"
	defaultImage         = "drupalwxt/site-canada"
	codeSrcMountPath     = "/var/run/sylus.ca/code/src"
	defaultCodeMountPath = "/var/www/site/web/droplet-content"
)

// SetDefaults sets Drupal field defaults
func (o *Drupal) SetDefaults() {
	if len(o.Spec.Image) == 0 {
		o.Spec.Image = defaultImage
	}

	if len(o.Spec.Tag) == 0 {
		o.Spec.Tag = defaultTag
	}

	if o.Spec.CodeVolumeSpec != nil && len(o.Spec.CodeVolumeSpec.MountPath) == 0 {
		o.Spec.CodeVolumeSpec.MountPath = defaultCodeMountPath
	}
}
