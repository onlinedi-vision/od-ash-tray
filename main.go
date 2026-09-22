package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ashID            = "ash"
	ashTrayDir       = os.Getenv("ASH_TRAY_DIRECTORY")
	ashHighSignature = os.Getenv("ASH_HIGH_SIGNATURE")
	ashLowSignature  = os.Getenv("ASH_LOW_SIGNATURE")
	ashKey           = []byte(os.Getenv("ASH_TRAY_KEY"))
)

const (
	MaxFormMemorySize = 20000000
	encSize           = 16096
	BlockSize         = 16
)

func createNewAshTray(req *http.Request) (string, string, string, string) {
	var extension string

	for _, value := range req.MultipartForm.File {
		for _, part := range value {
			split := strings.Split(part.Filename, ".")
			extension = split[len(split)-1]
			break
		}
		break
	}

	dirUUID, err := uuid.NewV7()
	fileUUID, err2 := uuid.NewV7()
	if err != nil || err2 != nil {
		return "", "", "", ""
	}

	dirSha := sha256.Sum256([]byte(dirUUID[:]))
	unhashedFilename := hex.EncodeToString(fileUUID[:])
	fileSha := sha256.Sum256([]byte(unhashedFilename[:]))

	directory := hex.EncodeToString(dirSha[:])
	filename := hex.EncodeToString(fileSha[:])

	return fmt.Sprintf("%s/%s/%s.%s", ashID, directory, filename, extension), directory, unhashedFilename, extension
}

func writeToFile(filePath string, directory string, data []byte) {
	_ = os.Mkdir(fmt.Sprintf("%s/%s/%s", ashTrayDir, ashID, directory), os.ModePerm)

	ashFile, err := os.OpenFile(fmt.Sprintf("%s/%s", ashTrayDir, filePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: wow what a retarded fucking way to do things...
	defer ashFile.Close()

	ashFile.Write(data)
}

func encryptData(data string, uf string) ([]byte, error) {
	keySha := sha256.Sum256([]byte(fmt.Sprintf("%s%s", ashKey, uf)))

	aesBlock, err := aes.NewCipher([]byte(hex.EncodeToString(keySha[:]))[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(data), nil)

	encryptedLen := len(nonce) + len(ciphertext)
	result := make([]byte, 4+encryptedLen)
	binary.BigEndian.PutUint32(result[:4], uint32(encryptedLen))
	copy(result[4:], nonce)
	copy(result[4+len(nonce):], ciphertext)

	return result, nil
}

func decryptData(data []byte, uf string) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	expectedLen := int(binary.BigEndian.Uint32(data[:4]))
	if len(data) < 4+expectedLen {
		return nil, fmt.Errorf("incomplete data")
	}

	encryptedData := data[4 : 4+expectedLen]

	keySha := sha256.Sum256([]byte(fmt.Sprintf("%s%s", ashKey, uf)))

	aesBlock, err := aes.NewCipher([]byte(hex.EncodeToString(keySha[:]))[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func fileDownload(httpWriter http.ResponseWriter, req *http.Request) {
	urlSplit := strings.Split(req.URL.Path, "/")
	currentAshId := urlSplit[1]
	currentFileDirectory := urlSplit[2]

	filenameSplit := strings.Split(urlSplit[3], ".")
	uf := filenameSplit[0]
	ext := filenameSplit[1]

	fileSha := sha256.Sum256([]byte(uf[:]))

	hashedFilename := hex.EncodeToString(fileSha[:])
	filePath := fmt.Sprintf("%s/%s/%s/%s.%s", ashTrayDir, currentAshId, currentFileDirectory, hashedFilename, ext)
	fmt.Printf(" + fileDownload(): filePath=%s\n", filePath)

	switch state, err := isEtagValid(filePath, httpWriter, req); state {
	case CannotOpenFile:
		fmt.Printf("   %s cannot open file: %s\n", err, filePath)

		httpWriter.WriteHeader(500)
		fmt.Fprintf(httpWriter, "500 Failed to open file.\r\n")
		return
	case CannotCreateEtag:
		fmt.Printf("   %s cannot create etag for:\n", err)
		httpWriter.WriteHeader(500)
		fmt.Fprintf(httpWriter, "500 Failed to create etag \r\n")
		return
	case ValidEtag:
		fmt.Printf("   etag HIT\n")
		return
	case InvalidEtag:
		fmt.Printf("   etag MISS\n")
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("   %s file doesn't exist\n", filePath)

		httpWriter.WriteHeader(404)
		fmt.Fprintf(httpWriter, "404 Not Found \r\n")
		return
	}
	defer file.Close()

	buffer := make([]byte, 65536)
	var leftover []byte

	fmt.Println("   sending file")

	httpWriter.WriteHeader(200)

	for {
		readTotal, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			httpWriter.WriteHeader(500)
			fmt.Fprintf(httpWriter, "Failed at reading from file.")
			break
		}

		if readTotal == 0 {
			break
		}

		data := append(leftover, buffer[:readTotal]...)
		leftover = nil

		offset := 0
		for offset+4 <= len(data) {
			blockLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
			if offset+4+blockLen > len(data) {
				leftover = data[offset:]
				break
			}

			decryptedData, err := decryptData(data[offset : offset+4+blockLen], uf)
			if err != nil {
				fmt.Println(err)
				httpWriter.WriteHeader(500)
				fmt.Fprintf(httpWriter, "Failed at file decryption.")
				return
			}

			httpWriter.Write(decryptedData)
			offset += 4 + blockLen
		}

		if offset < len(data) && leftover == nil {
			leftover = data[offset:]
		}

		if err == io.EOF {
			break
		}
	}
}

func fileUpload(httpWriter http.ResponseWriter, req *http.Request) {
	filePath, directory, uf, ext := createNewAshTray(req)
	fmt.Printf(" + fileUpload(): filePath=%s   directory=%s\n", filePath, directory)

	for _, value := range req.MultipartForm.File {
		for _, file_part := range value {
			file, err := file_part.Open()
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			buffer := make([]byte, encSize)

			for {
				readTotal, err := file.Read(buffer)
				if err != nil {
					if err != io.EOF {
						fmt.Println(err)
					}
					break
				}
				data, err := encryptData(string(buffer[:readTotal]), uf)
				if err != nil {
					fmt.Println(err)
					return
				}
				writeToFile(filePath, directory, data)
			}
		}
	}
	httpWriter.WriteHeader(201)
	fmt.Fprintf(httpWriter, "%s/%s/%s.%s", ashID, directory, uf, ext)
}

func higherTrayTimer() func() {
	start := time.Now()
	return func() {
		fmt.Printf("Duration: %v\n\n", time.Since(start))
	}
}

func ashGet(httpWriter http.ResponseWriter, req *http.Request) {
	fileDownload(httpWriter, req)
}

func higherTray(httpWriter http.ResponseWriter, req *http.Request) {
	defer higherTrayTimer()()

	fmt.Printf("@ %s : \n > [%s] %s: %s\n", time.Now(), req.Method, req.RemoteAddr, req.URL)
	req.ParseMultipartForm(MaxFormMemorySize)

	httpWriter.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method == "GET" {
		if req.URL.Path == "/ash/ping" {
			fmt.Fprintf(httpWriter, "ping")
		} else {
			ashGet(httpWriter, req)
		}
	} else if req.Method == "POST" && req.URL.Path == "/ash/upload" {
		fileUpload(httpWriter, req)
	} else {
		httpWriter.WriteHeader(400)
		fmt.Fprintf(httpWriter, "Please GET for download or POST /upload for multipart form upload.")
	}
}

func main() {
	if ashTrayDir == "" {
		fmt.Println("ASH_TRAY_DIRECTORY env var MUST be set")
		return
	}

	err := os.Mkdir(fmt.Sprintf("%s/%s", ashTrayDir, ashID), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	tray_server := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%s", os.Getenv("ASH_TRAY_PORT")),
		Handler:        http.HandlerFunc(higherTray),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(tray_server.ListenAndServe())
}
