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
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/sylus/drupal-operator/pkg/apis/drupal/v1beta1"
)

var t *envtest.Environment
var cfg *rest.Config
var c client.Client

func TestV1beta1(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "API v1 Suite", []ginkgo.Reporter{envtest.NewlineReporter{}})
}

var _ = ginkgo.BeforeSuite(func() {
	var err error

	t = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "config", "crds")},
	}

	err = v1beta1.SchemeBuilder.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	cfg, err = t.Start()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})

var _ = ginkgo.AfterSuite(func() {
	gomega.Expect(t.Stop()).To(gomega.Succeed())
})
