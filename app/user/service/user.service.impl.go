package user

import (
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	v "finance/app/user/validation"
	m "finance/model"
)

type UserService struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewUserService(db *gorm.DB, v *validator.Validate) IUserService {
	return &UserService{
		DB:       db,
		Validate: v,
	}
}

func (s UserService) UserRegistartion(req v.UserRegistrationRequest) (v.UserResponse, error) {
	var err error
	var resp v.UserResponse
	// Validate User Request
	if req.Provider == "MANUAL" {
		err = s.Validate.Struct(v.ManualRegistrationRequest{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
			Fullname: req.Fullname,
			Provider: "MANUAL",
		})
	}

	if req.Provider == "GOOGLE" {
		err = s.Validate.Struct(v.GoogleRegistrationRequest{
			Username:   req.Username,
			Password:   req.Password,
			Email:      req.Email,
			Fullname:   req.Fullname,
			Provider:   "GOOGLE",
			ProviderId: req.ProviderId,
		})
	}

	if err != nil {
		return resp, err
	}

	// Cek Existing User
	var existingUser int64
	s.DB.Model(&m.User{}).Where("username = ?", req.Username).Or("email = ?", req.Email).Count(&existingUser)
	if existingUser > 0 && req.Provider == "MANUAL" {
		return resp, errors.New("user already exist")
	}

	// TODO: Lakukan Linking jika user login dengan google tetapi user sudah tersimpan melalui registrasi manual sebelumnya

	// Enkripsi Password Jika Registrasi Manual
	if req.Provider == "MANUAL" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		if err != nil {
			return resp, err
		}
		req.Password = string(bytes)
	}

	// Find Base Role
	var baseRole m.Role
	err = s.DB.Model(&m.Role{}).Where("name = ?", "BASE").First(&baseRole).Error

	if err != nil {
		return resp, errors.New("default user role not found, please contact administartor")
	}

	// Insert Ke DB
	newUser := m.User{
		Username:   req.Username,
		Password:   sql.NullString{Valid: false},
		Email:      req.Email,
		Fullname:   req.Fullname,
		Provider:   req.Provider,
		ProviderId: sql.NullString{Valid: false},
		IdRole:     sql.NullInt64{Int64: baseRole.Id_Role, Valid: true},
	}

	if req.Provider == "GOOGLE" {
		newUser.ProviderId = sql.NullString{String: req.ProviderId, Valid: true}
	}

	if req.Provider == "MANUAL" {
		newUser.Password = sql.NullString{String: req.Password, Valid: true}
	}

	err = s.DB.Create(&newUser).Error
	if err != nil {
		log.Error(err)
		return resp, errors.New("failed to create new user")
	}

	resp.Email = req.Email
	resp.Fullname = req.Fullname
	resp.Username = req.Username
	resp.Role = baseRole.Name

	// Returun
	return resp, nil
}
