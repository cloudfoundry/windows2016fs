package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

type Metadata struct {
	outDir  string
	layers  []v1.Descriptor
	diffIds []digest.Digest
}

func New(outDir string, layers []v1.Descriptor, diffIds []digest.Digest) *Metadata {
	return &Metadata{
		outDir:  outDir,
		layers:  layers,
		diffIds: diffIds,
	}
}

func (m *Metadata) Write() error {
	if err := m.writeOCILayout(); err != nil {
		return err
	}

	configDescriptor, err := m.writeConfig()
	if err != nil {
		return err
	}

	manifestDescriptor, err := m.writeManifest(configDescriptor)
	if err != nil {
		return err
	}

	return m.writeIndexJson(manifestDescriptor)
}

func (m *Metadata) writeOCILayout() error {
	il := v1.ImageLayout{
		Version: specs.Version,
	}
	data, err := json.Marshal(il)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(m.outDir, "oci-layout"), data, 0644)
}

func (m *Metadata) writeConfig() (v1.Descriptor, error) {
	ic := v1.Image{
		Architecture: "amd64",
		OS:           "windows",
		RootFS:       v1.RootFS{Type: "layers", DiffIDs: m.diffIds},
	}

	d, err := m.writeBlob(ic)
	if err != nil {
		return v1.Descriptor{}, err
	}

	d.MediaType = v1.MediaTypeImageConfig
	return d, nil
}

func (m *Metadata) writeManifest(config v1.Descriptor) (v1.Descriptor, error) {
	im := v1.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config:    config,
		Layers:    m.layers,
	}

	d, err := m.writeBlob(im)
	if err != nil {
		return v1.Descriptor{}, err
	}

	d.MediaType = v1.MediaTypeImageManifest
	d.Platform = &v1.Platform{OS: "windows", Architecture: "amd64"}
	return d, nil
}

func (m *Metadata) writeBlob(blob interface{}) (v1.Descriptor, error) {
	data, err := json.Marshal(blob)
	if err != nil {
		return v1.Descriptor{}, err
	}

	blobSha := fmt.Sprintf("%x", sha256.Sum256(data))

	blobsDir := filepath.Join(m.outDir, "blobs", "sha256")
	if err := os.MkdirAll(blobsDir, 0755); err != nil {
		return v1.Descriptor{}, err
	}

	if err := ioutil.WriteFile(filepath.Join(blobsDir, blobSha), data, 0644); err != nil {
		return v1.Descriptor{}, err
	}

	return v1.Descriptor{
		Size:   int64(len(data)),
		Digest: digest.NewDigestFromEncoded(digest.SHA256, blobSha),
	}, nil
}

func (m *Metadata) writeIndexJson(manifest v1.Descriptor) error {
	ii := v1.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Manifests: []v1.Descriptor{manifest},
	}

	data, err := json.Marshal(ii)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(m.outDir, "index.json"), data, 0644)
}
