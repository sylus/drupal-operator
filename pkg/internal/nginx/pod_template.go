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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	nginxHTTPPort  = 80
	nginxHTTPSPort = 443
)

var (
	wwwDataUserID int64 = 33
)

func (droplet *Nginx) image() string {
	return fmt.Sprintf("%s:%s", defaultImage, defaultTag)
}

func (droplet *Nginx) env() []corev1.EnvVar {
	out := append([]corev1.EnvVar{
		{
			Name:  "NGINX_HOST",
			Value: fmt.Sprintf("http://%s", droplet.Spec.Domains[0]),
		},
	}, droplet.Spec.Nginx.Env...)

	return out
}

func (droplet *Nginx) envFrom() []corev1.EnvFromSource {
	out := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: droplet.ComponentName(NginxSecret),
				},
			},
		},
	}

	out = append(out, droplet.Spec.Nginx.EnvFrom...)

	return out
}

func (droplet *Nginx) volumeMounts() (out []corev1.VolumeMount) {
	out = droplet.Spec.Nginx.VolumeMounts

	out = append(out, corev1.VolumeMount{
		Name:      "cm-nginx",
		MountPath: "/etc/nginx/nginx.conf",
		ReadOnly:  false,
		SubPath:   "nginx.conf",
	})

	return out
}

func (droplet *Nginx) configMap() corev1.Volume {
	configMap := corev1.Volume{
		Name: "cm-nginx",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-%s", droplet.ComponentName(NginxConfigMap), "nginx"),
				},
			},
		},
	}

	return configMap
}

func (droplet *Nginx) volumes() []corev1.Volume {
	return append(droplet.Spec.Nginx.Volumes, droplet.configMap())
}

// PodTemplateSpec generates a pod template spec suitable for use with Nginx
func (droplet *Nginx) PodTemplateSpec() (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = droplet.PodLabels()

	out.Spec.ImagePullSecrets = droplet.Spec.Nginx.ImagePullSecrets
	if len(droplet.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = droplet.Spec.ServiceAccountName
	}

	out.Spec.Containers = []corev1.Container{
		{
			Name:         "nginx",
			Image:        droplet.image(),
			VolumeMounts: droplet.volumeMounts(),
			Env:          droplet.env(),
			EnvFrom:      droplet.envFrom(),
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: int32(nginxHTTPPort),
				},
				{
					Name:          "https",
					ContainerPort: int32(nginxHTTPSPort),
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("400m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				},
			},
		},
	}

	out.Spec.Volumes = droplet.volumes()

	out.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup: &wwwDataUserID,
	}

	return out
}
