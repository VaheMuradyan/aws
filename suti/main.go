package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
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
		c.JSON(http.StatusOK, gin.H{"message": "get"})
	})

	router.POST("/", func(c *gin.Context) {

		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Faile to upload image"})
			return
		}

		f, openError := file.Open()

		if openError != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Faile to open image"})
			return
		}

		defer f.Close()

		fileExt := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("%d%s", time.Now().Unix(), fileExt)

		_, uploadError := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(fileName),
			Body:        f,
			ACL:         types.ObjectCannedACLPublicRead,
			ContentType: aws.String(http.DetectContentType([]byte{})),
		})

		if uploadError != nil {
			log.Printf("Upload error: %v", uploadError)
			c.JSON(http.StatusOK, gin.H{"error": "Faile to upload image"})
			return
		}

		imageURL := fmt.Sprintf("http://localhost:9000/%s/%s", bucketName, file.Filename)
		log.Printf("Upload successful, URL: %s", imageURL)

		c.JSON(http.StatusOK, gin.H{"seccess": true, "url": imageURL})
	})

	router.GET("/images", func(c *gin.Context) {
		listObjsResponse, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})

		if err != nil {
			log.Printf("Error listing objects: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list images"})
			return
		}

		var images []map[string]string
		for _, obj := range listObjsResponse.Contents {
			imageURL := fmt.Sprintf("http://localhost:9000/%s/%s", bucketName, *obj.Key)
			images = append(images, map[string]string{
				"name": *obj.Key,
				"url":  imageURL,
				"size": fmt.Sprintf("%.2f KB", float64(*obj.Size)/1024),
				"date": obj.LastModified.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, gin.H{"images": images})
	})

	router.GET("/view/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		imageUrl := fmt.Sprintf("http://localhost:9000/%s/%s", bucketName, filename)
		c.JSON(http.StatusOK, gin.H{"imageUrl": imageUrl, "filename": filename})
	})

	router.GET("/api/images/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(filename),
		})

		if err != nil {
			log.Printf("Error getting object: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		defer result.Body.Close()

		contentType := aws.ToString(result.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		c.DataFromReader(http.StatusOK, *result.ContentLength, contentType, result.Body, nil)
	})

	router.Run()
}
