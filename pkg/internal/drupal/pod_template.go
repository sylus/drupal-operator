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

	corev1 "k8s.io/api/core/v1"
)

//nolint
const (
	gitCloneImage   = "docker.io/library/buildpack-deps:stretch-scm"
	drupalPort      = 9000
	codeVolumeName  = "code"
	mediaVolumeName = "media"
)

const gitCloneScript = `#!/bin/bash
set -e
set -o pipefail

export HOME="$(mktemp -d)"
export GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=$HOME/.ssh/known_hosts -o StrictHostKeyChecking=no"

test -d "$HOME/.ssh" || mkdir "$HOME/.ssh"

if [ ! -z "$SSH_RSA_PRIVATE_KEY" ] ; then
    echo "$SSH_RSA_PRIVATE_KEY" > "$HOME/.ssh/id_rsa"
    chmod 0400 "$HOME/.ssh/id_rsa"
    export GIT_SSH_COMMAND="$GIT_SSH_COMMAND -o IdentityFile=$HOME/.ssh/id_rsa"
fi

if [ -z "$GIT_CLONE_URL" ] ; then
    echo "No \$GIT_CLONE_URL specified" >&2
    exit 1
fi

find "$SRC_DIR" -maxdepth 1 -mindepth 1 -print0 | xargs -0 /bin/rm -rf

set -x
git clone "$GIT_CLONE_URL" "$SRC_DIR"
cd "$SRC_DIR"
git checkout -B "$GIT_CLONE_REF" "$GIT_CLONE_REF"
`

var (
	wwwDataUserID int64 = 33
)

var (
	s3EnvVars = map[string]string{
		"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
		"AWS_CONFIG_FILE":       "AWS_CONFIG_FILE",
		"ENDPOINT":              "S3_ENDPOINT",
	}
	gcsEnvVars = map[string]string{
		"GOOGLE_CREDENTIALS":             "GOOGLE_CREDENTIALS",
		"GOOGLE_APPLICATION_CREDENTIALS": "GOOGLE_APPLICATION_CREDENTIALS",
	}
)

func (droplet *Drupal) image() string {
	return fmt.Sprintf("%s:%s", droplet.Spec.Image, droplet.Spec.Tag)
}

func (droplet *Drupal) env() []corev1.EnvVar {
	out := append([]corev1.EnvVar{
		{
			Name:  "DRUPAL_HOME",
			Value: fmt.Sprintf("http://%s", droplet.Spec.Domains[0]),
		},
		{
			Name:  "DRUPAL_SITEURL",
			Value: fmt.Sprintf("http://%s/droplet", droplet.Spec.Domains[0]),
		},
	}, droplet.Spec.Env...)

	if droplet.Spec.MediaVolumeSpec != nil {
		if droplet.Spec.MediaVolumeSpec.S3VolumeSource != nil {
			for _, env := range droplet.Spec.MediaVolumeSpec.S3VolumeSource.Env {
				if name, ok := s3EnvVars[env.Name]; ok {
					_env := env.DeepCopy()
					_env.Name = name
					out = append(out, *_env)
				}
			}
		}

		if droplet.Spec.MediaVolumeSpec.GCSVolumeSource != nil {
			out = append(out, corev1.EnvVar{
				Name:  "MEDIA_BUCKET",
				Value: fmt.Sprintf("gs://%s", droplet.Spec.MediaVolumeSpec.GCSVolumeSource.Bucket),
			})
			out = append(out, corev1.EnvVar{
				Name:  "MEDIA_BUCKET_PREFIX",
				Value: droplet.Spec.MediaVolumeSpec.GCSVolumeSource.PathPrefix,
			})
			for _, env := range droplet.Spec.MediaVolumeSpec.GCSVolumeSource.Env {
				if name, ok := gcsEnvVars[env.Name]; ok {
					_env := env.DeepCopy()
					_env.Name = name
					out = append(out, *_env)
				}
			}
		}
	}
	return out
}

func (droplet *Drupal) envFrom() []corev1.EnvFromSource {
	out := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-%s", droplet.ComponentName(DrupalSecret), "mysql-root-password"),
				},
			},
		},
	}

	out = append(out, droplet.Spec.EnvFrom...)

	return out
}

func (droplet *Drupal) gitCloneEnv() []corev1.EnvVar {
	if droplet.Spec.CodeVolumeSpec.GitDir == nil {
		return []corev1.EnvVar{}
	}

	out := []corev1.EnvVar{
		{
			Name:  "GIT_CLONE_URL",
			Value: droplet.Spec.CodeVolumeSpec.GitDir.Repository,
		},
		{
			Name:  "SRC_DIR",
			Value: codeSrcMountPath,
		},
	}

	if len(droplet.Spec.CodeVolumeSpec.GitDir.GitRef) > 0 {
		out = append(out, corev1.EnvVar{
			Name:  "GIT_CLONE_REF",
			Value: droplet.Spec.CodeVolumeSpec.GitDir.GitRef,
		})
	}

	out = append(out, droplet.Spec.CodeVolumeSpec.GitDir.Env...)

	return out
}

func (droplet *Drupal) volumeMounts() (out []corev1.VolumeMount) {
	out = droplet.Spec.VolumeMounts

	out = append(out, corev1.VolumeMount{
		Name:      "cm-drupal",
		MountPath: "/var/www/html/sites/default/settings.php",
		ReadOnly:  false,
		SubPath:   "d8.settings.php",
	})

	if droplet.Spec.CodeVolumeSpec != nil {
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: codeSrcMountPath,
			ReadOnly:  droplet.Spec.CodeVolumeSpec.ReadOnly,
		})
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: droplet.Spec.CodeVolumeSpec.MountPath,
			ReadOnly:  droplet.Spec.CodeVolumeSpec.ReadOnly,
			SubPath:   droplet.Spec.CodeVolumeSpec.ContentSubPath,
		})
	}
	return out
}

func (droplet *Drupal) codeVolume() corev1.Volume {
	drupalCodeVolume := corev1.Volume{
		Name: codeVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	if droplet.Spec.CodeVolumeSpec != nil {
		switch {
		case droplet.Spec.CodeVolumeSpec.GitDir != nil:
			if droplet.Spec.CodeVolumeSpec.GitDir.EmptyDir != nil {
				drupalCodeVolume.EmptyDir = droplet.Spec.CodeVolumeSpec.GitDir.EmptyDir
			}
		case droplet.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil:
			drupalCodeVolume = corev1.Volume{
				Name: codeVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: droplet.ComponentName(DrupalCodePVC),
					},
				},
			}
		case droplet.Spec.CodeVolumeSpec.HostPath != nil:
			drupalCodeVolume = corev1.Volume{
				Name: codeVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: droplet.Spec.CodeVolumeSpec.HostPath,
				},
			}
		case droplet.Spec.CodeVolumeSpec.EmptyDir != nil:
			drupalCodeVolume.EmptyDir = droplet.Spec.CodeVolumeSpec.EmptyDir
		}
	}

	return drupalCodeVolume
}

func (droplet *Drupal) mediaVolume() corev1.Volume {
	drupalMediaVolume := corev1.Volume{
		Name: mediaVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	if droplet.Spec.MediaVolumeSpec != nil {
		switch {
		case droplet.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil:
			drupalMediaVolume = corev1.Volume{
				Name: mediaVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: droplet.ComponentName(DrupalMediaPVC),
					},
				},
			}
		case droplet.Spec.MediaVolumeSpec.HostPath != nil:
			drupalMediaVolume = corev1.Volume{
				Name: mediaVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: droplet.Spec.MediaVolumeSpec.HostPath,
				},
			}
		case droplet.Spec.MediaVolumeSpec.EmptyDir != nil:
			drupalMediaVolume.EmptyDir = droplet.Spec.MediaVolumeSpec.EmptyDir
		}
	}

	return drupalMediaVolume
}

func (droplet *Drupal) configMap() corev1.Volume {
	configMap := corev1.Volume{
		Name: "cm-drupal",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: droplet.ComponentName(DrupalConfigMap),
				},
			},
		},
	}

	return configMap
}

func (droplet *Drupal) volumes() []corev1.Volume {
	return append(droplet.Spec.Volumes, droplet.configMap(), droplet.codeVolume(), droplet.mediaVolume())
}

func (droplet *Drupal) gitCloneContainer() corev1.Container {
	return corev1.Container{
		Name:    "git",
		Args:    []string{"/bin/bash", "-c", gitCloneScript},
		Image:   gitCloneImage,
		Env:     droplet.gitCloneEnv(),
		EnvFrom: droplet.Spec.CodeVolumeSpec.GitDir.EnvFrom,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      codeVolumeName,
				MountPath: codeSrcMountPath,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &wwwDataUserID,
		},
	}
}

// PodTemplateSpec generates a pod template spec suitable for use with Drupal
func (droplet *Drupal) PodTemplateSpec() (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = droplet.PodLabels()

	out.Spec.ImagePullSecrets = droplet.Spec.ImagePullSecrets
	if len(droplet.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = droplet.Spec.ServiceAccountName
	}

	if droplet.Spec.CodeVolumeSpec != nil && droplet.Spec.CodeVolumeSpec.GitDir != nil {
		out.Spec.InitContainers = []corev1.Container{
			droplet.gitCloneContainer(),
		}
	}

	out.Spec.Containers = []corev1.Container{
		{
			Name:         "drupal",
			Image:        droplet.image(),
			VolumeMounts: droplet.volumeMounts(),
			Env:          droplet.env(),
			EnvFrom:      droplet.envFrom(),
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: int32(drupalPort),
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

// JobPodTemplateSpec generates a pod template spec suitable for use with Drupal jobs
func (droplet *Drupal) JobPodTemplateSpec(cmd ...string) (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = droplet.JobPodLabels()

	out.Spec.ImagePullSecrets = droplet.Spec.ImagePullSecrets
	if len(droplet.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = droplet.Spec.ServiceAccountName
	}

	out.Spec.RestartPolicy = corev1.RestartPolicyNever

	if droplet.Spec.CodeVolumeSpec != nil && droplet.Spec.CodeVolumeSpec.GitDir != nil {
		out.Spec.InitContainers = []corev1.Container{
			droplet.gitCloneContainer(),
		}
	}

	out.Spec.Containers = []corev1.Container{
		{
			Name:         "droplet-cli",
			Image:        droplet.image(),
			Args:         cmd,
			VolumeMounts: droplet.volumeMounts(),
			Env:          droplet.env(),
			EnvFrom:      droplet.envFrom(),
		},
	}

	out.Spec.Volumes = droplet.volumes()

	out.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup: &wwwDataUserID,
	}

	return out
}
