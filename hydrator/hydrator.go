package hydrator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/windows2016fs/compress"
	"code.cloudfoundry.org/windows2016fs/downloader"
	"code.cloudfoundry.org/windows2016fs/registry"
)

type Hydrator struct {
	outDir    string
	imageName string
	imageTag  string
}

func New(outDir, imageName, imageTag string) *Hydrator {
	return &Hydrator{
		outDir:    outDir,
		imageName: imageName,
		imageTag:  imageTag,
	}
}

func (h *Hydrator) Run() error {
	if err := os.MkdirAll(h.outDir, 0755); err != nil {
		return errors.New("Could not create output directory")
	}

	if h.imageName == "" {
		return errors.New("No image name provided")
	}

	nameParts := strings.Split(h.imageName, "/")
	if len(nameParts) != 2 {
		return errors.New("Invalid image name")
	}
	outFile := filepath.Join(h.outDir, fmt.Sprintf("%s-%s.tgz", nameParts[1], h.imageTag))

	tempDir, err := ioutil.TempDir("", "hydrate")
	if err != nil {
		return fmt.Errorf("Could not create tmp dir: %s", tempDir)
	}
	defer os.RemoveAll(tempDir)

	r := registry.New("https://auth.docker.io", "https://registry.hub.docker.com", h.imageName, h.imageTag)
	c := compress.New()
	d := downloader.New(tempDir, outFile, r, c, os.Stdout)

	if err := d.Run(); err != nil {
		return err
	}
	fmt.Println("Done.")
	return nil
}
