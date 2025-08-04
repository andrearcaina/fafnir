package main

import (
	"context"
	"errors"
	"fafnir/security-service/internal/db/generated"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/joho/godotenv"

	"fafnir/security-service/internal/config"
	"fafnir/security-service/internal/db"
)

type SeedUserRole struct {
	ID   string `yaml:"id"`
	Role string `yaml:"role"`
}

type SeedFile struct {
	Users []SeedUserRole `yaml:"users"`
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
		params := generated.InsertUserRoleWithIDParams{
			UserID:   uuid.MustParse(u.ID),
			RoleName: u.Role,
		}

		userID, err := dbConn.GetQueries().InsertUserRoleWithID(ctx, params)

		if err != nil {
			return errors.New("failed to insert user: " + err.Error())
		}

		fmt.Printf("Seed in Security DB: Role %s with ID %s\n", u.Role, userID)
	}

	return nil
}
