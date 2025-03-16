package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")
	router.MaxMultipartMemory = 8 << 20

	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cutomerResolver := aws.EndpointResolverWithOptionsFunc(func(sevice, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               "http://localhost:9000",
			HostnameImmutable: true,
			SigningRegion:     region,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(cutomerResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	uploader := manager.NewUploader(client)

	bucketName := "code"
	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		log.Printf("Bucket does not exist, creating it: %v", err)
		_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Printf("Failed to create bucket: %v", err)
		} else {
			log.Printf("Bucket created successfully")
		}
	}

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.POST("/", func(c *gin.Context) {

		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Faile to upload image"})
			return
		}

		f, openError := file.Open()

		if openError != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Faile to open image"})
			return
		}

		result, uploadError := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    types.ObjectCannedACLPublicRead,
		})

		if uploadError != nil {
			log.Printf("Upload error: %v", uploadError)
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Faile to upload image"})
			return
		}

		imageURL := fmt.Sprintf("http://localhost:9000/%s/%s", bucketName, file.Filename)
		log.Printf("Upload successful, URL: %s", imageURL)

		c.HTML(http.StatusOK, "index.html", gin.H{"image": result.Location})
	})
	router.Run()
}
