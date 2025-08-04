package main

import (
	"context"
	"errors"
	"fafnir/user-service/internal/db/generated"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/joho/godotenv"

	"fafnir/user-service/internal/config"
	"fafnir/user-service/internal/db"
)

type SeedUserProfile struct {
	ID        string `yaml:"id"`
	FirstName string `yaml:"first_name"`
	LastName  string `yaml:"last_name"`
}

type SeedFile struct {
	Users []SeedUserProfile `yaml:"users"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Seeder failed: %v", err)
	}
}

func run() error {
	err := godotenv.Load("../../infra/env/.env.dev")
	if err != nil {
		return errors.New("error loading .env file")
	}

	data, err := os.ReadFile("../../infra/postgres/seed.yml")
	if err != nil {
		return errors.New("seed: " + err.Error())
	}

	var seed SeedFile
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return errors.New("seed: " + err.Error())
	}

	cfg := config.NewConfig()

	cfg.DB.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		"localhost", // change for local seeding
		cfg.DB.Port,
		cfg.DB.DbName,
	)

	dbConn, err := db.NewDBConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	ctx := context.Background()

	for _, u := range seed.Users {
		params := generated.InsertUserProfileByIdParams{
			ID:        uuid.MustParse(u.ID),
			FirstName: u.FirstName,
			LastName:  u.LastName,
		}

		userID, err := dbConn.GetQueries().InsertUserProfileById(ctx, params)

		if err != nil {
			return errors.New("failed to insert user: " + err.Error())
		}

		fmt.Printf("Seeded in User DB: %s %s with ID %s\n", u.FirstName, u.LastName, userID)
	}

	return nil
}
