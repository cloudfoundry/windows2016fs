package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/windows2016fs/image"
	"code.cloudfoundry.org/windows2016fs/layer"
	"code.cloudfoundry.org/windows2016fs/writer"

	"code.cloudfoundry.org/archiver/extractor"

	"github.com/Microsoft/hcsshim"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {
	if err := mainBody(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		os.Exit(1)
	}
}

func mainBody() error {
	rootfstgz := os.Args[1]
	outputDir := os.Args[2]

	layerTempDir, err := ioutil.TempDir("", "hcslayers")
	if err != nil {
		return err
	}

	if err := extractor.NewTgz().Extract(rootfstgz, layerTempDir); err != nil {
		return err
	}
	defer os.RemoveAll(layerTempDir)

	manifestData, err := ioutil.ReadFile(filepath.Join(layerTempDir, "manifest.json"))
	if err != nil {
		return err
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	lm := layer.NewManager(hcsshim.DriverInfo{HomeDir: outputDir, Flavour: 1}, &writer.Writer{})
	im := image.NewManager(layerTempDir, manifest, lm, os.Stderr)

	topLayerId, err := im.Extract()
	if err != nil {
		return err
	}

	fmt.Printf(filepath.Join(outputDir, topLayerId))
	return nil
}
