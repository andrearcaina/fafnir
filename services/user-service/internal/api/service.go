package api

import "fmt"

type Service struct{}

func NewUserService() *Service {
	return &Service{}
}

func (s *Service) GetUserInfo(UserId int) string {
	return fmt.Sprintf("User Info for ID: %d", UserId)
}
