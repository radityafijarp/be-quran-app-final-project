package main_test

import (
	main "a21hc3NpZ25tZW50"
	"a21hc3NpZ25tZW50/model"
	"a21hc3NpZ25tZW50/repository/authRepository"
	dbRepository "a21hc3NpZ25tZW50/repository/dbRepository"
	"bytes"
	"encoding/json"

	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func generateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString("helloWorld") // Use your JWT secret here
}

var (
	resp     *httptest.ResponseRecorder
	router   *gin.Engine
	dbRepo   *dbRepository.Repository
	authRepo *authRepository.Repository
)

var _ = Describe("Main", Ordered, func() {
	dbCredential := main.Credential{
		Host:         "localhost",
		Username:     "postgres",
		Password:     "tanggal9bulan5",
		DatabaseName: "kampusmerdeka",
		Port:         5432,
		Schema:       "public",
	}

	BeforeAll(func() {
		db, err := main.Connect(&dbCredential)
		if err != nil {
			panic("failed connecting to database, please check Connect credentials")
		}

		// Drop tables in reverse order of their dependencies
		if err = db.Migrator().DropTable("memorizes", "users"); err != nil {
			panic("failed dropping tables:" + err.Error())
		}

		err = db.AutoMigrate(&model.User{}, &model.Memorize{})
		if err != nil {
			panic("failed migrating tables:" + err.Error())
		}

		dbRepo = dbRepository.NewRepository(db)
		authRepo = authRepository.NewRepository()

		// Insert test data
		user := model.User{
			Username:   "user",
			Password:   "password",
			Fullname:   "User",
			Desc:       "User User",
			ProfilePic: "https://www.google.com",
		}

		err = db.Create(&user).Error
		if err != nil {
			panic("failed creating user")
		}

		memorize := model.Memorize{
			UserID:          user.ID, // Use the ID of the created user
			SurahName:       "Al-Fatiha",
			AyahRange:       "1-7",
			TotalAyah:       7,
			DateStarted:     time.Now(),
			DateCompleted:   time.Time{}, // Empty if not completed
			ReviewFrequency: "Weekly",
			LastReviewDate:  time.Now(),
			AccuracyLevel:   "High",
			NextReviewDate:  time.Now().AddDate(0, 0, 7), // Next review in a week
			Notes:           "Focused on tajweed",
		}

		err = db.Create(&memorize).Error
		if err != nil {
			panic("failed creating memorize record")
		}
	})

	BeforeEach(func() {
		router = main.SetupRouter(dbRepo, authRepo)
		resp = httptest.NewRecorder()
		authRepo.Logout()
	})

	When("GET /health", func() {
		It("should return 200 OK", func() {
			req, _ := http.NewRequest(http.MethodGet, "/health", nil)
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(Equal("OK"))
		})
	})

	When("POST /users", func() {
		It("should register new user", func() {

			user := model.User{
				Username:   "aditira",
				Password:   "password",
				Desc:       "Nama saya Aditira",
				Fullname:   "Aditira Jamhuri",
				ProfilePic: "https://google.com",
			}

			body, _ := json.Marshal(user)
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))

			router.ServeHTTP(resp, req)

			Expect(resp.Result().StatusCode).To(Equal(http.StatusCreated))

			u, err := dbRepo.GetUserByUsername("aditira")

			Expect(err).To(BeNil())
			Expect(u.Fullname).To(Equal("Aditira Jamhuri"))
		})

		It("should reject if user already exist", func() {
			existingUser := model.User{
				Username:   "eddy",
				Password:   "password",
				Desc:       "Eddy",
				Fullname:   "Eddy Permana",
				ProfilePic: "https://google.com",
			}

			_, err := dbRepo.AddUser(existingUser)
			Expect(err).To(BeNil())

			body, _ := json.Marshal(existingUser)
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusConflict))
			Expect(resp.Body.String()).To(ContainSubstring("username already registered"))
		})
	})

	When("POST /signin", func() {
		It("should sign in user successfully", func() {
			signInData := map[string]string{
				"username": "user",
				"password": "password",
			}
			body, _ := json.Marshal(signInData)

			req, _ := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var response map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			Expect(err).To(BeNil())

			Expect(response["status"]).To(Equal("Logged in"))
			Expect(response["token"]).NotTo(BeNil()) // Check that a JWT token is returned
		})

		It("should return 401 Unauthorized with invalid credentials", func() {
			signInData := map[string]string{
				"username": "user",
				"password": "wrongpassword",
			}
			body, _ := json.Marshal(signInData)

			req, _ := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Invalid username or password"))
		})
	})

	When("GET /memorizes", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			req, _ := http.NewRequest(http.MethodGet, "/memorizes", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should return memorization records if user is logged in", func() {
			// Generate a JWT token for the logged-in user
			token, _ := generateJWT("user")

			req, _ := http.NewRequest(http.MethodGet, "/memorizes", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			// Add further checks to validate the returned data if necessary
		})
	})

	When("POST /memorizes", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			memorize := model.Memorize{
				SurahName:       "Al-Baqarah",
				AyahRange:       "1-10",
				TotalAyah:       10,
				DateStarted:     time.Now(),
				ReviewFrequency: "Daily",
				Notes:           "Initial review",
			}
			body, _ := json.Marshal(memorize)
			req, _ := http.NewRequest(http.MethodPost, "/memorizes", bytes.NewBuffer(body))
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should add a new memorization record if user is logged in", func() {
			// Generate a JWT token for the logged-in user
			token, _ := generateJWT("user")

			memorize := model.Memorize{
				SurahName:       "Al-Baqarah",
				AyahRange:       "1-10",
				TotalAyah:       10,
				DateStarted:     time.Now(),
				ReviewFrequency: "Daily",
				Notes:           "Initial review",
			}
			body, _ := json.Marshal(memorize)
			req, _ := http.NewRequest(http.MethodPost, "/memorizes", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			Expect(err).To(BeNil())

			memorizeID, ok := response["memorize_id"].(float64)
			Expect(ok).To(BeTrue())
			Expect(memorizeID).To(BeNumerically(">", 0))

			addedMemorize, err := dbRepo.GetMemorizeByID(uint(memorizeID))
			Expect(err).To(BeNil())
			Expect(addedMemorize.SurahName).To(Equal("Al-Baqarah"))
			Expect(addedMemorize.AyahRange).To(Equal("1-10"))
			Expect(addedMemorize.TotalAyah).To(Equal(10))
			Expect(addedMemorize.DateStarted).To(BeTemporally("~", time.Now(), time.Minute))
			Expect(addedMemorize.ReviewFrequency).To(Equal("Daily"))
			Expect(addedMemorize.Notes).To(Equal("Initial review"))
			Expect(addedMemorize.DateCompleted.IsZero()).To(BeTrue())
		})
	})

	When("GET /memorizes/:id", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			req, _ := http.NewRequest(http.MethodGet, "/memorizes/1", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should get a memorization record by ID if user is logged in", func() {
			token, _ := generateJWT("user")

			memorize := model.Memorize{
				SurahName:       "Al-Baqarah",
				AyahRange:       "1-10",
				TotalAyah:       10,
				DateStarted:     time.Now(),
				ReviewFrequency: "Daily",
				Notes:           "Initial review",
			}

			memorizeID, err := dbRepo.AddMemorize(memorize)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/memorizes/%d", memorizeID), nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var fetchedMemorize model.Memorize
			err = json.Unmarshal(resp.Body.Bytes(), &fetchedMemorize)
			Expect(err).To(BeNil())
			Expect(fetchedMemorize.SurahName).To(Equal("Al-Baqarah"))
			Expect(fetchedMemorize.AyahRange).To(Equal("1-10"))
			Expect(fetchedMemorize.TotalAyah).To(Equal(10))
			Expect(fetchedMemorize.DateStarted).To(BeTemporally("~", time.Now(), time.Minute))
			Expect(fetchedMemorize.ReviewFrequency).To(Equal("Daily"))
			Expect(fetchedMemorize.Notes).To(Equal("Initial review"))
			Expect(fetchedMemorize.DateCompleted.IsZero()).To(BeTrue())
		})
	})

	// When("GET /memorizes/:id", func() {
	// 	It("should return 401 Unauthorized if user is not logged in", func() {
	// 		req, _ := http.NewRequest(http.MethodGet, "/memorizes/1", nil)
	// 		router.ServeHTTP(resp, req)

	// 		Expect(resp.Code).To(Equal(http.StatusUnauthorized))
	// 		Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
	// 	})

	// 	It("should get a memorization record by ID if user is logged in", func() {
	// 		authRepo.Login("user")

	// 		memorize := model.Memorize{
	// 			SurahName:       "Al-Baqarah",
	// 			AyahRange:       "1-10",
	// 			TotalAyah:       10,
	// 			DateStarted:     time.Now(),
	// 			ReviewFrequency: "Daily",
	// 			Notes:           "Initial review",
	// 		}

	// 		memorizeID, err := dbRepo.AddMemorize(memorize)
	// 		Expect(err).To(BeNil())

	// 		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/memorizes/%d", memorizeID), nil)
	// 		router.ServeHTTP(resp, req)

	// 		Expect(resp.Code).To(Equal(http.StatusOK))

	// 		var fetchedMemorize model.Memorize
	// 		err = json.Unmarshal(resp.Body.Bytes(), &fetchedMemorize)
	// 		Expect(err).To(BeNil())
	// 		Expect(fetchedMemorize.SurahName).To(Equal("Al-Baqarah"))
	// 		Expect(fetchedMemorize.AyahRange).To(Equal("1-10"))
	// 		Expect(fetchedMemorize.TotalAyah).To(Equal(10))
	// 		Expect(fetchedMemorize.DateStarted).To(BeTemporally("~", time.Now(), time.Minute))
	// 		Expect(fetchedMemorize.ReviewFrequency).To(Equal("Daily"))
	// 		Expect(fetchedMemorize.Notes).To(Equal("Initial review"))
	// 		Expect(fetchedMemorize.DateCompleted.IsZero()).To(BeTrue())
	// 	})
	// })

	// When("DELETE /memorizes/:id", func() {
	// 	It("should return 401 Unauthorized if user is not logged in", func() {
	// 		req, _ := http.NewRequest(http.MethodDelete, "/memorizes/1", nil)
	// 		router.ServeHTTP(resp, req)

	// 		Expect(resp.Code).To(Equal(http.StatusUnauthorized))
	// 		Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
	// 	})

	// 	It("should delete a memorization record if user is logged in", func() {
	// 		authRepo.Login("user")

	// 		memorize := model.Memorize{
	// 			SurahName:       "Al-Baqarah",
	// 			AyahRange:       "1-10",
	// 			TotalAyah:       10,
	// 			DateStarted:     time.Now(),
	// 			ReviewFrequency: "Daily",
	// 			Notes:           "Initial review",
	// 		}

	// 		memorizeID, err := dbRepo.AddMemorize(memorize)
	// 		Expect(err).To(BeNil())

	// 		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/memorizes/%d", memorizeID), nil)
	// 		router.ServeHTTP(resp, req)

	// 		Expect(resp.Code).To(Equal(http.StatusOK))

	// 		_, err = dbRepo.GetMemorizeByID(memorizeID)
	// 		Expect(err).To(Equal(fmt.Errorf("memorize record not found")))
	// 	})
	// })
})
