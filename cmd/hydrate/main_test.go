package main_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/archiver/extractor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

const v1sha = "7406625cf20ba9cc3caeb6a56b1709665e8d094fef5047499603af51915b3f54"
const v2sha = "c7de8bbeff027c47044143e7801269863160155afa90ddffc47dd0fc6920e2b7"

var _ = Describe("Hydrate", func() {
	var (
		outputDir        string
		hydrateArgs      []string
		hydrateSess      *gexec.Session
		imageName        string
		imageTag         string
		imageTarballName string
	)

	BeforeEach(func() {
		var err error
		outputDir, err = ioutil.TempDir("", "hydrateOutput")
		Expect(err).NotTo(HaveOccurred())

		imageName = "pivotalgreenhouse/windows2016fs-hydrate"
		imageTag = "1.0.0"
		nameParts := strings.Split(imageName, "/")
		Expect(len(nameParts)).To(Equal(2))
		imageTarballName = fmt.Sprintf("%s-%s.tgz", nameParts[1], imageTag)

		hydrateArgs = []string{}
	})

	JustBeforeEach(func() {
		var err error

		command := exec.Command(hydrateBin, hydrateArgs...)
		hydrateSess, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(outputDir)).To(Succeed())
	})

	Context("when provided an output directory", func() {
		Context("when provided an image tag", func() {
			BeforeEach(func() {
				hydrateArgs = []string{"--outputDir", outputDir, "--image", imageName, "--tag", imageTag}
			})

			It("downloads the correct version of the image", func() {
				Eventually(hydrateSess).Should(gexec.Exit(0))

				tarball := filepath.Join(outputDir, imageTarballName)
				ItHasTheCorrectSHA256(tarball, v1sha)

				imageContentsDir := extractTarball(tarball)
				defer os.RemoveAll(imageContentsDir)

				manifestFile := filepath.Join(imageContentsDir, "manifest.json")
				Expect(manifestFile).To(BeAnExistingFile())

				var im v1.Manifest
				content, err := ioutil.ReadFile(manifestFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(json.Unmarshal(content, &im)).To(Succeed())

				for _, layer := range im.Layers {
					layerSHA := strings.TrimPrefix(string(layer.Digest), "sha256:")
					blob := filepath.Join(imageContentsDir, layerSHA)
					ItHasTheCorrectSHA256(blob, layerSHA)
				}
			})

			Context("when not provided an image tag", func() {
				BeforeEach(func() {
					imageTag = "latest"
					nameParts := strings.Split(imageName, "/")
					Expect(len(nameParts)).To(Equal(2))
					imageTarballName = fmt.Sprintf("%s-%s.tgz", nameParts[1], imageTag)
					hydrateArgs = []string{"--outputDir", outputDir, "--image", imageName}
				})

				It("downloads the latest image version", func() {
					Eventually(hydrateSess).Should(gexec.Exit(0))

					tarball := filepath.Join(outputDir, imageTarballName)
					ItHasTheCorrectSHA256(tarball, v2sha)

					imageContentsDir := extractTarball(tarball)
					defer os.RemoveAll(imageContentsDir)

					manifestFile := filepath.Join(imageContentsDir, "manifest.json")
					Expect(manifestFile).To(BeAnExistingFile())

					var im v1.Manifest
					content, err := ioutil.ReadFile(manifestFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(json.Unmarshal(content, &im)).To(Succeed())

					for _, layer := range im.Layers {
						layerSHA := strings.TrimPrefix(string(layer.Digest), "sha256:")
						blob := filepath.Join(imageContentsDir, layerSHA)
						ItHasTheCorrectSHA256(blob, layerSHA)
					}
				})
			})
		})

		Context("when not provided an image", func() {
			BeforeEach(func() {
				hydrateArgs = []string{"--outputDir", outputDir}
			})

			It("errors", func() {
				Eventually(hydrateSess).Should(gexec.Exit())
				Expect(hydrateSess.ExitCode()).ToNot(Equal(0))
				Expect(string(hydrateSess.Err.Contents())).To(ContainSubstring("ERROR: No image name provided"))
			})
		})
	})

	Context("when the output directory does not exist", func() {
		BeforeEach(func() {
			hydrateArgs = []string{"--image", imageName, "--tag", imageTag, "--outputDir", filepath.Join(outputDir, "random-dir")}
		})

		It("creates it and outputs the image tarball to that directory", func() {
			Eventually(hydrateSess).Should(gexec.Exit(0))
			ItHasTheCorrectSHA256(filepath.Join(outputDir, "random-dir", imageTarballName), v1sha)
		})
	})

	Context("when no output directory is provided", func() {
		BeforeEach(func() {
			hydrateArgs = []string{"--image", imageName, "--tag", imageTag}
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), imageTarballName))).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), imageTarballName))).To(Succeed())
		})

		It("outputs to the system temp directory", func() {
			Eventually(hydrateSess).Should(gexec.Exit(0))
			ItHasTheCorrectSHA256(filepath.Join(os.TempDir(), imageTarballName), v1sha)
		})
	})
})

func extractTarball(path string) string {
	tmpDir, err := ioutil.TempDir("", "hydrated")
	Expect(err).NotTo(HaveOccurred())
	err = extractor.NewTgz().Extract(path, tmpDir)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return tmpDir
}

func ItHasTheCorrectSHA256(file, expected string) {
	By("having the correct SHA256", func() {
		Expect(file).To(BeAnExistingFile())
		f, err := os.Open(file)
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()

		h := sha256.New()
		_, err = io.Copy(h, f)
		Expect(err).NotTo(HaveOccurred())
		actualSHA := fmt.Sprintf("%x", h.Sum(nil))
		Expect(actualSHA).To(Equal(expected))
	})
}
