package auth

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(username, password string) bool {
	// placeholder logic for authentication (just as a test)
	return username == "user" && password == "pass"
}
