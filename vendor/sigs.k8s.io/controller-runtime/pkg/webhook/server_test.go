/*
Copyright 2018 The Kubernetes Authors.

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

package webhook

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/internal/cert"
	"sigs.k8s.io/controller-runtime/pkg/webhook/internal/cert/generator"
	"sigs.k8s.io/controller-runtime/pkg/webhook/internal/cert/writer"
	"sigs.k8s.io/testing_frameworks/integration/addr"
)

type fakeCertWriter struct {
	changed bool
}

func (cw *fakeCertWriter) EnsureCert(dnsName string) (*generator.Artifacts, bool, error) {
	return &generator.Artifacts{}, cw.changed, nil
}

func (cw *fakeCertWriter) Inject(objs ...runtime.Object) error {
	return nil
}

var _ = Describe("webhook server", func() {
	Describe("run", func() {
		var stop chan struct{}
		var s *Server
		var cn = "example.com"

		BeforeEach(func() {
			port, _, err := addr.Suggest()
			Expect(err).NotTo(HaveOccurred())
			s = &Server{
				sMux: http.NewServeMux(),
				ServerOptions: ServerOptions{
					Port: int32(port),
					BootstrapOptions: &BootstrapOptions{
						Host: &cn,
					},
				},
			}

			cg := &generator.SelfSignedCertGenerator{}
			s.CertDir, err = ioutil.TempDir("/tmp", "controller-runtime-")
			Expect(err).NotTo(HaveOccurred())
			certWriter, err := writer.NewFSCertWriter(writer.FSCertWriterOptions{CertGenerator: cg, Path: s.CertDir})
			Expect(err).NotTo(HaveOccurred())
			_, _, err = certWriter.EnsureCert(cn)
			Expect(err).NotTo(HaveOccurred())

			stop = make(chan struct{})
		})

		It("should stop if the stop channel is closed", func() {
			var e error
			go func() {
				defer GinkgoRecover()
				e = s.run(stop)
			}()

			Eventually(func() *http.Server {
				return s.httpServer
			}).ShouldNot(BeNil())

			close(stop)
			Expect(e).NotTo(HaveOccurred())
		})

		It("should exit if the server encounter an unexpected error", func() {
			var e error
			go func() {
				defer GinkgoRecover()
				e = s.run(stop)
			}()

			Eventually(func() *http.Server {
				return s.httpServer
			}).ShouldNot(BeNil())

			err := s.httpServer.Shutdown(context.Background())
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				return e
			}).Should(Equal(http.ErrServerClosed))

			close(stop)
		})

		It("should be able to keep existing valid cert when timer fires", func() {
			var e error
			defaultCertRefreshInterval = 500 * time.Millisecond

			s.certProvisioner = &cert.Provisioner{
				CertWriter: &fakeCertWriter{changed: false},
			}

			go func() {
				defer GinkgoRecover()
				e = s.run(stop)
			}()

			// Wait for multiple cycles of timer firing
			time.Sleep(2 * time.Second)
			Expect(e).NotTo(HaveOccurred())

			close(stop)
		})

		It("should be able to rotate the cert when timer fires", func() {
			var e error
			defaultCertRefreshInterval = 500 * time.Millisecond
			s.certProvisioner = &cert.Provisioner{
				CertWriter: &fakeCertWriter{changed: true},
			}

			go func() {
				defer GinkgoRecover()
				e = s.run(stop)
			}()

			Eventually(func() *http.Server {
				return s.httpServer
			}).ShouldNot(BeNil())

			// Wait for multiple cycles of timer firing
			time.Sleep(2 * time.Second)
			Expect(e).NotTo(HaveOccurred())

			close(stop)
		})

		AfterEach(func() {
			defaultCertRefreshInterval = 3 * 30 * 24 * time.Hour
			err := os.RemoveAll(s.CertDir)
			Expect(err).NotTo(HaveOccurred())
		}, 60)
	})
})
