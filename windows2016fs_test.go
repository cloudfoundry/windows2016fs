package windows2016fs_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func expectCommand(executable string, params ...string) {
	command := exec.Command(executable, params...)
	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 10*time.Minute).Should(Exit(0))
}

func lookupEnv(envName string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		Fail(fmt.Sprintf("Environment variable %s must be set", envName))
	}

	return value
}

func buildDockerImage(tempDirPath, depDir, imageId, tag string) {
	dockerSrcPath := filepath.Join(tag, "Dockerfile")
	Expect(dockerSrcPath).To(BeARegularFile())

	Expect(depDir).To(BeADirectory())

	expectCommand("powershell", "Copy-Item", "-Path", dockerSrcPath, "-Destination", tempDirPath)

	expectCommand("powershell", "Copy-Item", "-Path", filepath.Join(depDir, "*"), "-Destination", tempDirPath)

	expectCommand("docker", "build", "-f", filepath.Join(tempDirPath, "Dockerfile"), "--tag", imageId, tempDirPath)
}

func expectMountSMBImage(shareUnc, shareUsername, sharePassword, tempDirPath, imageId string) {
	command := exec.Command(
		"docker",
		"run",
		"--rm",
		"--interactive",
		"--env", fmt.Sprintf("SHARE_UNC=%s", shareUnc),
		"--env", fmt.Sprintf("SHARE_USERNAME=%s", shareUsername),
		"--env", fmt.Sprintf("SHARE_PASSWORD=%s", sharePassword),
		imageId,
		"powershell",
	)

	stdin, err := command.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	containerTestPs1Content, err := ioutil.ReadFile("container-test.ps1")
	Expect(err).ToNot(HaveOccurred())

	_, err = io.WriteString(stdin, string(containerTestPs1Content))
	Expect(err).ToNot(HaveOccurred())
	stdin.Close()

	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 5*time.Minute).Should(Exit(0))
}

var _ = Describe("Windows2016fs", func() {
	var (
		tag            string
		imageId        string
		tempDirPath    string
		shareUsername  string
		shareUsername2 string
		sharePassword  string
		shareName      string
		shareIP        string
		shareFqdn      string
		err            error
	)

	BeforeSuite(func() {
		tempDirPath, err = ioutil.TempDir("", "build")
		Expect(err).NotTo(HaveOccurred())

		shareName = lookupEnv("SHARE_NAME")
		shareUsername = lookupEnv("SHARE_USERNAME")
		shareUsername2 = lookupEnv("SHARE_USERNAME2")
		sharePassword = lookupEnv("SHARE_PASSWORD")
		shareFqdn = lookupEnv("SHARE_FQDN")
		shareIP = lookupEnv("SHARE_IP")
		tag = lookupEnv("VERSION_TAG")
		imageId = fmt.Sprintf("windows2016fs-ci:%s", tag)
		depDir := lookupEnv("DEPENDENCIES_DIR")

		buildDockerImage(tempDirPath, depDir, imageId, tag)
	})

	It("can write to an IP-based smb share", func() {
		shareUnc := fmt.Sprintf("\\\\%s\\%s", shareIP, shareName)
		expectMountSMBImage(shareUnc, shareUsername, sharePassword, tempDirPath, imageId)
	})

	It("can write to an FQDN-based smb share", func() {
		if tag == "1709" {
			Skip("FQDNs not yet enabled on 1709")
		}

		shareUnc := fmt.Sprintf("\\\\%s\\%s", shareFqdn, shareName)
		expectMountSMBImage(shareUnc, shareUsername, sharePassword, tempDirPath, imageId)
	})

	It("can access one share multiple times, with multiple different credentials on the same VM", func() {
		shareUnc := fmt.Sprintf("\\\\%s\\%s", shareIP, shareName)

		wg := new(sync.WaitGroup)
		wg.Add(2)

		go func() {
			expectMountSMBImage(shareUnc, shareUsername, sharePassword, tempDirPath, imageId)
			wg.Done()
		}()

		go func() {
			expectMountSMBImage(shareUnc, shareUsername2, sharePassword, tempDirPath, imageId)
			wg.Done()
		}()
		wg.Wait()
	})
})
