package service

import (
	"a21hc3NpZ25tZW50/model"
	"a21hc3NpZ25tZW50/repository/authRepository"
	dbRepository "a21hc3NpZ25tZW50/repository/dbRepository"
	"errors"
	"log"
	"reflect"
)

type Service struct {
	repository     dbRepository.Repository
	authRepository *authRepository.Repository
	UserLogin      model.User
}

func NewService(repo dbRepository.Repository, auth *authRepository.Repository) *Service {
	return &Service{repository: repo, UserLogin: model.User{}, authRepository: auth}
}

func IsEmptyUser(user model.User) bool {
	return reflect.DeepEqual(user, model.User{})
}

func (s *Service) Register(user model.User) error {
	if !IsEmptyUser(s.UserLogin) {
		return errors.New("user already login")
	}

	userDB, err := s.repository.GetUserByUsername(user.Username)
	if err != nil {
		return err
	}

	if userDB.Username != "" && userDB.Username == user.Username {
		return errors.New("username already registered")
	}

	s.repository.AddUser(user)

	return nil
}

func (s *Service) Login(username string, password string) error {
	log.Printf("Checking if user is already logged in: %v", s.authRepository.IsLoggedIn())
	if s.authRepository.IsLoggedIn() {
		return errors.New("user already login")
	}

	user, err := s.repository.GetUserByUsername(username)
	if err != nil {
		return err
	}

	log.Printf("Fetched user: %+v", user)
	if IsEmptyUser(user) {
		return errors.New("username or password is wrong")
	}

	if user.Password != password {
		return errors.New("username or password is wrong")
	}

	s.authRepository.Login(username)
	return nil
}
