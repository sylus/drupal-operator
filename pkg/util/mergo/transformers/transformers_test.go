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

package transformers_test

import (
	"fmt"
	"math/rand"

	"github.com/imdario/mergo"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sylus/drupal-operator/pkg/util/mergo/transformers"
)

var _ = ginkgo.Describe("PodSpec Transformer", func() {
	var deployment *appsv1.Deployment

	ginkgo.BeforeEach(func() {
		r := rand.Int31()
		name := fmt.Sprintf("depl-%d", r)
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: "main-image",
								Env: []corev1.EnvVar{
									{
										Name:  "TEST",
										Value: "me",
									},
								},
							},
							{
								Name:  "helper",
								Image: "helper-image",
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 80,
										Protocol:      corev1.ProtocolTCP,
									},
									{
										Name:          "prometheus",
										ContainerPort: 9125,
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "code",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
							{
								Name: "media",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				},
			},
		}
	})

	ginkgo.It("removes unused containers", func() {
		newSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "helper",
					Image: "helper-image",
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "prometheus",
							ContainerPort: 9125,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		}

		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(gomega.Equal("helper"))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Ports).To(gomega.HaveLen(2))
	})
	ginkgo.It("allows container rename", func() {
		newSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "new-helper",
					Image: "helper-image",
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "prometheus",
							ContainerPort: 9125,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		}

		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(gomega.Equal("new-helper"))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Ports).To(gomega.HaveLen(2))
	})
	ginkgo.It("allows container image update", func() {
		newSpec := deployment.Spec.Template.Spec.DeepCopy()
		newSpec.Containers[0].Image = "main-image-v2"
		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(gomega.Equal("main"))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal("main-image-v2"))
	})
	ginkgo.It("merges env vars", func() {
		newSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "main-image",
					Env: []corev1.EnvVar{
						{
							Name:  "TEST-2",
							Value: "me-2",
						},
					},
				},
			},
		}

		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(gomega.Equal("TEST-2"))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Value).To(gomega.Equal("me-2"))
	})
	ginkgo.It("merges container ports", func() {
		newSpec := deployment.Spec.Template.Spec.DeepCopy()
		newSpec.Containers[1].Ports = []corev1.ContainerPort{
			{
				Name:          "prometheus",
				ContainerPort: 9125,
				Protocol:      corev1.ProtocolTCP,
			},
		}
		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(2))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[1].Ports).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[1].Ports[0].ContainerPort).To(gomega.Equal(int32(9125)))
	})
	ginkgo.It("allows prepending volume", func() {
		newSpec := deployment.Spec.Template.Spec.DeepCopy()
		newSpec.Volumes = []corev1.Volume{
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "code",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{},
				},
			},
			{
				Name: "media",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Volumes).To(gomega.HaveLen(3))
		gomega.Expect(deployment.Spec.Template.Spec.Volumes[0].Name).To(gomega.Equal(newSpec.Volumes[0].Name))
		gomega.Expect(deployment.Spec.Template.Spec.Volumes[1].Name).To(gomega.Equal(newSpec.Volumes[1].Name))
		gomega.Expect(deployment.Spec.Template.Spec.Volumes[2].Name).To(gomega.Equal(newSpec.Volumes[2].Name))

		gomega.Expect(deployment.Spec.Template.Spec.Volumes[1].EmptyDir).To(gomega.BeNil())
		gomega.Expect(deployment.Spec.Template.Spec.Volumes[1].HostPath).ToNot(gomega.BeNil())
	})
	ginkgo.It("allows replacing volume list", func() {
		newSpec := deployment.Spec.Template.Spec.DeepCopy()
		newSpec.Volumes = []corev1.Volume{
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
		gomega.Expect(mergo.Merge(&deployment.Spec.Template.Spec, newSpec, mergo.WithTransformers(transformers.PodSpec))).To(gomega.Succeed())
		gomega.Expect(deployment.Spec.Template.Spec.Volumes).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Volumes[0].Name).To(gomega.Equal(newSpec.Volumes[0].Name))
	})
})
