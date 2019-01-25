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

var i = 0
var p = 0
var totalFiles = 0
var fileCountNow = 0
var s = spinner.New(spinner.CharSets[9], 100*time.Millisecond) // Build our new spinner

func main() {

	s.Color("red") // Update the speed the spinner spins at
	s.Prefix = ""

	totalFiles = 0
	fileCountNow = 0
	fmt.Println("-Begin-")
	totFile("./")
	readEachFile()
	logg("progressLog.log")
	fmt.Println("Updating and uploading LogFiles")
	Uploader("progressLog.log")
	fmt.Println("-Completed-")
}

func readEachFile() error {

	var files []string

	root := "./"

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files {

		if file != "./" {
			fileCountNow++
			fmt.Println(file)
			isAlreadyAdded(file)
		}

	}
	return nil

}

func isAlreadyAdded(filename string) error {

	f, err := os.Open("./progressLog.log")
	if err != nil {
		return err
	}

	contentAsBytes, err := ioutil.ReadFile("./progressLog.log")
	contentToString := string(contentAsBytes)

	defer f.Close()

	if strings.Contains(contentToString, filename) {
		fmt.Printf("( %d / %d ) file already exists\n", fileCountNow, totalFiles)
		return err
	} else {
		Uploader(filename)
	}

	return nil

}

func Uploader(singlefile string) error {

	// https://*bucketnanme*.*s3 region*.digitaloceanspaces.com
	// The session the S3 Uploader will use
	endpoint := "sgp1.digitaloceanspaces.com"
	region := "sgp1"
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint: &endpoint,
		Region:   &region,
	}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	filename := "./" + singlefile
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	myBucket := "bucketname" //enter your bucketname here*
	myString := filename

	//start or Restart Spinner
	updateSpinner()

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(myString),
		Body:   f,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	logg(singlefile)
	// fmt.Printf("------(%d/%d)----------file uploaded to, %s\n", fileCountNow, totalFiles, aws.StringValue(&result.Location))
	s.Stop()
	// }
	fmt.Printf(" ✔ (%d/%d) - Successfully Uploaded : %s\n", fileCountNow, totalFiles, aws.StringValue(&result.Location))
	return nil
}

func logg(singlefile string) {

	f, err := os.OpenFile("./progressLog.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// log.Println(err)
	}

	defer f.Close()

	logger := log.New(f, " Logger | ", log.LstdFlags)
	logger.Printf("Succesfully uploaded : %s \n", singlefile)

}

func totFile(dir string) {
	files, _ := ioutil.ReadDir(dir)
	totalFiles = len(files)
	fmt.Printf("Total Number of Files : %d\n", totalFiles)
}

func updateSpinner() {

	s.Suffix = "  : (" + strconv.Itoa(fileCountNow) + "/" + strconv.Itoa(totalFiles) + ")" // Append text after the spinner
	// Set the spinner color to red
	s.Start() // Update spinner to use a different character set
	s.UpdateSpeed(100 * time.Millisecond)

}
