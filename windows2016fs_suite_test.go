package windows2016fs_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	tag                 string
	imageNameAndTag     string
	testImageNameAndTag string
	tempDirPath         string
	shareUsername       string
	sharePassword       string
	shareName           string
	sharePort           string
	shareIP             string
	shareFqdn           string
	err                 error
)

var _ = BeforeSuite(func() {
	tempDirPath, err = os.MkdirTemp("", "build")
	Expect(err).NotTo(HaveOccurred())

	shareName = lookupEnv("SHARE_NAME")
	shareUsername = lookupEnv("SHARE_USERNAME")
	sharePassword = lookupEnv("SHARE_PASSWORD")
	sharePort = lookupEnv("SHARE_PORT")
	shareFqdn = lookupEnv("SHARE_FQDN")
	shareIP = lookupEnv("SHARE_IP")
	tag = lookupEnv("VERSION_TAG")
	testImageNameAndTag = fmt.Sprintf("windows2016fs-test:%s", tag)

	if os.Getenv("TEST_CANDIDATE_IMAGE") == "" {
		depDir := lookupEnv("DEPENDENCIES_DIR")
		imageNameAndTag = fmt.Sprintf("windows2016fs-candidate:%s", tag)
		buildDockerImage(tempDirPath, depDir, imageNameAndTag, tag)
	} else {
		imageNameAndTag = os.Getenv("TEST_CANDIDATE_IMAGE")
	}
})

func TestWindows2016fs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Windows2016fs Suite")
}
