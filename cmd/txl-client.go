package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

// TODO: Reenable HTTPS
const serverUrl = "http://txl.transferxl.com/"

type uploadCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type uploadCredentialsResult struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accesskey"`
	SecretKey string `json:"secretkey"`
	Bucket    string `json:"bucket"`
}

func getUploadCredentials(username, password string) (result uploadCredentialsResult, err error) {
	bytesRepresentation, err := json.Marshal(uploadCredentials{username, password})
	if err != nil {
		return
	}

	resp, err := http.Post(serverUrl+"uploadCredentials", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&result)
	} else {
		err = errors.New("Bad credentials")
	}
	return
}

type createTransferRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Bucket       string `json:"bucket"`
	Object       string `json:"object"`
	Size         int64  `json:"size"`
	Encrypted    bool   `json:"encrypted"`
	Filename     string `json:"filename"`
	Message      string `json:"message"`
	TransferType string `json:"transfertype"`
}

type createTransferResult struct {
	Shorturl string `json:"shorturl"`
}

func createTransfer(username, password, bucket, key, filename, message string, size int64, encrypted bool) (result createTransferResult, err error) {
	bytesRepresentation, err := json.Marshal(createTransferRequest{Username: username, Password: password, Bucket: bucket, Object: key, Size: size,
		Encrypted: encrypted, Filename: filename, Message: message, TransferType: "link"})
	if err != nil {
		return
	}

	resp, err := http.Post(serverUrl+"createTransfer", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&result)
	} else {
		err = errors.New("Bad credentials")
	}
	return
}

type downloadCredentials struct {
	Shorturl string `json:"shorturl"`
}

type downloadCredentialsResult struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accesskey"`
	SecretKey string `json:"secretkey"`
	Bucket    string `json:"bucket"`
	Object    string `json:"object"`
	Filename  string `json:"filename"`
}

func getDownloadCredentials(shorturl string) (result downloadCredentialsResult, err error) {

	bytesRepresentation, err := json.Marshal(downloadCredentials{shorturl})
	if err != nil {
		return
	}

	resp, err := http.Post(serverUrl+"downloadCredentials", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&result)
	} else {
		err = errors.New("Bad shorturl")
	}
	return
}

type listTransfersRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type listTransfersResult struct {
	Username  string     `json:"username"`
	Transfers []Transfer `json:"transfers"`
}

type Transfer struct {
	Shorturl     string `json:"shorturl"`
	Bucket       string `json:"bucket"`
	Object       string `json:"object"`
	Filename     string `json:"filename"`
	Username     string `json:"username"`
	Message      string `json:"message"`
	TransferType string `json:"transfertype"`
	Size         int64  `json:"size"`
	Encrypted    bool   `json:"encrypted"`
	CreationDate string `json:"creationdate"`
	Expiry       string `json:"expiry"`
}

func listTransfers(username, password string) (transfers []Transfer, err error) {
	bytesRepresentation, err := json.Marshal(listTransfersRequest{Username: username, Password: password})
	if err != nil {
		return
	}

	resp, err := http.Post(serverUrl+"listTransfers", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		result := listTransfersResult{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		transfers = result.Transfers
	} else {
		err = errors.New("Bad credentials")
	}
	return
}
