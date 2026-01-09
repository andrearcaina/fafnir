package api

import (
	"context"
)

type Server struct {
}

func NewServer() *Server {
	return nil
}

func (s *Server) Run() error {
	return nil
}

func (s *Server) Close(ctx context.Context) error {
	return nil
}
