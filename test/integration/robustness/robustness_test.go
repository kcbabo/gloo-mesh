package robustness_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Robustness", func() {
	It("deterministically translates inputs into expected outputs", func() {
		//meshConfig := &istiov1alpha1.MeshConfig{
		//	DefaultConfig: &istiov1alpha1.ProxyConfig{
		//		ProxyMetadata: map[string]string{
		//			"ISTIO_META_DNS_CAPTURE": strconv.FormatBool(true),
		//		},
		//	},
		//	TrustDomain: istio.IstioTrustDomain,
		//}
		//yaml, err := protomarshal.ToYAML(meshConfig)
		//Expect(err).To(BeNil())
		//v := sprintObj(&corev1.ConfigMap{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Namespace:   "istio-namespace",
		//		Name:        "istio",
		//	},
		//	Data: map[string]string{
		//		"mesh": yaml,
		//	},
		//})
		//Expect(v).NotTo(HaveOccurred())

		testCase{states: []testState{
			{
				name:        "basic",
				description: "mesh detected when istiod present",
			},
		}}.execute(rootCtx)
	})
})
