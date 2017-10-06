package compress_test

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/windows2016fs/compress"

	"code.cloudfoundry.org/archiver/extractor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compress", func() {
	Describe("WriteTgz", func() {
		var (
			c          *compress.Compressor
			srcDir     string
			outputDir  string
			outputFile string
		)

		const outputTarSha = "5ba93d2385fc2e96b9c30757ec3a9040bc33d3ebe939e614724c18bf646400db"

		BeforeEach(func() {
			var err error
			srcDir, err = ioutil.TempDir("", "write-tgz.src")
			Expect(err).NotTo(HaveOccurred())

			outputDir, err = ioutil.TempDir("", "write-tgz.out")
			Expect(err).NotTo(HaveOccurred())

			outputFile = filepath.Join(outputDir, "image.tgz")

			Expect(ioutil.WriteFile(filepath.Join(srcDir, "file1"), []byte("contents1"), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(srcDir, "file2"), []byte("contents2"), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(srcDir, "file3"), []byte("contents3"), 0644)).To(Succeed())

			c = compress.New()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(srcDir)).To(Succeed())
			Expect(os.RemoveAll(outputDir)).To(Succeed())
		})

		It("creates a .tgz file with all of the layers + the manifest.json", func() {
			Expect(c.WriteTgz(srcDir, outputFile)).To(Succeed())

			contents := extractTarball(outputFile)
			defer os.RemoveAll(contents)

			data, err := ioutil.ReadFile(filepath.Join(contents, "file1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("contents1"))

			data, err = ioutil.ReadFile(filepath.Join(contents, "file2"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("contents2"))

			data, err = ioutil.ReadFile(filepath.Join(contents, "file3"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("contents3"))

			ItHasTheCorrectSHA256(outputFile, outputTarSha)
		})

		Context("the source dir contains a sub directory", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(srcDir, "sub-dir"), 0755)).To(Succeed())
			})

			It("returns an error", func() {
				err := c.WriteTgz(srcDir, outputFile)
				Expect(err).To(BeAssignableToTypeOf(&compress.ErrInvalidSource{}))
			})
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
