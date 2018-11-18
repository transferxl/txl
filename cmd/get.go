package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/encrypt"
	"github.com/spf13/cobra"
)

var (
	outputfile string
	decryptkey string
)

func init() {
	getCmd.Flags().StringVarP(&outputfile, "output", "o", "", "file for output")
	getCmd.Flags().StringVarP(&logfile, "log", "l", "", "file for logging")
	getCmd.Flags().StringVarP(&decryptkey, "decrypt", "d", "", "decryption phrase")
	getCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download transfer",
	Long:  `Download a transfer`,
	Example: `  txl get https://transferxl.com/<shorl-url>
  txl get -d secret https://transferxl.com/<encrypted-shorl-url>`,
	Run: func(cmd *cobra.Command, args []string) {

		shorturl := ""
		if len(args) == 0 {

			reader := bufio.NewReader(os.Stdin)
			var output []rune

			for {
				input, _, err := reader.ReadRune()
				if err != nil && err == io.EOF {
					break
				}
				output = append(output, input)
			}

			shorturl = strings.TrimSuffix(string(output), "\n")
		} else {
			shorturl = args[0]
		}

		shorturl = strings.TrimPrefix(shorturl, "https://transferxl.com/")

		result, err := getDownloadCredentials(shorturl)
		if err != nil {
			fmt.Println(err)
			return
		}

		output := outputfile
		if output == "" {
			output = result.Filename
		}

		// New SSE-C where the cryptographic key is derived from a password and the objectname + bucketname as salt
		var encryption encrypt.ServerSide
		if decryptkey != "" {
			encryption = encrypt.DefaultPBKDF([]byte(decryptkey), []byte(result.Bucket+result.Object))
		}

		start := time.Now()
		totalSize := get(result.AccessKey, result.SecretKey, result.Bucket, result.Object, output, encryption)
		elapsed := time.Since(start)

		if logfile != "" {
			export := fmt.Sprintf("get,%s,%s,%d,%.1fs,%.0f Mbit/s\n", result.Bucket, result.Object, totalSize, elapsed.Seconds(), float64(totalSize*8/1024/1024)/elapsed.Seconds())
			appendFile(logfile, export)
		}

		fmt.Printf("Downloaded '%s' in %.1fs (%.1f MB at %.0f Mbit/s)\n", output, elapsed.Seconds(), float64(totalSize)/1024.0/1024.0, float64(totalSize*8/1024/1024)/elapsed.Seconds())
	},
}

// Worker routine for downloading a part
func getWorker(minioClient *minio.Client, bucket, object string, parts <-chan partInput, results chan<- partOutput, encryption encrypt.ServerSide) {

	for p := range parts {

		options := minio.GetObjectOptions{}
		options.SetRange(p.start, p.end)
		options.ServerSideEncryption = encryption

		object, err := minioClient.GetObject(bucket, object, options)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Allocate buffer
		chunk := make([]byte, p.end-p.start)

		n, err := io.ReadFull(object, chunk)
		object.Close()
		if err != nil {
			fmt.Println(err)
			return
		} else if n < len(chunk) {
			fmt.Println("Received %d bytes, expected %d", n, len(chunk))
			return
		}

		results <- partOutput{part: p.part, chunk: chunk}
	}
}

type partInput struct {
	part  int64
	start int64
	end   int64
}

type partOutput struct {
	part  int64
	chunk []byte
}

func get(access, secret, bucket, object, file string, encryption encrypt.ServerSide) (totalSize int64) {

	const workers = 16

	// New returns an Amazon S3 compatible client object.
	s3Client, err := minio.New("s3.wasabisys.com", access, secret, true)
	if err != nil {
		log.Fatalln("Failed to create client", err)
	}

	options := minio.StatObjectOptions{}
	options.ServerSideEncryption = encryption

	stat, err := s3Client.StatObject(bucket, object, options)
	if err != nil {
		fmt.Println(err)
		return
	}

	// TODO: Make size dependent on overall size of object
	const size = 100 * 1024 * 1024

	var wg sync.WaitGroup
	parts := make(chan partInput)
	results := make(chan partOutput)

	// Start one go routine for each worker
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			getWorker(s3Client, bucket, object, parts, results, encryption)
		}()
	}

	ahead := int64(0)

	// Push parts onto input channel
	go func() {
		for part := int64(0); part*size < stat.Size; part++ {
			for atomic.LoadInt64(&ahead) > workers {
				time.Sleep(10 * time.Millisecond) // loop around until the queue has catched up
			}
			parts <- partInput{part: part, start: int64(part * size), end: int64(math.Min(float64(stat.Size), float64((part+1)*size)))}
			atomic.AddInt64(&ahead, 1)
		}

		// Close input channel
		close(parts)
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results) // Close output channel
	}()

	part := int64(0)
	partMap := make(map[int64]partOutput)

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	for r := range results {
		if part == r.part {
			chunk, ok := r.chunk, false
			for {
				if verbose {
					fmt.Printf(".")
				}
				atomic.AddInt64(&ahead, -1)
				n, err := f.Write(chunk)
				if err != nil {
					fmt.Println(err)
					return
				} else if n < len(chunk) {
					fmt.Println("Too little data written")
					return
				}

				totalSize += int64(len(chunk))
				part++
				// Check for more parts received earlier
				if r, ok = partMap[part]; !ok {
					break
				}
				delete(partMap, part) // remove element from map (free resources)
				chunk = r.chunk
			}
		} else {
			partMap[r.part] = r
		}
	}
	if verbose {
		fmt.Println()
	}
	return
}
