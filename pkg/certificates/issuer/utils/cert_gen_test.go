package utils_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/utils"
	. "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	"istio.io/istio/security/pkg/pki/util"
)

var _ = Describe("CertGen workflow", func() {
	assertCsrWorks := func(signingRoot, signingKey []byte) {
		privateKey, err := utils.GeneratePrivateKey(4096)
		Expect(err).NotTo(HaveOccurred())

		hosts := []string{"spiffe://custom-domain/ns/istio-system/sa/istio-pilot-service-account"}
		csr, err := utils.GenerateCertificateSigningRequest(
			hosts,
			"gloo-mesh",
			"mesh-name",
			privateKey,
		)
		Expect(err).NotTo(HaveOccurred())

		inetermediaryCert, err := GenCertForCSR(
			context.Background(),
			hosts,
			csr,
			signingRoot,
			signingKey,
			0,
		)
		Expect(err).NotTo(HaveOccurred())
		pemByt, _ := pem.Decode(inetermediaryCert)
		cert, err := x509.ParseCertificate(pemByt.Bytes)
		Expect(err).NotTo(HaveOccurred())
		Expect(cert.IsCA).To(BeTrue())
		Expect(cert.Subject.OrganizationalUnit).To(ConsistOf("mesh-name"))
		Expect(cert.Subject.Organization).To(ConsistOf("gloo-mesh"))
	}

	It("generates a certificate using generated self signed cert, private key, and certificate signing request", func() {

		options := util.CertOptions{
			Org:          "org",
			IsCA:         true,
			IsSelfSigned: true,
			TTL:          time.Hour * 24 * 365,
			RSAKeySize:   4096,
			PKCS8Key:     false, // currently only supporting PKCS1
		}
		signingRoot, signingKey, err := util.GenCertKeyFromOptions(options)
		Expect(err).NotTo(HaveOccurred())

		assertCsrWorks(signingRoot, signingKey)
	})
})
