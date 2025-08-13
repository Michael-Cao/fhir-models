package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Download(url string, tmpDir string) (*string, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	items := strings.Split(url, "/")
	filename := fmt.Sprintf("%s/%s", tmpDir, items[len(items)-1])
	out, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to create file %v: %v\n", filename, err)
		return nil, err
	}
	defer out.Close()

	fmt.Printf("Downloading %s\n", url)

	totalBytes := res.ContentLength
	var downloadedBytes int64 = 0

	buffer := make([]byte, 32*1024)
	for {
		n, err := res.Body.Read(buffer)
		if n > 0 {
			_, err := out.Write(buffer[:n])
			if err != nil {
				return nil, err
			}
			downloadedBytes += int64(n)
			fmt.Printf("\rDownloading... %d%% complete", 100*downloadedBytes/totalBytes)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("\rDownloaded %s successfully", url)
	return &filename, nil
}
