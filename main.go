package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// printUsage prints the usage for shoots3.
func printUsage() {
	fmt.Println("Syntax: shoot [flags] filename")
	flag.PrintDefaults()
	os.Exit(0)
}

// genKey generates a random string of characters with the specified length.
func genKey(length int) string {
	key := make([]byte, length)
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())

	for i := range key {
		key[i] = chars[rand.Intn(len(chars))]
	}

	return string(key)
}

// getContentType determines the ContentType of a file.
func getContentType(file *os.File) (string, error) {
	buffer := make([]byte, 512)

	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func main() {
	args := os.Args

	if len(args) < 2 {
		printUsage()
	}

	// Parse command line arguments
	key := flag.String("k", "", "custom key")
	length := flag.Int("l", 6, "generated url length")
	bucket := flag.String("b", "", "S3 bucket to upload the file")
	region := flag.String("r", "", "AWS region")
	flag.Parse()
	args = flag.Args()

	if len(args) != 1 {
		printUsage()
	}

	file := args[0]

	// Generate random key
	if *key == "" {
		generated := genKey(*length)
		key = &generated
	}

	// Get the default bucket
	if *bucket == "" {
		bucketEnv, ok := os.LookupEnv("SHOOTS3_DEFAULT_BUCKET")
		if !ok {
			fmt.Println("SHOOTS3_DEFAULT_BUCKET is not set")
			os.Exit(1)
		}

		bucket = &bucketEnv
	}

	// Get the default region
	if *region == "" {
		regionEnv, ok := os.LookupEnv("AWS_REGION")
		if !ok {
			fmt.Println("AWS_REGION is not set")
			os.Exit(1)
		}

		region = &regionEnv
	}

	reader, err := os.Open(file)
	if reader == nil {
		println("File does not exist")
		os.Exit(1)
	}

	contentType, err := getContentType(reader)
	if err != nil {
		panic(err.Error())
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("Unable to load SDK config, " + err.Error())
	}

	cfg.Region = *region

	svc := s3.New(cfg)
	obj := s3.PutObjectInput{
		Key:         key,
		Bucket:      bucket,
		Body:        reader,
		ContentType: &contentType,
	}
	req := svc.PutObjectRequest(&obj)

	_, err = req.Send()
	if err != nil {
		fmt.Printf("Failed to upload your file: %s", err.Error())
		os.Exit(1)
	}

	// URL of the uploaded file
	fmt.Printf("https://s3-%s.amazonaws.com/%s/%s\n", *region, *bucket, *key)
}
