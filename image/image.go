package image

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/windows2016fs/layer"

	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate counterfeiter . LayerManager
type LayerManager interface {
	Extract(string, string, []string) error
	Delete(string) error
	State(string) (layer.State, error)
}

type Manager struct {
	srcDir       string
	manifest     v1.Manifest
	layerManager LayerManager
	output       io.Writer
}

func NewManager(srcDir string, manifest v1.Manifest, layerManager LayerManager, output io.Writer) *Manager {
	return &Manager{
		srcDir:       srcDir,
		manifest:     manifest,
		layerManager: layerManager,
		output:       output,
	}
}

func (m *Manager) Extract() (string, error) {
	parentLayerIds := []string{}
	for _, l := range m.manifest.Layers {
		if !validMediaType(l.MediaType) {
			return "", fmt.Errorf("invalid layer media type: %s", l.MediaType)
		}

		layerId, err := getLayerSHA(l.Digest)
		if err != nil {
			return "", err
		}
		layerTgz := filepath.Join(m.srcDir, layerId)

		state, err := m.layerManager.State(layerId)
		if err != nil {
			return "", err
		}

		switch state {
		case layer.Incomplete:
			if err := m.layerManager.Delete(layerId); err != nil {
				return "", err
			}
			fallthrough
		case layer.NotExist:
			fmt.Fprintf(m.output, "Extracting %s... ", layerId)
			if err := m.layerManager.Extract(layerTgz, layerId, parentLayerIds); err != nil {
				return "", err
			}
			fmt.Fprintln(m.output, "Done.")
		case layer.Valid:
			fmt.Fprintf(m.output, "Layer %s already exists\n", layerId)
		default:
			panic(fmt.Sprintf("invalid layer state: %d", state))
		}

		parentLayerIds = append([]string{layerId}, parentLayerIds...)
	}

	return parentLayerIds[0], nil
}

func getLayerSHA(d digest.Digest) (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}

	return d.Encoded(), nil
}

func validMediaType(mediaType string) bool {
	return strings.HasSuffix(mediaType, ".tar.gzip") || strings.HasSuffix(mediaType, ".tar+gzip")
}
