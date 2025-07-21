package schema

import (
	"github.com/graphql-go/graphql"
)

var LoginResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LoginResponse",
	Fields: graphql.Fields{
		"code": &graphql.Field{
			Type: graphql.Int,
		},
		"message": &graphql.Field{
			Type: graphql.String,
		},
		"error": &graphql.Field{
			Type: graphql.String,
		},
	},
})
var LoginInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "LoginInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"user": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"password": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

// QueryType is for GET requests (used for getting data)
var QueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"health": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// MutationType is for POST, PUT, PATCH, DELETE requests (used for modifying/mutating data)
var MutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"login": &graphql.Field{
			Type: LoginResponseType,
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(LoginInputType),
				},
			},
		},
	},
})

func NewAuthSchema(schemaConfig graphql.SchemaConfig) (graphql.Schema, error) {
	return graphql.NewSchema(schemaConfig)
}
