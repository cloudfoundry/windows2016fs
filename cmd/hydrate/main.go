package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/windows2016fs/compress"
	"code.cloudfoundry.org/windows2016fs/hydrator"
	"code.cloudfoundry.org/windows2016fs/registry"
)

func main() {
	if err := mainBody(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		os.Exit(1)
	}
}

func mainBody() error {
	outDir, imageName, imageTag := parseFlags()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return errors.New("Could not create output directory")
	}

	if imageName == "" {
		return errors.New("No image name provided")
	}

	nameParts := strings.Split(imageName, "/")
	if len(nameParts) != 2 {
		return errors.New("Invalid image name")
	}
	outFile := filepath.Join(outDir, fmt.Sprintf("%s-%s.tgz", nameParts[1], imageTag))

	tempDir, err := ioutil.TempDir("", "hydrate")
	if err != nil {
		return fmt.Errorf("Could not create tmp dir: %s", tempDir)
	}
	defer os.RemoveAll(tempDir)

	r := registry.New("https://auth.docker.io", "https://registry.hub.docker.com", imageName, imageTag)
	c := compress.New()
	h := hydrator.New(tempDir, outFile, r, c, os.Stdout)

	if err := h.Run(); err != nil {
		return err
	}
	fmt.Println("Done.")
	return nil
}

func parseFlags() (string, string, string) {
	outDir := flag.String("outputDir", os.TempDir(), "Output directory for downloaded image")
	imageName := flag.String("image", "", "Name of the image to download")
	imageTag := flag.String("tag", "latest", "Image tag to download")
	flag.Parse()

	return *outDir, *imageName, *imageTag
}
