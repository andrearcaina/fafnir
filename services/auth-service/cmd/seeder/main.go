package main

import (
	"context"
	"errors"
	"fafnir/auth-service/internal/db"
	"fafnir/auth-service/internal/db/generated"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	"github.com/joho/godotenv"

	"fafnir/auth-service/internal/config"
)

type SeedUser struct {
	ID       string `yaml:"id"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type SeedFile struct {
	Users []SeedUser `yaml:"users"`
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

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("failed to hash password: " + err.Error())
		}

		params := generated.InsertUserWithIDParams{
			ID:           uuid.MustParse(u.ID),
			Email:        u.Email,
			PasswordHash: string(hashedPassword),
		}

		userID, err := dbConn.GetQueries().InsertUserWithID(ctx, params)

		if err != nil {
			return errors.New("failed to insert user: " + err.Error())
		}

		fmt.Printf("Seeded in Auth DB: User %s with ID %s\n", u.Email, userID)
	}

	return nil
}
