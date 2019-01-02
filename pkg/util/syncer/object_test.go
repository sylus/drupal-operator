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

package syncer_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"golang.org/x/net/context"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	. "github.com/sylus/drupal-operator/pkg/util/syncer"
)

var _ = ginkgo.Describe("ObjectSyncer", func() {
	var syncer *ObjectSyncer
	var deployment *appsv1.Deployment
	var recorder *record.FakeRecorder
	key := types.NamespacedName{Name: "example", Namespace: "default"}

	ginkgo.BeforeEach(func() {
		deployment = &appsv1.Deployment{}
		recorder = record.NewFakeRecorder(100)
	})

	ginkgo.AfterEach(func() {
		// nolint: errcheck
		c.Delete(context.TODO(), deployment)
	})

	ginkgo.When("syncing", func() {
		ginkgo.It("successfully creates an ownerless object when owner is nil", func() {
			syncer = NewDeploymentSyncer(nil).(*ObjectSyncer)
			gomega.Expect(Sync(context.TODO(), syncer, recorder)).To(gomega.Succeed())

			gomega.Expect(c.Get(context.TODO(), key, deployment)).To(gomega.Succeed())

			gomega.Expect(deployment.ObjectMeta.OwnerReferences).To(gomega.HaveLen(0))

			gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
			gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(gomega.Equal("busybox"))
			gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal("busybox"))

			// since this is an ownerless object, no event is emitted
			gomega.Consistently(recorder.Events).ShouldNot(gomega.Receive())
		})

		ginkgo.It("successfully creates an object and set owner references", func() {
			owner := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
			}
			gomega.Expect(c.Create(context.TODO(), owner)).To(gomega.Succeed())
			syncer = NewDeploymentSyncer(owner).(*ObjectSyncer)
			gomega.Expect(Sync(context.TODO(), syncer, recorder)).To(gomega.Succeed())

			gomega.Expect(c.Get(context.TODO(), key, deployment)).To(gomega.Succeed())

			gomega.Expect(deployment.ObjectMeta.OwnerReferences).To(gomega.HaveLen(1))
			gomega.Expect(deployment.ObjectMeta.OwnerReferences[0].Name).To(gomega.Equal(owner.ObjectMeta.Name))
			gomega.Expect(*deployment.ObjectMeta.OwnerReferences[0].Controller).To(gomega.BeTrue())

			var event string
			gomega.Expect(recorder.Events).To(gomega.Receive(&event))
			gomega.Expect(event).To(gomega.ContainSubstring("ExampleDeploymentSyncSuccessfull"))
			gomega.Expect(event).To(gomega.ContainSubstring("*v1.Deployment default/example created successfully"))
		})
	})
})
