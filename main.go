package main

import (
	"a21hc3NpZ25tZW50/model"
	"a21hc3NpZ25tZW50/repository/authRepository"
	dbRepository "a21hc3NpZ25tZW50/repository/dbRepository"
	"a21hc3NpZ25tZW50/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Credential struct {
	Host         string
	Username     string
	Password     string
	DatabaseName string
	Port         int
	Schema       string
}

func AuthMiddleware(authRepo *authRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authRepo.IsLoggedIn() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func Connect(creds *Credential) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Jakarta",
		creds.Host, creds.Username, creds.Password, creds.DatabaseName, creds.Port)

	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func SetupRouter(dbRepo *dbRepository.Repository, authRepo *authRepository.Repository) *gin.Engine {
	svc := service.NewService(*dbRepo, authRepo)
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.POST("/users", func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := svc.Register(user)
		if err != nil {
			if err.Error() == "username already registered" {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "Created", "User": user})
	})

	router.POST("/signin", func(c *gin.Context) {
		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&credentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := dbRepo.GetUserByUsername(credentials.Username)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
			return
		}

		if user.Username == "" || user.Password != credentials.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		err = svc.Login(credentials.Username, credentials.Password)
		if err != nil {
			log.Printf("Login error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "Logged in", "User": user})
	})

	// router.GET("/test", func(c *gin.Context) {
	// 	var request struct {
	// 		Username string `json:"username"`
	// 	}
	// 	if err := c.ShouldBindJSON(&request); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	user, err := dbRepo.GetUserByUsername(request.Username)
	// 	if err != nil {
	// 		log.Printf("Error fetching user: %v", err)
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
	// 		return
	// 	}

	// 	if user.Username == "" {
	// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 		return
	// 	}

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"username": user.Username,
	// 		"password": user.Password,
	// 		"fullname": user.Fullname,
	// 	})
	// })

	protected := router.Group("/")
	protected.Use(AuthMiddleware(authRepo))
	{
		protected.GET("/photos", func(c *gin.Context) {
			username := authRepo.LoggedInUser.Username
			photos, err := dbRepo.GetAllPhotosByUser(username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, photos)
		})

		protected.GET("/photos/:id", func(c *gin.Context) {
			id := c.Param("id")
			photoID := 0
			fmt.Sscanf(id, "%d", &photoID)
			photo, err := dbRepo.GetPhotoByID(uint(photoID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
				return
			}
			c.JSON(http.StatusOK, photo)
		})

		protected.POST("/photos", func(c *gin.Context) {
			var photo model.Photo
			if err := c.ShouldBindJSON(&photo); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			user, err := dbRepo.GetUserByUsername(authRepo.LoggedInUser.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
				return
			}

			photo.UserID = user.ID
			photoID, err := dbRepo.AddPhoto(photo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"photo_id": photoID})
		})

		protected.DELETE("/photos/:id", func(c *gin.Context) {
			id := c.Param("id")
			photoID := 0
			fmt.Sscanf(id, "%d", &photoID)
			err := dbRepo.DeletePhoto(uint(photoID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "Photo deleted"})
		})
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})

	return router
}

func main() {
	dbCredential := Credential{
		Host:         "localhost",
		Username:     "postgres",
		Password:     "tanggal9bulan5",
		DatabaseName: "kampusmerdeka",
		Port:         5432,
	}

	dbConn, err := Connect(&dbCredential)
	if err != nil {
		log.Fatal(err)
	}

	if err = dbConn.Migrator().DropTable("users", "photos"); err != nil {
		log.Fatal("failed dropping table:" + err.Error())
	}

	if err = dbConn.AutoMigrate(&model.User{}, &model.Photo{}); err != nil {
		log.Fatal("failed migrating table:" + err.Error())
	}

	authRepo := authRepository.NewRepository()
	dbRepo := dbRepository.NewRepository(dbConn)
	router := SetupRouter(dbRepo, authRepo)
	router.Run()
}
