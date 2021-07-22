package peerAuth_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPeerAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Auth Suite")
}
