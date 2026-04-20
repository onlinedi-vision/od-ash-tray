package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

type EtagState int

const (
	CannotOpenFile EtagState = iota
	CannotCreateEtag
	ValidEtag
	InvalidEtag
)

func shaFile(file fs.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isEtagValid(filePath string, httpWriter http.ResponseWriter, req *http.Request) (EtagState, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return CannotOpenFile, err
	}
	defer file.Close()

	etag, err := shaFile(file)
	if err != nil {
		return CannotCreateEtag, err
	}

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			httpWriter.WriteHeader(http.StatusNotModified)
			return ValidEtag, nil
		}
	}

	httpWriter.Header().Set("Etag", etag)
	return InvalidEtag, nil
}
