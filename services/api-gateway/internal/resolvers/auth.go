package resolvers

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"log"

	"github.com/andrearcaina/den/services/api-gateway/internal/clients"
)

type AuthPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MutationResolver struct {
	AuthClient clients.AuthClient
}

func NewMutationResolver(authClient *clients.AuthClient) *MutationResolver {
	return &MutationResolver{
		AuthClient: *authClient,
	}
}

// LoginResolver handles the GraphQL 'login' mutation
func (r *MutationResolver) LoginResolver(p graphql.ResolveParams) (interface{}, error) {
	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid 'input' argument type")
	}

	username, ok := input["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("'user' is required and must be a string")
	}

	password, ok := input["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("'password' is required and must be a string")
	}

	authResp, err := r.AuthClient.Login(p.Context, username, password)
	if err != nil {
		log.Printf("Error calling auth service for login: %v", err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &AuthPayload{
		Code:    authResp.Code,
		Message: authResp.Message,
	}, nil
}
