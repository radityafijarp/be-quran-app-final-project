package main

import (
	"a21hc3NpZ25tZW50/model"
	"a21hc3NpZ25tZW50/repository/authRepository"
	dbRepository "a21hc3NpZ25tZW50/repository/dbRepository"
	"a21hc3NpZ25tZW50/service"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
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

var jwtSecret = []byte("helloWorld") // secret key

func AuthMiddleware(authRepo *authRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["username"] == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("username", claims["username"])
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

	// Enable CORS for all origins, methods, and headers
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

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

		// Create JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expiration
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "Logged in",
			"token":  tokenString,
		})
	})

	protected := router.Group("/")
	protected.Use(AuthMiddleware(authRepo))
	{
		protected.GET("/memorizes", func(c *gin.Context) {
			// Retrieve the username from the context
			username := c.GetString("username")

			// Fetch memorizes for the user
			memorizes, err := dbRepo.GetAllMemorizesByUser(username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, memorizes)
		})

		protected.GET("/memorizes/:id", func(c *gin.Context) {
			id := c.Param("id")
			memorizeID := 0
			fmt.Sscanf(id, "%d", &memorizeID)
			memorize, err := dbRepo.GetMemorizeByID(uint(memorizeID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Memorize record not found"})
				return
			}
			c.JSON(http.StatusOK, memorize)
		})

		protected.POST("/memorizes", func(c *gin.Context) {
			var memorize model.Memorize

			// Bind the JSON request body to the `memorize` struct
			if err := c.ShouldBindJSON(&memorize); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Get the logged-in username from the context
			username := c.GetString("username")

			// Retrieve the user by username from the database
			user, err := dbRepo.GetUserByUsername(username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
				return
			}

			// Set the UserID field for the memorize record
			memorize.UserID = user.ID

			// Add the memorize record to the database
			memorizeID, err := dbRepo.AddMemorize(memorize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Return the created memorize record ID
			c.JSON(http.StatusCreated, gin.H{"memorize_id": memorizeID})
		})

		protected.DELETE("/memorizes/:id", func(c *gin.Context) {
			id := c.Param("id")
			memorizeID := 0
			fmt.Sscanf(id, "%d", &memorizeID)
			err := dbRepo.DeleteMemorize(uint(memorizeID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Memorize record not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "Memorize record deleted"})
		})

		protected.PUT("/memorizes/:id", func(c *gin.Context) {
			id := c.Param("id")
			memorizeID := 0
			fmt.Sscanf(id, "%d", &memorizeID)

			// Retrieve the existing memorize record by ID
			existingMemorize, err := dbRepo.GetMemorizeByID(uint(memorizeID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Memorize record not found"})
				return
			}

			// Bind the JSON request body to the memorize struct
			var updatedMemorize model.Memorize
			if err := c.ShouldBindJSON(&updatedMemorize); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Update the existing memorize fields with the new data
			existingMemorize.SurahName = updatedMemorize.SurahName
			existingMemorize.AyahRange = updatedMemorize.AyahRange
			existingMemorize.TotalAyah = updatedMemorize.TotalAyah
			existingMemorize.DateStarted = updatedMemorize.DateStarted
			existingMemorize.DateCompleted = updatedMemorize.DateCompleted
			existingMemorize.ReviewFrequency = updatedMemorize.ReviewFrequency
			existingMemorize.LastReviewDate = updatedMemorize.LastReviewDate
			existingMemorize.AccuracyLevel = updatedMemorize.AccuracyLevel
			existingMemorize.NextReviewDate = updatedMemorize.NextReviewDate
			existingMemorize.Notes = updatedMemorize.Notes

			// Save the updated memorize record to the database
			err = dbRepo.UpdateMemorize(existingMemorize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memorize record"})
				return
			}

			c.JSON(http.StatusOK, existingMemorize)
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

	if err = dbConn.Migrator().DropTable("users", "memorizes"); err != nil {
		log.Fatal("failed dropping table:" + err.Error())
	}

	if err = dbConn.AutoMigrate(&model.User{}, &model.Memorize{}); err != nil {
		log.Fatal("failed migrating table:" + err.Error())
	}

	authRepo := authRepository.NewRepository()
	dbRepo := dbRepository.NewRepository(dbConn)
	router := SetupRouter(dbRepo, authRepo)
	router.Run()
}
