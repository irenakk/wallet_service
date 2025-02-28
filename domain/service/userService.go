package service

import (
	"golang.org/x/crypto/bcrypt"
	"wallet-service/domain/entity"
)

type UserRepository interface {
	UserRegistration(user entity.User) error
	GetUserHashedPass(user entity.User) (string, error)
	GetUserID(user entity.User) (int, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Registration(user entity.User) error {
	hashedPassword, err := hashPassword(user.Password)
	user.Password = hashedPassword
	err = s.repo.UserRegistration(user)
	if err != nil {
		return err
	}
	return err
}

func (s *UserService) Authorization(user entity.User) error {
	userID, err := s.repo.GetUserID(user)
	user.ID = userID

	var hashedPassword string
	hashedPassword, err = s.repo.GetUserHashedPass(user)
	if err != nil {
		return err
	}
	if checkPasswordHash(user.Password, hashedPassword) != nil {
		return err
	}
	return err
}

func (s *UserService) GetUserID(user entity.User) (int, error) {
	userID, err := s.repo.GetUserID(user)
	return userID, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
