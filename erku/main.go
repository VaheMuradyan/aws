package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	s3Client *s3.Client
)

const defaultUploadDir = "uploads/"

// ImageResponse - Պատկերի արձագանքի կառուցվածք
type ImageResponse struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func main() {
	// Բեռնել .env ֆայլը
	err := godotenv.Load()
	if err != nil {
		log.Printf("Զգուշացում: .env ֆայլը բեռնելու սխալ: %v", err)
	}

	// Ստանալ կարգավորումները միջավայրի փոփոխականներից
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpointURL := getEnvWithDefault("MINIO_ENDPOINT", "http://localhost:9000")

	// Ստուգել պարտադիր պարամետրերը
	if accessKey == "" || secretKey == "" {
		log.Fatal("Սխալ: MinIO access key և secret key պարտադիր են: Կարգավորեք AWS_ACCESS_KEY_ID և AWS_SECRET_ACCESS_KEY .env ֆայլում:")
	}

	// Կարգավորել MinIO հաճախորդին
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpointURL,
			HostnameImmutable: true,
			SigningRegion:     "us-east-1",
		}, nil
	})

	cfg := aws.Config{
		Credentials:                 credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		EndpointResolverWithOptions: customResolver,
		Region:                      "us-east-1",
		RetryMaxAttempts:            3,
	}

	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Կարևոր է MinIO համատեղելիության համար
	})

	// Ստուգել արդյոք bucket-ը գոյություն ունի, ստեղծել եթե չկա
	_, err = s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Printf("Bucket %s-ը գոյություն չունի, ստեղծվում է...", bucketName)
		_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("Չհաջողվեց ստեղծել bucket: %v", err)
		}
	}

	// Կարգավորել Gin ռոութերը
	router := gin.Default()

	// CORS-ի կարգավորում React հավելվածի հետ աշխատելու համար
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API մարշրուտների կարգավորում
	api := router.Group("/api")
	{
		api.GET("/images", getImagesHandler)
		api.POST("/upload", uploadHandler)
		// Ավելացնենք թեստային մարշրուտ
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "API աշխատում է",
				"status":  "success",
			})
		})
	}

	fmt.Println("API սերվերը գործարկվել է՝ :8080")
	router.Run(":8080")
}

// getImagesHandler - Պատկերների ցանկի ստացման մշակիչ
func getImagesHandler(c *gin.Context) {
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	uploadDir := getEnvWithDefault("UPLOAD_DIR", defaultUploadDir)

	// Ստանալ bucket-ում առկա օբյեկտների ցանկը
	resp, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(uploadDir),
	})

	// Լրացուցիչ լոգեր
	log.Printf("Բաքետ: %s, Պանակ: %s", bucketName, uploadDir)

	if err != nil {
		log.Printf("Սխալ օբյեկտները թվարկելիս: %v", err)
		// Սխալի դեպքում դատարկ զանգված վերադարձնել (ոչ թե null)
		c.JSON(http.StatusInternalServerError, []ImageResponse{})
		return
	}

	// Ստեղծել նախապես ստորագրված URL-ներ յուրաքանչյուր օբյեկտի համար
	var images []ImageResponse

	if resp.Contents == nil || len(resp.Contents) == 0 {
		log.Printf("Բաքետում պատկերներ չեն գտնվել")
		c.JSON(http.StatusOK, []ImageResponse{})
		return
	}

	for _, item := range resp.Contents {
		// Բացառել պանակները
		if *item.Key == uploadDir {
			continue
		}

		presignClient := s3.NewPresignClient(s3Client)
		presignedURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    item.Key,
		}, func(opts *s3.PresignOptions) {
			opts.Expires = 1 * time.Hour
		})

		if err != nil {
			log.Printf("Սխալ նախապես ստորագրված URL ստեղծելիս: %v", err)
			continue
		}

		// Ստանալ ֆայլի անունը ուղուց
		filename := filepath.Base(*item.Key)

		images = append(images, ImageResponse{
			URL:  presignedURL.URL,
			Name: filename,
		})
	}

	log.Printf("Վերադարձվում է %d պատկերներ", len(images))
	c.JSON(http.StatusOK, images)
}

// uploadHandler - Նկարների վերբեռնման մշակիչ
func uploadHandler(c *gin.Context) {
	// Ստանալ ֆայլը ձևից
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		log.Printf("Սխալ ձևի ֆայլը կարդալիս: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Չհաջողվեց կարդալ ֆայլը"})
		return
	}
	defer file.Close()

	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	uploadDir := getEnvWithDefault("UPLOAD_DIR", defaultUploadDir)

	// Կառուցել ֆայլի ուղին և վերբեռնել MinIO
	s3Key := filepath.Join(uploadDir, header.Filename)
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	if err != nil {
		log.Printf("Սխալ MinIO վերբեռնելիս: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց վերբեռնել MinIO"})
		return
	}

	// Վերադարձնել պատասխան React հավելվածին
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"file":   header.Filename,
	})
}

// getEnvWithDefault - Օգնող ֆունկցիա միջավայրի փոփոխականը լռելյայն արժեքով ստանալու համար
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
