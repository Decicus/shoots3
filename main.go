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
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
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

func objExists(svc *s3.S3, obj *s3.HeadObjectInput) bool {
	req := svc.HeadObjectRequest(obj)

	_, err := req.Send()
	if err != nil {
		return false
	}

	return true
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
	force := flag.Bool("f", false, "force override existing file")
	endpoint := flag.String("e", "", "Use a custom S3 endpoint (such as a MinIO deployment)")
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

	// Add custom resolver for a custom S3 endpoint
	// https://docs.aws.amazon.com/sdk-for-go/v2/api/aws/endpoints/#hdr-Using_Custom_Endpoints
	if *endpoint != "" {
		endpointResolver := func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		}

		cfg.EndpointResolver = aws.EndpointResolverFunc(endpointResolver)
	}

	svc := s3.New(cfg)

	// Ensure there isn't already a file with the same key
	if !*force {
		obj := s3.HeadObjectInput{
			Key:    key,
			Bucket: bucket,
		}
		exists := objExists(svc, &obj)
		if exists {
			fmt.Printf("File already exists with the same key: %s\n", *key)
			os.Exit(0)
		}
	}

	// Upload the file
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
	// TODO: Need to figure out how to deal with URLs on custom endpoints.
	fmt.Printf("https://s3-%s.amazonaws.com/%s/%s\n", *region, *bucket, *key)
}
