package main

import (
	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/windows2016fs/hydrator"
)

func main() {
	outDir, imageName, imageTag, noTarball := parseFlags()

	if err := hydrator.New(outDir, imageName, imageTag, noTarball).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		os.Exit(1)
	}
}

func parseFlags() (string, string, string, bool) {
	outDir := flag.String("outputDir", os.TempDir(), "Output directory for downloaded image")
	imageName := flag.String("image", "", "Name of the image to download")
	imageTag := flag.String("tag", "latest", "Image tag to download")
	noTarball := flag.Bool("noTarball", false, "Do not output image as a tarball")
	flag.Parse()

	return *outDir, *imageName, *imageTag, *noTarball
}
