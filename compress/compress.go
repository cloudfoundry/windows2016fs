package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Compressor struct{}

func New() *Compressor {
	return &Compressor{}
}

func (c *Compressor) WriteTgz(srcDir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	return writeTar(srcDir, gzw)
}

func writeTar(srcDir string, dest io.Writer) error {
	dir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	tw := tar.NewWriter(dest)
	defer tw.Close()

	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, fi := range files {
		if fi.IsDir() {
			return &ErrInvalidSource{path: srcDir}
		}
		hdr := tarHeader(fi.Name(), fi.Size())
		if err := addTarFile(filepath.Join(dir, fi.Name()), &hdr, tw); err != nil {
			return err
		}
	}

	return nil
}

const c_ISREG = 0100000 // Regular file
func tarHeader(filename string, size int64) tar.Header {
	return tar.Header{
		Name:     filename,
		Mode:     0644 | c_ISREG,
		Size:     size,
		Typeflag: tar.TypeReg,
	}
}

func addTarFile(file string, hdr *tar.Header, tw *tar.Writer) error {
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(tw, f)
	if err != nil {
		return err
	}

	return nil
}

type ErrInvalidSource struct {
	path string
}

func (e *ErrInvalidSource) Error() string {
	return fmt.Sprintf("source directory %s cannot contain sub directories", e.path)
}
