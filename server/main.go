package main
import (
	"log"
	"net/http"
	"fmt"
	"time"
	"os"
	"io"
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

const (
	encSize = 4096
	BlockSize = 16
)

func createNewAshTray(extension string) (string, string) {
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

func fileUpload(httpWriter http.ResponseWriter, req *http.Request) {
	filePath, directory := createNewAshTray("gif")

	for key,value := range req.MultipartForm.File {
		fmt.Println(key, " : ")
		for _, file_part := range value {
			file, err := file_part.Open()
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

		    b := make([]byte, encSize)

		    for {
		        readTotal, err := file.Read(b)
		        if err != nil {
		            if err != io.EOF {
		                fmt.Println(err)
		            }
		            break
		        }
		        data, err := encryptData(fmt.Sprintf("%s%s", ashHighSignature, string(b[:readTotal]))) 
				if err != nil {
					fmt.Println(err)
					return
				}
				writeToFile(filePath, directory, data)
		    }
		}
	}
	fmt.Fprintf(httpWriter, filePath)
}

func higherTray(httpWriter http.ResponseWriter, req *http.Request) {
	fmt.Printf("[%s] %s: %s\n", req.Method, req.Host, req.URL)
	req.ParseMultipartForm(4096000)

	if req.Method == "POST" && req.URL.Path == "/upload" {
		fileUpload(httpWriter, req)		
	} else {
		fmt.Fprintf(httpWriter, "WRONG")
	}
}

func main() {
	fmt.Println(ashTrayDir)

	err := os.Mkdir(fmt.Sprintf("%s/%s", ashTrayDir, ashID), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
		
	tray_server := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("ASH_TRAY_PORT")),
		Handler:        http.HandlerFunc(higherTray),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(tray_server.ListenAndServe())
}
