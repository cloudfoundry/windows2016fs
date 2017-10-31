package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/Microsoft/hcsshim"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Extract", func() {
	var (
		outputDir   string
		extractSess *gexec.Session
		extractArgs []string

		imageLayers = []string{
			"bb09eb0ec8384b31c735f4d5be5877cd85464c6202b718633ef5ea8299caad86",
			"c5626ce5a7415723bc0fc31bc6b61240b1a1dc0fdbd2757b01ed6141a1ec1a56",
			"ad09b0550b6c41c96a80f476f16b2ad5160d9c10545a05a73b8eece84b5d9d49",
			"407ada6e90de9752a53cb9f52b7947a0e38a9b21a349970ace15c68890d72511",
		}
	)

	BeforeEach(func() {
		var err error
		outputDir, err = ioutil.TempDir("", "extract.integration.test")
		Expect(err).NotTo(HaveOccurred())
		extractArgs = []string{rootfsTgz, outputDir}
	})

	JustBeforeEach(func() {
		var err error
		cmd := exec.Command(extractBin, extractArgs...)
		extractSess, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		for _, layer := range imageLayers {
			Expect(hcsshim.DestroyLayer(hcsshim.DriverInfo{HomeDir: outputDir, Flavour: 1}, layer)).To(Succeed())
		}
		Expect(os.RemoveAll(outputDir)).To(Succeed())
	})

	It("extracts a rootfs tarball to the output directory", func() {
		Eventually(extractSess).Should(gexec.Exit(0))
		for _, layer := range imageLayers {
			Expect(filepath.Join(outputDir, layer)).To(BeADirectory())
			data, err := ioutil.ReadFile(filepath.Join(outputDir, layer, ".complete"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(layer))
		}
		Expect(filepath.Join(outputDir, imageLayers[1], "Files", "ProgramData", "out.txt")).To(BeAnExistingFile())
		Expect(filepath.Join(outputDir, imageLayers[0], "Files", "ProgramData", "out1.txt")).To(BeAnExistingFile())
	})

	It("only outputs the top layer path to stdout", func() {
		Eventually(extractSess).Should(gexec.Exit(0))
		Expect(string(extractSess.Out.Contents())).To(Equal(filepath.Join(outputDir, imageLayers[0])))
	})

	Context("when an image has a layer that has been partially extracted", func() {
		var originalCreateTime int64

		BeforeEach(func() {
			cmd := exec.Command(extractBin, rootfsTgz, outputDir)
			firstSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(firstSess).Should(gexec.Exit(0))
			Expect(os.Remove(filepath.Join(outputDir, imageLayers[0], "Files", "ProgramData", "out1.txt"))).To(Succeed())
			Expect(os.Remove(filepath.Join(outputDir, imageLayers[0], ".complete"))).To(Succeed())
			originalCreateTime = getCreatedTime(filepath.Join(outputDir, imageLayers[0]))
		})

		It("re-extracts the layer", func() {
			Eventually(extractSess).Should(gexec.Exit(0))
			Expect(filepath.Join(outputDir, imageLayers[0], "Files", "ProgramData", "out1.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir, imageLayers[0], ".complete")).To(BeAnExistingFile())
			Expect(getCreatedTime(filepath.Join(outputDir, imageLayers[0]))).To(BeNumerically(">", originalCreateTime))
		})
	})

	Context("when an image has a layer that has been fully extracted", func() {
		var originalCreateTimes []int64

		BeforeEach(func() {
			cmd := exec.Command(extractBin, rootfsTgz, outputDir)
			firstSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(firstSess).Should(gexec.Exit(0))
			for _, layer := range imageLayers {
				originalCreateTimes = append(originalCreateTimes, getCreatedTime(filepath.Join(outputDir, layer)))
			}
		})

		It("does not modify the layer", func() {
			Eventually(extractSess).Should(gexec.Exit(0))
			Expect(filepath.Join(outputDir, imageLayers[0], "Files", "ProgramData", "out1.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir, imageLayers[0], ".complete")).To(BeAnExistingFile())

			for i, layer := range imageLayers {
				Expect(getCreatedTime(filepath.Join(outputDir, layer))).To(Equal(originalCreateTimes[i]))
			}
		})
	})

	Context("when not provided rootfs tgz and output dir arguments", func() {
		BeforeEach(func() {
			extractArgs = []string{"some-arg"}
		})

		It("fails with a helpful error message", func() {
			Eventually(extractSess).Should(gexec.Exit(1))
			Expect(extractSess.Err).To(gbytes.Say("ERROR: Invalid arguments, usage: .*extract.exe <rootfs-tarball> <output-dir>"))
		})
	})
})

func getCreatedTime(file string) int64 {
	fi, err := os.Stat(file)
	Expect(err).NotTo(HaveOccurred())
	return fi.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds()
}
