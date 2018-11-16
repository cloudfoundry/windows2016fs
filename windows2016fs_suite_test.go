package windows2016fs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWindows2016fs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Windows2016fs Suite")
}
