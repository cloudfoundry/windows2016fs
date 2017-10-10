package downloader

import (
	"fmt"
	"io"
	"sync"

	"github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate counterfeiter . Registry
type Registry interface {
	DownloadManifest(string) (v1.Manifest, error)
	DownloadLayer(v1.Descriptor, string) error
}

//go:generate counterfeiter . Compressor
type Compressor interface {
	WriteTgz(string, string) error
}

type Downloader struct {
	downloadDir string
	outputTgz   string
	registry    Registry
	compressor  Compressor
	logger      io.Writer
}

func New(downloadDir string, outputTgz string, registry Registry, compressor Compressor, logger io.Writer) *Downloader {
	d := &Downloader{
		downloadDir: downloadDir,
		outputTgz:   outputTgz,
		registry:    registry,
		compressor:  compressor,
		logger:      logger,
	}
	return d
}

func (d *Downloader) Run() error {
	manifest, err := d.registry.DownloadManifest(d.downloadDir)
	if err != nil {
		return err
	}

	totalLayers := len(manifest.Layers)
	fmt.Fprintf(d.logger, "Downloading %d layers...\n", totalLayers)
	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)

	for _, layer := range manifest.Layers {
		l := layer
		wg.Add(1)
		go func() {
			fmt.Fprintf(d.logger, "Layer %.15s begin\n", l.Digest)
			defer wg.Done()
			if err := d.registry.DownloadLayer(l, d.downloadDir); err != nil {
				errChan <- err
				return
			}
			fmt.Fprintf(d.logger, "Layer %.15s end\n", l.Digest)
		}()
	}

	wgEmpty := make(chan interface{}, 1)
	go func() {
		wg.Wait()
		wgEmpty <- nil
	}()

	select {
	case <-wgEmpty:
	case downloadErr := <-errChan:
		return downloadErr
	}

	fmt.Fprintf(d.logger, "\nAll layers downloaded, writing %s...\n", d.outputTgz)
	return d.compressor.WriteTgz(d.downloadDir, d.outputTgz)
}
