package repository

import (
	"back_music/internal/database"
	"back_music/internal/models"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	FindUserByEmail(email string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepo{db: database.DB}
}

func (r *userRepo) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User tidak ditemukan, bukan error
		}
		return nil, err // Error database lainnya
	}

	return &user, nil
}

func (r *userRepo) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Likes").Preload("Plays").First(&user, id).Error
	return &user, err
}

func (r *userRepo) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (r *userRepo) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
