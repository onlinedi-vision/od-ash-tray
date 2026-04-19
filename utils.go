package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

func shaFile(file fs.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isEtagValid(filePath string, httpWriter http.ResponseWriter, req *http.Request) (bool, string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("   %s cannot open file: %s\n", err, filePath)

		httpWriter.WriteHeader(500)
		fmt.Fprintf(httpWriter, "500 Failed to open file.\r\n")

		// Same remark as below...
		return true, ""
	}
	defer file.Close()

	etag, err := shaFile(file)
	if err != nil {
		fmt.Printf("   %s cannot create etag for:\n", err)

		httpWriter.WriteHeader(500)
		fmt.Fprintf(httpWriter, "500 Failed to create etag \r\n")

		// Hmmm... Unfortunately, I can't think of a better value to return here.
		// Saying that the etag is valid here means that we just respond
		// with 500 higher up, and no more processing returns. But... yeah
		// it's sort of unintuitive...
		return true, ""
	}

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			httpWriter.WriteHeader(http.StatusNotModified)
			return true, etag
		}
	}

	return false, etag
}
