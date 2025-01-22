package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	v "finance/app/user/validation"
	m "finance/model"
	"finance/provider/jwt"
	g "finance/utility/generator"
)

type UserService struct {
	DB             *gorm.DB
	Validate       *validator.Validate
	AdditionalData UserServiceAdditionalData
}

func NewUserService(db *gorm.DB, v *validator.Validate, data UserServiceAdditionalData) IUserService {
	return &UserService{
		DB:             db,
		Validate:       v,
		AdditionalData: data,
	}
}

func (s UserService) UserRegistartion(req v.UserRegistrationRequest) (v.UserResponse, error) {
	var err error
	var resp v.UserResponse
	var newUser m.User
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
	type UserQueryResult struct {
		IdUser     int64
		Email      string
		Fullname   string
		Username   string
		RoleId     string
		RoleName   string
		Provider   string
		Password   sql.NullString
		ProviderId sql.NullString
	}
	var existingUser UserQueryResult
	countExistingUser := s.DB.Raw(`SELECT u.*, r.name role_name, r.id_role FROM "user" u JOIN "role" r ON r.id_role = u.id_role WHERE u.username = ? or u.email = ?`, req.Username, req.Email).Scan(&existingUser).RowsAffected
	if countExistingUser > 0 && req.Provider == "MANUAL" {
		return resp, errors.New("user already exist")
	}

	// Jika Awalnya Login manual (countExistingUser lebih dari 0) dan sekarang login dengan google, lakukan linking
	if countExistingUser > 0 && req.Provider == "GOOGLE" {
		if existingUser.Provider == "MANUAL" && !existingUser.ProviderId.Valid {
			err = s.DB.Model(&m.User{}).Where("id_user", existingUser.IdUser).Update("provider_id", req.ProviderId).Error
			if err != nil {
				return resp, err
			}
		}

		resp.Email = req.Email
		resp.Fullname = req.Fullname
		resp.Username = req.Username
		resp.Role = existingUser.RoleName
		return resp, nil
	}

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
	newUser.Username = req.Username
	newUser.Password = sql.NullString{Valid: false}
	newUser.Email = req.Email
	newUser.Fullname = req.Fullname
	newUser.Provider = req.Provider
	newUser.ProviderId = sql.NullString{Valid: false}
	newUser.IdRole = sql.NullInt64{Int64: baseRole.Id_Role, Valid: true}

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

func (s *UserService) UserLogin(req v.UserLoginRequest) (v.UserResponse, error) {
	var resp v.UserResponse
	var err error

	// Validasi Request Body
	if req.Provider == "MANUAL" {
		err = s.Validate.Struct(v.ManualLoginRequest{
			UsernameOrEmail: req.UsernameOrEmail,
			Password:        req.Password,
			Provider:        "MANUAL",
		})
	}

	if req.Provider == "GOOGLE" {
		err = s.Validate.Struct(v.GoogleLoginRequest{
			UsernameOrEmail: req.UsernameOrEmail,
			ProviderId:      req.ProviderId,
			Provider:        "GOOGLE",
		})
	}

	if err != nil {
		return resp, err
	}

	type UserQueryResult struct {
		Email      string
		Fullname   string
		Username   string
		RoleId     string
		RoleName   string
		Provider   string
		Password   sql.NullString
		ProviderId sql.NullString
	}
	var existingUser UserQueryResult
	// Cari User berdasarkan username ataupun password
	err = s.DB.Raw(`SELECT u.*, r.name role_name, r.id_role FROM "user" u JOIN "role" r ON r.id_role = u.id_role WHERE u.username = ? or u.email = ?`, req.UsernameOrEmail, req.UsernameOrEmail).Or("u.email = ?", req.UsernameOrEmail).Scan(&existingUser).Error
	if err != nil {
		return resp, errors.New("username or password not match")
	}

	// Jika login manual maka lakukan validasi password
	if req.Provider == "MANUAL" && existingUser.Password.Valid {
		err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password.String), []byte(req.Password))
	}

	// Jika login mengggunakan google maka lakukan validasi provider id
	if req.Provider == "GOOGLE" && existingUser.ProviderId.Valid {
		if existingUser.ProviderId.String != req.ProviderId {
			err = errors.New("provider id not match")
		}
	}

	if err != nil {
		return resp, errors.New("username or password not matchs")
	}

	// Generate JWT
	tokenString, err := jwt.GenerateJWT(jwt.TokenData{
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.RoleName,
		Fullname: existingUser.Fullname,
	})

	if err != nil {
		return resp, errors.New("failed to process login, contact administrator")
	}

	// Response
	resp.Token = &tokenString
	resp.Email = existingUser.Email
	resp.Fullname = existingUser.Fullname
	resp.Username = existingUser.Username
	resp.Role = existingUser.RoleName

	return resp, nil
}

func (s *UserService) GenerateGoogleLoginUrl() (v.GoogleRedirectResponse, error) {
	var resp v.GoogleRedirectResponse
	googleBaseUrl := "https://accounts.google.com/o/oauth2/v2/auth"
	googleClinetId := s.AdditionalData.GoogleClinetId // TODO: implement read from ENV
	serverUrl := s.AdditionalData.ServerUrl           // TODO: implement read from ENV
	callbackUrl := serverUrl + "/api/user/v1/login/google/callback"
	state, err := g.GenerateRandomBase64Url()
	if err != nil {
		return resp, errors.New("failed to generate token")
	}
	tempUrl, err := url.Parse(googleBaseUrl) // TODO: implement random string
	if err != nil {
		return resp, errors.New("failed to generate google login link")
	}

	redirectUrl := *tempUrl
	copyUrl := redirectUrl.Query()

	copyUrl.Add("response_type", "code")
	copyUrl.Add("scope", "openid email profile")
	copyUrl.Add("redirect_uri", callbackUrl)
	copyUrl.Add("prompt", "select_account")
	copyUrl.Add("client_id", googleClinetId)
	copyUrl.Add("state", state)
	redirectUrl.RawQuery = copyUrl.Encode()

	resp.RedirectUrl = redirectUrl.String()
	resp.State = state
	return resp, nil
}

func (s *UserService) GoogleLoginCallback(payload string) (v.GoogleUserInfo, error) {
	var profileInfoData v.GoogleUserInfo

	// Load All Data
	googleTokenBaseUrl := "https://oauth2.googleapis.com/token"
	googleInfoBaseUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	serverUrl := s.AdditionalData.ServerUrl // TODO: implement read from ENV
	callbackUrl := serverUrl + "/api/user/v1/login/google/callback"

	// Send request to get access token
	tempTokenUrl, err := url.Parse(googleTokenBaseUrl)
	if err != nil {
		return profileInfoData, err
	}
	tokenUrl := *tempTokenUrl
	copyUrl := tokenUrl.Query()
	copyUrl.Add("client_secret", s.AdditionalData.GoogleClinetSecret)
	copyUrl.Add("client_id", s.AdditionalData.GoogleClinetId)
	copyUrl.Add("code", payload)
	copyUrl.Add("grant_type", "authorization_code")
	copyUrl.Add("redirect_uri", callbackUrl)
	tokenUrl.RawQuery = copyUrl.Encode()

	tokenReq, err := http.NewRequest(http.MethodPost, tokenUrl.String(), nil)
	if err != nil {
		return profileInfoData, err
	}

	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Add("Accept", "application/json")
	tokenReq.Header.Add("User-Agent", "go_http")

	client := &http.Client{Timeout: 10 * time.Second}
	respToken, err := client.Do(tokenReq)
	if err != nil {
		return profileInfoData, err
	}
	defer respToken.Body.Close()

	var tokenData v.OAuth2Token
	err = json.NewDecoder(respToken.Body).Decode(&tokenData)
	if err != nil {
		return profileInfoData, err
	}

	// Send request to get user data
	googleInfoUrl, err := url.Parse(googleInfoBaseUrl)
	if err != nil {
		return profileInfoData, err
	}

	profileInfoReq, err := http.NewRequest(http.MethodPost, googleInfoUrl.String(), nil)
	if err != nil {
		return profileInfoData, err
	}

	profileInfoReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))
	respInfo, err := client.Do(profileInfoReq)
	if err != nil {
		return profileInfoData, err
	}

	err = json.NewDecoder(respInfo.Body).Decode(&profileInfoData)
	if err != nil {
		return profileInfoData, err
	}

	return profileInfoData, nil
}
