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

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

		if err = db.Migrator().DropTable("users", "photos"); err != nil {
			panic("failed droping table:" + err.Error())
		}

		err = db.AutoMigrate(&model.User{}, &model.Photo{})
		if err != nil {
			panic("failed migrating table:" + err.Error())
		}

		dbRepo = dbRepository.NewRepository(db)
		authRepo = authRepository.NewRepository()

		_, err = main.Connect(&dbCredential)
		if err != nil {
			panic("failed connecting to database, please check Connect function")
		}

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

		photo := model.Photo{
			UserID:    1,
			URL:       "https://www.google.com",
			Caption:   "Caption",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = db.Create(&photo).Error

		if err != nil {
			panic("failed creating photo")
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
			Expect(response["User"]).NotTo(BeNil())
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

	When("GET /photos", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			req, _ := http.NewRequest(http.MethodGet, "/photos", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should return photos if user is logged in", func() {
			// Log in the user first
			authRepo.Login("user")

			req, _ := http.NewRequest(http.MethodGet, "/photos", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
		})
	})

	When("POST /photos", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			photo := model.Photo{
				UserID:  1,
				URL:     "https://example.com/photo.jpg",
				Caption: "A beautiful sunset",
			}
			body, _ := json.Marshal(photo)
			req, _ := http.NewRequest(http.MethodPost, "/photos", bytes.NewBuffer(body))
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should add a new photo if user is logged in", func() {
			// Log in the user first
			authRepo.Login("user")

			photo := model.Photo{
				UserID:  1,
				URL:     "https://example.com/photo.jpg",
				Caption: "A beautiful sunset",
			}
			body, _ := json.Marshal(photo)
			req, _ := http.NewRequest(http.MethodPost, "/photos", bytes.NewBuffer(body))
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			Expect(err).To(BeNil())

			photoID, ok := response["photo_id"].(float64)
			Expect(ok).To(BeTrue())
			Expect(photoID).To(BeNumerically(">", 0))

			addedPhoto, err := dbRepo.GetPhotoByID(uint(photoID))
			Expect(err).To(BeNil())
			Expect(addedPhoto.Caption).To(Equal("A beautiful sunset"))
		})
	})

	When("GET /photos/:id", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			req, _ := http.NewRequest(http.MethodGet, "/photos/1", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should get a photo by ID if user is logged in", func() {
			authRepo.Login("user")

			photo := model.Photo{
				UserID:    1,
				URL:       "https://example.com/photo2.jpg",
				Caption:   "A second photo",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			photoID, err := dbRepo.AddPhoto(photo)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/photos/%d", photoID), nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var retrievedPhoto model.Photo
			err = json.Unmarshal(resp.Body.Bytes(), &retrievedPhoto)
			Expect(err).To(BeNil())
			Expect(retrievedPhoto.ID).To(Equal(photoID))
			Expect(retrievedPhoto.Caption).To(Equal("A second photo"))
		})
	})

	When("DELETE /photos/:id", func() {
		It("should return 401 Unauthorized if user is not logged in", func() {
			req, _ := http.NewRequest(http.MethodDelete, "/photos/1", nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnauthorized))
			Expect(resp.Body.String()).To(ContainSubstring("Unauthorized"))
		})

		It("should delete a photo by ID if user is logged in", func() {
			authRepo.Login("user")

			photo := model.Photo{
				UserID:    1,
				URL:       "https://example.com/photo3.jpg",
				Caption:   "A third photo",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			photoID, err := dbRepo.AddPhoto(photo)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/photos/%d", photoID), nil)
			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(ContainSubstring("Photo deleted"))

			deletedPhoto, err := dbRepo.GetPhotoByID(photoID)
			Expect(err).To(BeNil())
			Expect(deletedPhoto.ID).To(Equal(uint(0))) // Expect ID to be 0 since the photo should be deleted
		})
	})
})
