package hydrator

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

type Hydrator struct {
	downloadDir string
	outputTgz   string
	registry    Registry
	compressor  Compressor
	logger      io.Writer
}

func New(downloadDir string, outputTgz string, registry Registry, compressor Compressor, logger io.Writer) *Hydrator {
	h := &Hydrator{
		downloadDir: downloadDir,
		outputTgz:   outputTgz,
		registry:    registry,
		compressor:  compressor,
		logger:      logger,
	}
	return h
}

func (h *Hydrator) Run() error {
	manifest, err := h.registry.DownloadManifest(h.downloadDir)
	if err != nil {
		return err
	}

	totalLayers := len(manifest.Layers)
	fmt.Fprintf(h.logger, "Downloading %d layers...\n", totalLayers)
	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)

	for _, layer := range manifest.Layers {
		l := layer
		wg.Add(1)
		go func() {
			fmt.Fprintf(h.logger, "Layer %.15s begin\n", l.Digest)
			defer wg.Done()
			if err := h.registry.DownloadLayer(l, h.downloadDir); err != nil {
				errChan <- err
				return
			}
			fmt.Fprintf(h.logger, "Layer %.15s end\n", l.Digest)
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

	fmt.Fprintf(h.logger, "\nAll layers downloaded, writing %s...\n", h.outputTgz)
	return h.compressor.WriteTgz(h.downloadDir, h.outputTgz)
}
