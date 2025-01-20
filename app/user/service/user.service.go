package user

import v "finance/app/user/validation"

type IUserService interface {
	UserRegistartion(req v.UserRegistrationRequest) (v.UserResponse, error)
	UserLogin(req v.UserLoginRequest) (v.UserResponse, error)
}
