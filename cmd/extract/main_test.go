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
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Extract", func() {
	var (
		outputDir   string
		extractSess *gexec.Session

		imageLayers = []string{
			"d9b2e5531a82b33cdd4312401a60c4ff7462531fa562131ee924d6d34ae8bdd7",
			"8ce9b6bd8d238aaedbc2e1765d8e537a5dcf5f2cfbb79535f38a26ef5e3846e8",
			"5cd49617cf500abea7b9f47d82b70455d816ae6b497cabc1fc86a9522d19a828",
			"bce2fbc256ea437a87dadac2f69aabd25bed4f56255549090056c1131fad0277",
		}
	)

	BeforeEach(func() {
		var err error
		outputDir, err = ioutil.TempDir("", "extract.integration.test")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		var err error
		cmd := exec.Command(extractBin, rootfsTgz, outputDir)
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
		Expect(filepath.Join(outputDir, imageLayers[0], "Files", "out.txt")).To(BeAnExistingFile())
		Expect(filepath.Join(outputDir, imageLayers[0], "Files", "out1.txt")).To(BeAnExistingFile())
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
			Expect(os.Remove(filepath.Join(outputDir, imageLayers[0], "Files", "out1.txt"))).To(Succeed())
			Expect(os.Remove(filepath.Join(outputDir, imageLayers[0], ".complete"))).To(Succeed())
			originalCreateTime = getCreatedTime(filepath.Join(outputDir, imageLayers[0]))
		})

		It("re-extracts the layer", func() {
			Eventually(extractSess).Should(gexec.Exit(0))
			Expect(filepath.Join(outputDir, imageLayers[0], "Files", "out1.txt")).To(BeAnExistingFile())
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
			Expect(filepath.Join(outputDir, imageLayers[0], "Files", "out1.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir, imageLayers[0], ".complete")).To(BeAnExistingFile())

			for i, layer := range imageLayers {
				Expect(getCreatedTime(filepath.Join(outputDir, layer))).To(Equal(originalCreateTimes[i]))
			}
		})
	})
})

func getCreatedTime(file string) int64 {
	fi, err := os.Stat(file)
	Expect(err).NotTo(HaveOccurred())
	return fi.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds()
}
