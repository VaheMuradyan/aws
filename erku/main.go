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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	s3Client *s3.Client
	db       *gorm.DB
)

const defaultUploadDir = "uploads/"

// Image - նկարի մոդելը GORM-ի համար
type Image struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ObjectKey   string    `json:"objectKey" gorm:"type:varchar(255);not null"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Size        int64     `json:"size" gorm:"type:bigint;not null"`
	ContentType string    `json:"contentType" gorm:"type:varchar(100);not null"`
	UploadedAt  time.Time `json:"uploadedAt" gorm:"type:datetime;not null"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Tag - թեգի մոդելը GORM-ի համար
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ImageTag - Նկարի և թեգի կապը (many-to-many)
type ImageTag struct {
	ImageID   string    `json:"imageId" gorm:"type:varchar(36);primaryKey"`
	TagID     uint      `json:"tagId" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// ImageResponse - Պատկերի արձագանքի կառուցվածք API-ի համար
type ImageResponse struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	UploadedAt  time.Time `json:"uploadedAt"`
	Tags        []string  `json:"tags,omitempty"`
}

func main() {
	// Բեռնել .env ֆայլը
	err := godotenv.Load()
	if err != nil {
		log.Printf("Զգուշացում: .env ֆայլը բեռնելու սխալ: %v", err)
	}

	// Ստանալ MinIO կարգավորումները միջավայրի փոփոխականներից
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

	// MySQL + GORM կապակցում
	initDB()

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

	// CORS-ի պարզեցված կարգավորում
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
		// Նկարների հետ աշխատանքի մարշրուտներ
		images := api.Group("/images")
		{
			images.GET("", getAllImagesHandler)             // Բոլոր նկարների ցանկի ստացում
			images.GET("/:id", getImageByIdHandler)         // Կոնկրետ նկարի ստացում ID-ով
			images.GET("/tags/:tag", getImagesByTagHandler) // Նկարների ստացում թեգով
			images.POST("", uploadImageHandler)             // Նկարի վերբեռնում
			images.DELETE("/:id", deleteImageHandler)       // Նկարի ջնջում
			images.POST("/:id/tags", addTagsToImageHandler) // Նկարին թեգերի ավելացում
		}

		// Թեգերի հետ աշխատանքի մարշրուտներ
		api.GET("/tags", getAllTagsHandler) // Բոլոր թեգերի ցանկի ստացում
	}

	fmt.Println("API սերվերը գործարկվել է՝ :8080")
	router.Run(":8080")
}

// initDB - GORM-ով MySQL բազայի կապակցում և միգրացիա
func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnvWithDefault("DB_USER", "root"),
		getEnvWithDefault("DB_PASSWORD", "java"),
		getEnvWithDefault("DB_HOST", "localhost:3306"),
		getEnvWithDefault("DB_NAME", "minio_gallery"),
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Սխալ MySQL բազային կապակցվելիս: %v", err)
		return
	}

	// Մոդելների միգրացիա
	err = db.AutoMigrate(&Image{}, &Tag{}, &ImageTag{})
	if err != nil {
		log.Printf("Սխալ միգրացիայի ժամանակ: %v", err)
		return
	}

	log.Println("MySQL բազային կապակցումը հաստատված")
}

// getAllImagesHandler - Բոլոր նկարների ցանկի ստացում
func getAllImagesHandler(c *gin.Context) {
	var images []Image
	var result []ImageResponse

	// Ստանալ բոլոր նկարները բազայից
	if db != nil {
		if err := db.Find(&images).Error; err != nil {
			log.Printf("Սխալ նկարների ցանկը ստանալիս: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց ստանալ նկարների ցանկը"})
			return
		}

		// Ձևավորել պատասխանը
		for _, img := range images {
			response, err := createImageResponse(img)
			if err != nil {
				continue
			}
			result = append(result, response)
		}
	} else {
		// Եթե DB-ն հասանելի չէ, օգտագործել MinIO-ն ուղղակիորեն
		result = getImagesFromMinIO()
	}

	// Եթե արդյունքը դատարկ է, վերադարձնել դատարկ զանգված, ոչ թե null
	if result == nil {
		result = []ImageResponse{}
	}

	c.JSON(http.StatusOK, result)
}

// getImageByIdHandler - Կոնկրետ նկարի ստացում ID-ով
func getImageByIdHandler(c *gin.Context) {
	id := c.Param("id")

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Տվյալների բազան հասանելի չէ"})
		return
	}

	var image Image
	if err := db.First(&image, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Նկարը չի գտնվել"})
		return
	}

	// Ստանալ նկարի թեգերը
	var imageTags []ImageTag
	var tags []Tag
	var tagNames []string

	db.Where("image_id = ?", id).Find(&imageTags)

	for _, it := range imageTags {
		var tag Tag
		if err := db.First(&tag, it.TagID).Error; err == nil {
			tags = append(tags, tag)
			tagNames = append(tagNames, tag.Name)
		}
	}

	// Ստեղծել նախապես ստորագրված URL նկարի համար
	presignClient := s3.NewPresignClient(s3Client)
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")

	presignedURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(image.ObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց ստեղծել նկարի URL"})
		return
	}

	response := ImageResponse{
		ID:          image.ID,
		URL:         presignedURL.URL,
		Name:        image.Name,
		Size:        image.Size,
		ContentType: image.ContentType,
		UploadedAt:  image.UploadedAt,
		Tags:        tagNames,
	}

	c.JSON(http.StatusOK, response)
}

// getImagesByTagHandler - Նկարների ստացում թեգով
func getImagesByTagHandler(c *gin.Context) {
	tagName := c.Param("tag")

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Տվյալների բազան հասանելի չէ"})
		return
	}

	// Ստանալ թեգը
	var tag Tag
	if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Թեգը չի գտնվել"})
		return
	}

	// Ստանալ բոլոր նկարների ID-ները, որոնք ունեն այդ թեգը
	var imageTags []ImageTag
	db.Where("tag_id = ?", tag.ID).Find(&imageTags)

	var result []ImageResponse

	for _, it := range imageTags {
		var image Image
		if err := db.First(&image, "id = ?", it.ImageID).Error; err != nil {
			continue
		}

		response, err := createImageResponse(image)
		if err != nil {
			continue
		}

		result = append(result, response)
	}

	// Եթե արդյունքը դատարկ է, վերադարձնել դատարկ զանգված, ոչ թե null
	if result == nil {
		result = []ImageResponse{}
	}

	c.JSON(http.StatusOK, result)
}

// uploadImageHandler - Նկարի վերբեռնում
func uploadImageHandler(c *gin.Context) {
	// Ստանալ ֆայլը ձևից
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		log.Printf("Սխալ ձևի ֆայլը կարդալիս: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Չհաջողվեց կարդալ ֆայլը"})
		return
	}
	defer file.Close()

	// Ստանալ թեգերը, եթե տրված են
	tagsParam := c.PostForm("tags")
	var tagNames []string
	if tagsParam != "" {
		tagNames = strings.Split(tagsParam, ",")
		// Մաքրել սպիտակ տարածությունները
		for i := range tagNames {
			tagNames[i] = strings.TrimSpace(tagNames[i])
		}
	}

	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	uploadDir := getEnvWithDefault("UPLOAD_DIR", defaultUploadDir)

	// Ստեղծել յունիք ID նկարի համար
	imageID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(100000))

	// Կառուցել ֆայլի ուղին և վերբեռնել MinIO
	extension := filepath.Ext(header.Filename)
	objectKey := filepath.Join(uploadDir, imageID+extension)

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(getContentType(extension)),
	})

	if err != nil {
		log.Printf("Սխալ MinIO վերբեռնելիս: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց վերբեռնել նկարը"})
		return
	}

	// Եթե DB-ն հասանելի է, պահել նկարի մասին ինֆորմացիան
	if db != nil {
		// Ստեղծել նկարի գրառում
		image := Image{
			ID:          imageID,
			ObjectKey:   objectKey,
			Name:        header.Filename,
			Size:        header.Size,
			ContentType: getContentType(extension),
			UploadedAt:  time.Now(),
		}

		if err := db.Create(&image).Error; err != nil {
			log.Printf("Սխալ նկարը բազայում պահելիս: %v", err)
		}

		// Ավելացնել թեգերը, եթե տրված են
		for _, tagName := range tagNames {
			if tagName == "" {
				continue
			}

			// Ստուգել թեգի գոյությունը կամ ստեղծել նորը
			var tag Tag
			if err := db.Where("name = ?", tagName).FirstOrCreate(&tag, Tag{Name: tagName}).Error; err != nil {
				log.Printf("Սխալ թեգը ստուգելիս/ստեղծելիս: %v", err)
				continue
			}

			// Ստեղծել կապը նկարի և թեգի միջև
			imageTag := ImageTag{
				ImageID: imageID,
				TagID:   tag.ID,
			}

			if err := db.Create(&imageTag).Error; err != nil {
				log.Printf("Սխալ նկարին թեգ ավելացնելիս: %v", err)
			}
		}
	}

	// Վերադարձնել պատասխան React հավելվածին
	c.JSON(http.StatusOK, gin.H{
		"id":     imageID,
		"status": "success",
		"file":   header.Filename,
	})
}

// deleteImageHandler - Նկարի ջնջում
func deleteImageHandler(c *gin.Context) {
	id := c.Param("id")

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Տվյալների բազան հասանելի չէ"})
		return
	}

	// Ստանալ նկարի տվյալները նախքան ջնջելը
	var image Image
	if err := db.First(&image, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Նկարը չի գտնվել"})
		return
	}

	// Ջնջել նկարը MinIO-ից
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(image.ObjectKey),
	})

	if err != nil {
		log.Printf("Սխալ նկարը MinIO-ից ջնջելիս: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց ջնջել նկարը պահեստից"})
		return
	}

	// Ջնջել նկարի թեգերը
	if err := db.Where("image_id = ?", id).Delete(&ImageTag{}).Error; err != nil {
		log.Printf("Սխալ նկարի թեգերը ջնջելիս: %v", err)
	}

	// Ջնջել նկարի գրառումը բազայից
	if err := db.Delete(&image).Error; err != nil {
		log.Printf("Սխալ նկարը բազայից ջնջելիս: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց ջնջել նկարի մասին տվյալները"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Նկարը հաջողությամբ ջնջվել է",
	})
}

// addTagsToImageHandler - Նկարին թեգերի ավելացում
func addTagsToImageHandler(c *gin.Context) {
	id := c.Param("id")

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Տվյալների բազան հասանելի չէ"})
		return
	}

	// Ստուգել նկարի գոյությունը
	var image Image
	if err := db.First(&image, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Նկարը չի գտնվել"})
		return
	}

	// Ստանալ նոր թեգերը
	var input struct {
		Tags []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || len(input.Tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Անվավեր կամ դատարկ թեգերի ցանկ"})
		return
	}

	// Ավելացնել յուրաքանչյուր թեգը
	var addedTags []string

	for _, tagName := range input.Tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		// Ստուգել թեգի գոյությունը կամ ստեղծել նորը
		var tag Tag
		if err := db.Where("name = ?", tagName).FirstOrCreate(&tag, Tag{Name: tagName}).Error; err != nil {
			log.Printf("Սխալ թեգը ստուգելիս/ստեղծելիս: %v", err)
			continue
		}

		// Ստուգել, թե արդյոք այս կապը արդեն գոյություն ունի
		var existingImageTag ImageTag
		result := db.Where("image_id = ? AND tag_id = ?", id, tag.ID).First(&existingImageTag)

		if result.RowsAffected == 0 {
			// Ստեղծել կապը նկարի և թեգի միջև
			imageTag := ImageTag{
				ImageID: id,
				TagID:   tag.ID,
			}

			if err := db.Create(&imageTag).Error; err != nil {
				log.Printf("Սխալ նկարին թեգ ավելացնելիս: %v", err)
				continue
			}

			addedTags = append(addedTags, tagName)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"addedTags": addedTags,
	})
}

// getAllTagsHandler - Բոլոր թեգերի ցանկի ստացում
func getAllTagsHandler(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Տվյալների բազան հասանելի չէ"})
		return
	}

	var tags []Tag
	if err := db.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Չհաջողվեց ստանալ թեգերի ցանկը"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// getImagesFromMinIO - Ստանալ նկարները ուղղակիորեն MinIO-ից (առանց բազայի)
func getImagesFromMinIO() []ImageResponse {
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")
	uploadDir := getEnvWithDefault("UPLOAD_DIR", defaultUploadDir)

	// Ստանալ bucket-ում առկա օբյեկտների ցանկը
	resp, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(uploadDir),
	})

	if err != nil {
		log.Printf("Սխալ MinIO-ից օբյեկտները թվարկելիս: %v", err)
		return nil
	}

	// Ստեղծել նախապես ստորագրված URL-ներ յուրաքանչյուր օբյեկտի համար
	var images []ImageResponse

	if resp.Contents == nil || len(resp.Contents) == 0 {
		return images
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

		// Տվյալների կազմավորում (փորձում ենք ID-ն ստանալ ֆայլի անունից, եթե ֆորմատը համապատասխանում է)
		id := strings.TrimSuffix(filename, filepath.Ext(filename))

		images = append(images, ImageResponse{
			ID:          id,
			URL:         presignedURL.URL,
			Name:        filename,
			Size:        *item.Size,
			ContentType: getContentType(filepath.Ext(filename)),
			UploadedAt:  *item.LastModified,
		})
	}

	return images
}

// createImageResponse - Ստեղծել ImageResponse օբյեկտ նկարի մոդելից
func createImageResponse(image Image) (ImageResponse, error) {
	// Ստեղծել նախապես ստորագրված URL նկարի համար
	presignClient := s3.NewPresignClient(s3Client)
	bucketName := getEnvWithDefault("BUCKET_NAME", "images")

	presignedURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(image.ObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour
	})

	if err != nil {
		return ImageResponse{}, err
	}

	// Ստանալ նկարի թեգերը
	var tags []string

	if db != nil {
		var imageTags []ImageTag
		db.Where("image_id = ?", image.ID).Find(&imageTags)

		for _, it := range imageTags {
			var tag Tag
			if err := db.First(&tag, it.TagID).Error; err == nil {
				tags = append(tags, tag.Name)
			}
		}
	}

	return ImageResponse{
		ID:          image.ID,
		URL:         presignedURL.URL,
		Name:        image.Name,
		Size:        image.Size,
		ContentType: image.ContentType,
		UploadedAt:  image.UploadedAt,
		Tags:        tags,
	}, nil
}

// getContentType - Ստանալ MIME տեսակը ֆայլի ընդլայնումից
// getContentType - Ստանալ MIME տեսակը ֆայլի ընդլայնումից
func getContentType(extension string) string {
	extension = strings.ToLower(extension)

	switch extension {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
