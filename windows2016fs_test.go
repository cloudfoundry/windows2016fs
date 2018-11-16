package windows2016fs_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func expectCommand(executable string, params ...string) {
	command := exec.Command(executable, params...)
	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 10*time.Second).Should(Exit(0))
}

func buildDockerImage(tempDirPath, imageId, tag string) {
	dockerSrcPath := filepath.Join(tag, "Dockerfile")
	Expect(dockerSrcPath).To(BeARegularFile())

	depDir := os.Getenv("DEPENDENCIES_DIR")
	Expect(depDir).To(BeADirectory())

	expectCommand("powershell", "Copy-Item", "-Path", dockerSrcPath, "-Destination", tempDirPath)

	expectCommand("powershell", "Copy-Item", "-Path", filepath.Join(depDir, "*"), "-Destination", tempDirPath)

	expectCommand("powershell", "Copy-Item", "-Path", "container-test.ps1", "-Destination", tempDirPath)

	expectCommand("docker", "build", "-f", filepath.Join(tempDirPath, "Dockerfile"), "--tag", imageId, tempDirPath)
}

func setupSMBShare(tempDirPath string) {
	command := exec.Command("powershell", fmt.Sprintf("New-SmbShare -Name my-share -Path %s -ErrorAction Stop", tempDirPath))
	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	Eventually(session, 10*time.Second).Should(Exit(0))
}

func expectMountSMBImage(tempDirPath, imageId string) {
	volumeMapping := fmt.Sprintf("%s:c:/ci", tempDirPath)
	testFilePath := "c:/ci/container-test.ps1"
	command := exec.Command(
		"docker",
		"run",
		"--rm",
		"--volume", volumeMapping,
		"--env", fmt.Sprintf("SHARE_HOST=%s", os.Getenv("SHARE_HOST")),
		"--env", fmt.Sprintf("SHARE_USERNAME=%s", os.Getenv("SHARE_USERNAME")),
		"--env", fmt.Sprintf("SHARE_PASSWORD=%s", os.Getenv("SHARE_PASSWORD")),
		imageId,
		"powershell",
		testFilePath,
	)
	session, err := Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 5*time.Minute).Should(Exit(0))
}

var _ = Describe("Windows2016fs", func() {
	AfterEach(func() {
		exec.Command("powershell", "Remove-SmbShare -Name my-share -Force -ErrorAction SilentlyContinue").Run()
	})

	It("can write to an smb share", func() {
		tag := os.Getenv("VERSION_TAG")
		imageId := fmt.Sprintf("windows2016fs-ci:%s", tag)
		tempDirPath, err := ioutil.TempDir("", "build")
		Expect(err).ToNot(HaveOccurred())

		buildDockerImage(tempDirPath, imageId, tag)
		setupSMBShare(tempDirPath)
		expectMountSMBImage(tempDirPath, imageId)
	})
})
