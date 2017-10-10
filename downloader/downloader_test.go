package downloader_test

import (
	"errors"
	"io/ioutil"

	"code.cloudfoundry.org/windows2016fs/downloader"
	"code.cloudfoundry.org/windows2016fs/downloader/downloaderfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

var _ = Describe("Downloader", func() {
	const (
		downloadDir = "some-directory"
		outputTgz   = "a-path/some-file.tgz"
	)

	var (
		layers     []v1.Descriptor
		manifest   v1.Manifest
		registry   *downloaderfakes.FakeRegistry
		compressor *downloaderfakes.FakeCompressor
		d          *downloader.Downloader
	)

	BeforeEach(func() {
		layers = []v1.Descriptor{
			{Digest: "layer1"},
			{Digest: "layer2"},
		}
		manifest = v1.Manifest{Layers: layers}
		registry = &downloaderfakes.FakeRegistry{}
		compressor = &downloaderfakes.FakeCompressor{}

		registry.DownloadManifestReturnsOnCall(0, manifest, nil)

		d = downloader.New(downloadDir, outputTgz, registry, compressor, ioutil.Discard)
	})

	Describe("Run", func() {
		It("Downloads the manifest, all the layers, and tars them up", func() {
			Expect(d.Run()).To(Succeed())

			Expect(registry.DownloadManifestCallCount()).To(Equal(1))
			Expect(registry.DownloadManifestArgsForCall(0)).To(Equal("some-directory"))

			Expect(registry.DownloadLayerCallCount()).To(Equal(2))
			l1, dir := registry.DownloadLayerArgsForCall(0)
			Expect(dir).To(Equal("some-directory"))
			l2, dir := registry.DownloadLayerArgsForCall(1)
			Expect(dir).To(Equal("some-directory"))

			Expect(layers).To(ConsistOf(l1, l2))

			Expect(compressor.WriteTgzCallCount()).To(Equal(1))
			dir, file := compressor.WriteTgzArgsForCall(0)
			Expect(dir).To(Equal("some-directory"))
			Expect(file).To(Equal("a-path/some-file.tgz"))
		})
	})

	Context("downloading the manifest fails", func() {
		BeforeEach(func() {
			registry.DownloadManifestReturnsOnCall(0, v1.Manifest{}, errors.New("couldn't download manifest"))
		})

		It("returns an error", func() {
			Expect(d.Run().Error()).To(Equal("couldn't download manifest"))
			Expect(registry.DownloadLayerCallCount()).To(Equal(0))
			Expect(compressor.WriteTgzCallCount()).To(Equal(0))
		})
	})

	Context("downloading a layer fails", func() {
		BeforeEach(func() {
			registry.DownloadLayerReturnsOnCall(1, errors.New("couldn't download layer2"))
		})

		It("returns an error", func() {
			Expect(d.Run().Error()).To(Equal("couldn't download layer2"))
			Expect(compressor.WriteTgzCallCount()).To(Equal(0))
		})
	})

	Context("compressing the downloaded layers fails", func() {
		BeforeEach(func() {
			compressor.WriteTgzReturnsOnCall(0, errors.New("couldn't create tar"))
		})

		It("returns an error", func() {
			Expect(d.Run().Error()).To(Equal("couldn't create tar"))
		})
	})
})
