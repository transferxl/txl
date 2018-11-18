package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/encrypt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"time"
)

var (
	echoTimes  int
	user       string
	password   string
	message    string
	encryptkey string
	recipients string
	storage    string
	logfile    string
	verbose    bool
)

func init() {
	putCmd.Flags().StringVarP(&user, "user", "u", "", "user account")
	putCmd.Flags().StringVarP(&password, "password", "p", "", "password for account")
	putCmd.Flags().StringVarP(&message, "message", "m", "", "message for the transfer")
	putCmd.Flags().StringVarP(&encryptkey, "encrypt", "e", "", "encryption phrase")
	putCmd.Flags().StringVarP(&recipients, "recipients", "r", "", "email address of recipient(s)")
	putCmd.Flags().StringVarP(&storage, "storage", "s", "", "storage region for the transfer")
	putCmd.Flags().StringVarP(&logfile, "log", "l", "", "file for logging")
	putCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	putCmd.MarkFlagRequired("user")
	putCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(putCmd)
}

var putCmd = &cobra.Command{
	Use:   "put ",
	Short: "Upload transfer",
	Long:  `Create a transfer`,
	Run: func(cmd *cobra.Command, args []string) {

		start := time.Now()

		credentials, err := getUploadCredentials(user, password)
		if err != nil {
			fmt.Println(err)
			return
		}

		key, _ := GenerateRandomString(32)

		// New SSE-C where the cryptographic key is derived from a password and the objectname + bucketname as salt
		var encryption encrypt.ServerSide
		if encryptkey != "" {
			encryption = encrypt.DefaultPBKDF([]byte(encryptkey), []byte(credentials.Bucket+key))
		}

		name, n := "transfer", int64(0)

		if len(args) == 0 {
			// Read from stdin.
			n, err = putStream(credentials, key, os.Stdin, encryption)
		} else {
			n, err = put(credentials, args[0], key, encryption)
			_, name = path.Split(args[0])
		}
		if err != nil {
			fmt.Print(err)
			return
		}
		if verbose {
			fmt.Println("Uploaded", n)
		}

		transfer, err := createTransfer(user, password, credentials.Bucket, key, name, message, n, encryptkey != "")
		if err != nil {
			fmt.Println(err)
			return
		}

		if logfile != "" {
			elapsed := time.Since(start)
			export := fmt.Sprintf("put,%s,%s,%d,%.1fs,%.0f Mbit/s\n", credentials.Bucket, key, n, elapsed.Seconds(), float64(n*8/1024/1024)/elapsed.Seconds())
			appendFile(logfile, export)
		}

		// Just output the shorturl
		fmt.Println(transfer.Shorturl)
	},
}

func putStream(credentials uploadCredentialsResult, key string, r io.Reader, encryption encrypt.ServerSide) (n int64, err error) {

	// New returns an Amazon S3 compatible client object.
	minioClient, err := minio.New(credentials.Endpoint, credentials.AccessKey, credentials.SecretKey, true)
	if err != nil {
		fmt.Println("Failed to create client", err)
		return
	}

	options := minio.PutObjectOptions{}
	options.NumThreads = 50 // TODO: Get nr of CPUs
	options.ServerSideEncryption = encryption

	n, err = minioClient.PutObject(credentials.Bucket, key, r, -1, options)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func put(credentials uploadCredentialsResult, localFile, key string, encryption encrypt.ServerSide) (n int64, err error) {

	if verbose {
		fmt.Println("Uploading", localFile)
	}

	// New returns an Amazon S3 compatible client object.
	minioClient, err := minio.New(credentials.Endpoint, credentials.AccessKey, credentials.SecretKey, true)
	if err != nil {
		fmt.Println("Failed to create client", err)
		return
	}

	options := minio.PutObjectOptions{}
	options.ServerSideEncryption = encryption
	options.NumThreads = 50 // TODO: Get nr of CPUs

	n, err = minioClient.FPutObject(credentials.Bucket, key, localFile, options)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

// GenerateRandomString returns a URL-safe random string.
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func appendFile(filename, line string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed opening file", err)
	}
	defer file.Close()

	_, err = file.WriteString(line)
	if err != nil {
		fmt.Println("Failed writing to file", err)
	}
}
