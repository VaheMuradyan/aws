package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

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
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Faile to upload image"})
			return
		}

		result, uploadError := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("code"),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		})

		if uploadError != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{"error": "Faile to upload image"})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{"image": result.Location})
	})
	router.Run()
}
