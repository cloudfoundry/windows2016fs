package hydrator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHydrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hydrator Suite")
}
