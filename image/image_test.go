package image_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/windows2016fs/image"
	"code.cloudfoundry.org/windows2016fs/image/imagefakes"
	"code.cloudfoundry.org/windows2016fs/layer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

var _ = Describe("Image", func() {
	Describe("Extract", func() {
		var (
			srcDir   string
			destDir  string
			tempDir  string
			manifest v1.Manifest
			lm       *imagefakes.FakeLayerManager
			im       *image.Manager
		)

		const (
			layer1 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			layer2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
			layer3 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "windows2016fs.image")
			Expect(err).NotTo(HaveOccurred())

			layers := []v1.Descriptor{
				{Digest: digest.NewDigestFromEncoded("sha256", layer1), MediaType: "some.type.tar.gzip"},
				{Digest: digest.NewDigestFromEncoded("sha256", layer2), MediaType: "some.other.type.tar+gzip"},
				{Digest: digest.NewDigestFromEncoded("sha256", layer3), MediaType: "some.type.tar.gzip"},
			}

			manifest = v1.Manifest{
				Layers: layers,
			}

			srcDir = filepath.Join(tempDir, "src")
			destDir = filepath.Join(tempDir, "dest")
			lm = &imagefakes.FakeLayerManager{}
		})

		JustBeforeEach(func() {
			im = image.NewManager(srcDir, manifest, lm, ioutil.Discard)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		It("extracts all the layers, returning the top layer id", func() {
			topLayerId, err := im.Extract()
			Expect(err).NotTo(HaveOccurred())
			Expect(topLayerId).To(Equal(layer3))

			Expect(lm.DeleteCallCount()).To(Equal(0))

			Expect(lm.ExtractCallCount()).To(Equal(3))
			tgz, id, parentIds := lm.ExtractArgsForCall(0)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
			Expect(id).To(Equal(layer1))
			Expect(parentIds).To(Equal([]string{}))

			tgz, id, parentIds = lm.ExtractArgsForCall(1)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer2)))
			Expect(id).To(Equal(layer2))
			Expect(parentIds).To(Equal([]string{layer1}))

			tgz, id, parentIds = lm.ExtractArgsForCall(2)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
			Expect(id).To(Equal(layer3))
			Expect(parentIds).To(Equal([]string{layer2, layer1}))
		})

		Context("a layer has already been extracted", func() {
			BeforeEach(func() {
				lm.StateReturnsOnCall(1, layer.Valid, nil)
			})

			It("does not re-extract the existing layer", func() {
				_, err := im.Extract()
				Expect(err).NotTo(HaveOccurred())
				Expect(lm.DeleteCallCount()).To(Equal(0))

				Expect(lm.ExtractCallCount()).To(Equal(2))
				tgz, id, parentIds := lm.ExtractArgsForCall(0)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
				Expect(id).To(Equal(layer1))
				Expect(parentIds).To(Equal([]string{}))

				tgz, id, parentIds = lm.ExtractArgsForCall(1)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
				Expect(id).To(Equal(layer3))
				Expect(parentIds).To(Equal([]string{layer2, layer1}))
			})
		})

		Context("there is an invalid layer", func() {
			BeforeEach(func() {
				lm.StateReturnsOnCall(1, layer.Incomplete, nil)
			})

			It("deletes the incomplete layer and re-extracts", func() {
				_, err := im.Extract()
				Expect(err).NotTo(HaveOccurred())

				Expect(lm.DeleteCallCount()).To(Equal(1))
				Expect(lm.DeleteArgsForCall(0)).To(Equal(layer2))

				Expect(lm.ExtractCallCount()).To(Equal(3))
				tgz, id, parentIds := lm.ExtractArgsForCall(0)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
				Expect(id).To(Equal(layer1))
				Expect(parentIds).To(Equal([]string{}))

				tgz, id, parentIds = lm.ExtractArgsForCall(1)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer2)))
				Expect(id).To(Equal(layer2))
				Expect(parentIds).To(Equal([]string{layer1}))

				tgz, id, parentIds = lm.ExtractArgsForCall(2)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
				Expect(id).To(Equal(layer3))
				Expect(parentIds).To(Equal([]string{layer2, layer1}))
			})
		})

		Context("provided an invalid content digest", func() {
			BeforeEach(func() {
				manifest = v1.Manifest{
					Layers: []v1.Descriptor{
						{Digest: digest.Digest("hello"), MediaType: "something.tar.gzip"},
					},
				}
			})

			It("returns an error", func() {
				_, err := im.Extract()
				Expect(err).To(MatchError(digest.ErrDigestInvalidFormat))
			})
		})

		Context("the media type is not a .tar.gzip or .tar+gzip", func() {
			BeforeEach(func() {
				manifest = v1.Manifest{
					Layers: []v1.Descriptor{
						{Digest: digest.NewDigestFromEncoded("sha256", layer1), MediaType: "some-invalid-string"},
					},
				}
			})

			It("returns an error", func() {
				_, err := im.Extract()
				Expect(err).To(MatchError(errors.New("invalid layer media type: some-invalid-string")))
			})
		})
	})
})
