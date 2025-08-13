package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Unzip(archive string) error {

	fmt.Printf("unzipping %s...\n", archive)

	a, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	items := strings.Split(archive, "/")
	outDir := strings.Join(items[0:len(items)-1], "/")

	for _, f := range a.File {
		filename := filepath.Join(outDir, f.Name)
		fmt.Println(filename)

		dstFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		defer fileInArchive.Close()

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		if strings.HasSuffix(filename, ".zip") {
			Unzip(filename)
		}
	}

	fmt.Printf("unzipped %s\n", archive)
	return nil
}
