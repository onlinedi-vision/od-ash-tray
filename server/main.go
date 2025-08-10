package main
import (
	"log"
	"net/http"
	"fmt"
	"time"
	"os"
	"io"
	"strings"
	"crypto/tls"
	"encoding/hex"
	"crypto/aes"
	"github.com/google/uuid"
	"crypto/sha256"
)

var (
	ashID = os.Getenv("ASH_TRAY_ID")
	ashTrayDir = os.Getenv("ASH_TRAY_DIRECTORY")
	ashHighSignature = os.Getenv("ASH_HIGH_SIGNATURE")
	ashLowSignature = os.Getenv("ASH_LOW_SIGNATURE")
	ashKey = []byte(os.Getenv("ASH_TRAY_KEY"))
)

var (
	CertFilePath = "/etc/letsencrypt/live/onlinedi.vision/fullchain.pem"
	KeyFilePath  = "/etc/letsencrypt/live/onlinedi.vision/privkey.pem"
)

const (
	encSize = 1024000
	BlockSize = 16
)

func createNewAshTray(req *http.Request) (string, string) {
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
	newUUID, err2 := uuid.NewV7()
	if err != nil || err2 != nil {
		return "", ""
	}
	
	sha := sha256.Sum256([]byte(dirUUID[:]))
	directory := hex.EncodeToString(sha[:])
	filename := hex.EncodeToString(newUUID[:])
	return fmt.Sprintf("%s/%s/%s.%s", ashID, directory, filename, extension), directory
}

func writeToFile(filePath string, directory string, data []byte) {

	_ = os.Mkdir(fmt.Sprintf("%s/%s/%s", ashTrayDir, ashID, directory), os.ModePerm)
		
	ashFile, err := os.OpenFile(fmt.Sprintf("%s/%s", ashTrayDir,filePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
    	log.Fatal(err)
	}
	defer ashFile.Close()

	ashFile.Write(data)
}

func encryptData(data string) ([]byte, error) {

	aesBlock,err := aes.NewCipher(ashKey)
	if err != nil {
		return nil, err		
	}
	encryptedData := make([]byte, encSize)

	bytesOfData := make([]byte,encSize)
	copy(bytesOfData[:], data)

	for iter := range encSize / BlockSize {
		offset := iter * BlockSize
		aesBlock.Encrypt(encryptedData[offset:offset+BlockSize], bytesOfData[offset:offset+BlockSize])
	}
	
	return encryptedData, nil
}

func decryptData(data []byte) ([]byte, error) {
	
	aesBlock,err := aes.NewCipher(ashKey)
	if err != nil {
		return nil, err		
	}
	decryptedData := make([]byte, encSize)

	for iter := range encSize / BlockSize {
		offset := iter * BlockSize
		aesBlock.Decrypt(decryptedData[offset:offset+BlockSize], data[offset:offset+BlockSize])
	}
	
	return decryptedData, nil
}

func fileDownload(httpWriter http.ResponseWriter, req *http.Request) {
	filePath := fmt.Sprintf("%s%s", ashTrayDir, req.URL.Path)
	fmt.Printf(" + fileDownload(): filePath=%s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("   %s file doesn't exist\n", filePath)
		fmt.Fprintf(httpWriter, "HTTP/1.1 404 Not Found \r\n")
		return
	}
	defer file.Close()

	buffer := make([]byte, encSize)

	fmt.Println("   sending file")

	for {
	    readTotal, err := file.Read(buffer)
	    if err != nil {
	        if err != io.EOF {
	            fmt.Println(err)
	        }
	        break
	    }
		decryptedData, err := decryptData(buffer[:readTotal])
		if err != nil {
			break
		}
	    fmt.Fprintf(httpWriter, "%s", string(decryptedData[:]))
	    
	}
	
}

func fileUpload(httpWriter http.ResponseWriter, req *http.Request) {
	filePath, directory := createNewAshTray(req)
	fmt.Printf(" + fileUpload(): filePath=%s   directory=%s\n", filePath, directory)

	for _,value := range req.MultipartForm.File {
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
		        data, err := encryptData(string(buffer[:readTotal])) 
				if err != nil {
					fmt.Println(err)
					return
				}
				writeToFile(filePath, directory, data)
		    }
		}
	}
	fmt.Fprintf(httpWriter, "HTTP/1.1 201 Created \r\n\r\n%s", filePath)
}

func higherTrayTimer() func() {
	start := time.Now()
	return func() {
		fmt.Printf("Duration: %v\n\n", time.Since(start))
	}
}

func higherTray(httpWriter http.ResponseWriter, req *http.Request) {
	defer higherTrayTimer()()
	
	fmt.Printf("[%s] %s: %s\n", req.Method, req.RemoteAddr, req.URL)
	req.ParseMultipartForm(400000000)

	if req.Method == "GET" {
		fileDownload(httpWriter, req);	
	} else if req.Method == "POST" && req.URL.Path == "/upload" {
		fileUpload(httpWriter, req)		
	} else {
		fmt.Fprintf(httpWriter, "HTTP/1.1 400 Bad Request \r\n\r\nPlease GET for download or POST /upload for multipart form upload.")
	}
}

func main() {

	err := os.Mkdir(fmt.Sprintf("%s/%s", ashTrayDir, ashID), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	serverTLSCert, err := tls.LoadX509KeyPair(CertFilePath, KeyFilePath)
	if err != nil {
		fmt.Println("COULD NOT LOAD TLS CERTIFICATE... BAILING OUT...")
		return 
	}
	
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
	}
		
	tray_server := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("ASH_TRAY_PORT")),
		Handler:        http.HandlerFunc(higherTray),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig, 
	}
	log.Fatal(tray_server.ListenAndServeTLS("",""))
}
