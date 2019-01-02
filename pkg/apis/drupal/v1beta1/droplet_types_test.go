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

package v1beta1_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"golang.org/x/net/context"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/sylus/drupal-operator/pkg/apis/drupal/v1beta1"
)

var _ = ginkgo.Describe("Drupal CRUD", func() {
	var created *v1beta1.Droplet
	var key types.NamespacedName

	ginkgo.BeforeEach(func() {
		key = types.NamespacedName{Name: "foo", Namespace: "default"}

		created = &v1beta1.Droplet{}
		created.Name = key.Name
		created.Namespace = key.Namespace
		created.Spec.Domains = []v1beta1.Domain{"example.com"}
	})

	ginkgo.AfterEach(func() {
		// nolint: errcheck
		c.Delete(context.TODO(), created)
	})

	ginkgo.Describe("when sending a storage request", func() {
		ginkgo.Context("for a valid config", func() {
			ginkgo.It("should provide CRUD access to the object", func() {
				fetched := &v1beta1.Droplet{}

				ginkgo.By("returning success from the create request")
				gomega.Expect(c.Create(context.TODO(), created)).Should(gomega.Succeed())

				ginkgo.By("returning the same object as created")
				gomega.Expect(c.Get(context.TODO(), key, fetched)).Should(gomega.Succeed())
				gomega.Expect(fetched).To(gomega.Equal(created))

				ginkgo.By("allowing label updates")
				updated := fetched.DeepCopy()
				updated.Labels = map[string]string{"hello": "world"}
				gomega.Expect(c.Update(context.TODO(), updated)).Should(gomega.Succeed())
				gomega.Expect(c.Get(context.TODO(), key, fetched)).Should(gomega.Succeed())
				gomega.Expect(fetched).To(gomega.Equal(updated))

				ginkgo.By("deleting an fetched object")
				gomega.Expect(c.Delete(context.TODO(), fetched)).Should(gomega.Succeed())
				gomega.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
			})
		})
	})
})
