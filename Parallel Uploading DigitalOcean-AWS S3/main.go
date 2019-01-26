package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/briandowns/spinner"
)

var sortedSlice []string
var f, err = os.OpenFile("./progressLog.log",
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

func getFilesToUpload() error {

	dir, err := os.Open("./")
	if err != nil {
		return err
	}
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	dir.Close()

	f, err := os.Open("./progressLog.log")
	if err != nil {
		return err
	}

	contentAsBytes, err := ioutil.ReadFile("./progressLog.log")
	contentToString := string(contentAsBytes)

	defer f.Close()

	for _, filename := range names {

		if strings.Contains(contentToString, filename) {
			fileCountNow++
			fmt.Printf("\nFile already exists : %s", filename)

		} else {
			sortedSlice = append(sortedSlice, filename)
		}
	}

	return nil

}

func logg(singlefile string) {
	//initializes the logger file and logs in the next step
	logger := log.New(f, "ProjectName | ", log.LstdFlags)
	logger.Printf("Succesfully logged : %s \n", singlefile)

}

func uploadDir(names []string) error {

	// Copy names to a channel for workers to consume. Close the
	// channel so that workers stop when all work is complete.
	namesChan := make(chan string, len(names))
	for _, name := range names {
		namesChan <- name
	}
	close(namesChan)

	// Create a maximum of 8 workers
	workers := 12
	if len(names) < workers {
		workers = len(names)
	}

	errChan := make(chan error, 1)
	resChan := make(chan *s3manager.UploadOutput, len(names))
	nameChan := make(chan string, len(names))

	// Run workers
	for i := 0; i < workers; i++ {

		//the epic goroutine
		go func() {

			// puts the goroutine to sleep so that it only starts after few function execution in the func main()
			time.Sleep(time.Second * 1)
			// Consume work from namesChan. Loop will end when no more work.
			for name := range namesChan {

				fileCountNow++

				//opens file by concatinating directory + name of the file
				file, err := os.Open(filepath.Join("./", name))
				if err != nil {
					select {
					case errChan <- err:
						// will break parent goroutine out of loop
						fmt.Printf("\nFailed to open file %s, %v", name, err)

					default:
						// don't care, first error wins
					}
					return
				}
				endpoint := "sgp1.digitaloceanspaces.com"
				region := "sgp1"

				sess := session.Must(session.NewSession(&aws.Config{
					Endpoint: &endpoint,
					Region:   &region,
				}))

				// Create an uploader with the session and default options
				uploader := s3manager.NewUploader(sess)

				myBucket := "*_your_bucket_name_*"
				myFileName := name

				updateSpinner()

				result, err := uploader.Upload(&s3manager.UploadInput{
					Bucket: aws.String(myBucket),
					Key:    aws.String(myFileName),
					Body:   file,

					// sets your file perimission to public read, remove to make it private by default
					ACL: aws.String("public-read"),
				})

				//logs your upload history to logger file upon successful upload
				logg(name)

				//stops the spinner instance
				s.Stop()

				//closes the opened
				file.Close()

				if err != nil {
					select {
					case errChan <- err:
						// will break parent goroutine out of loop
					default:
						// don't care, first error wins
					}
					return
				}
				resChan <- result
				nameChan <- myFileName
			}
		}()
	}

	// Collect results from workers

	for i := 0; i < len(names); i++ {
		select {
		// case res := <-resChan:
		// 	log.Println(res)
		case res := <-nameChan:
			{
				fmt.Printf("\n\t\tSuccesfully uploaded : %s", res)
			}
		case err := <-errChan:
			return err
		}
	}
	return nil
}

var s = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
var name string
var totalFiles = 0
var fileCountNow = 0
var timer = 1 * time.Microsecond

func main() {

	start := time.Now().UTC()
	fmt.Println("--Begin--")

	totFile("./")
	getFilesToUpload()

	uploadDir(sortedSlice)
	elapsed := time.Since(start)
	fmt.Printf("\n--Finished--\n")

	fmt.Println("\n\nElapsed time :", elapsed)

}

func totFile(dir string) {
	files, _ := ioutil.ReadDir(dir)
	totalFiles = len(files)
	fmt.Printf("\nTotal Number of Files : %d\n", totalFiles)
}

func updateSpinner() {

	s.Suffix = "  : (" + strconv.Itoa(fileCountNow) + "/" + strconv.Itoa(totalFiles) + ")" // Append text after the spinner
	// Set the spinner color to red
	s.Start() // Update spinner to use a different character set
	s.UpdateSpeed(100 * time.Millisecond)

}
