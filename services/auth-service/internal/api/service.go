package api

type Service struct{}

func NewAuthService() *Service {
	return &Service{}
}

func (s *Service) Login(username, password string) bool {
	// placeholder logic for authentication (just as a test)
	return username == "username" && password == "pass"
}
