package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceDB  string
	ConfigPath string
	DataPath   string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type SeedUser struct {
	ID       string `yaml:"user_id"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type SeedUserProfile struct {
	ID        string `yaml:"user_id"`
	FirstName string `yaml:"first_name"`
	LastName  string `yaml:"last_name"`
}

type SeedUserRole struct {
	ID       string `yaml:"user_id"`
	RoleName string `yaml:"role_name"`
}

type AuthSeedFile struct {
	Users []SeedUser `yaml:"users"`
}

type UserSeedFile struct {
	Users []SeedUserProfile `yaml:"users"`
}

type SecuritySeedFile struct {
	Users []SeedUserRole `yaml:"users"`
}

func main() {
	var serviceDB = flag.String("db", "", "Service DB to seed (auth, user, security, all)")
	var configPath = flag.String("config", "../../infra/env/.env.dev", "Path to environment config file")
	var dataPath = flag.String("data", "./seed.yml", "Path to seed data file")
	flag.Parse()

	if *serviceDB == "" {
		log.Fatal("Service must be specified. Use: auth, user, security, or all")
	}

	config := &Config{
		ServiceDB:  *serviceDB,
		ConfigPath: *configPath,
		DataPath:   *dataPath,
	}

	if err := run(config); err != nil {
		log.Fatalf("Seeder failed: %v", err)
	}
}

func run(config *Config) error {
	if err := godotenv.Load(config.ConfigPath); err != nil {
		return errors.New("errors loading .env file")
	}

	switch config.ServiceDB {
	case "auth":
		return seedAuthService(config)
	case "user":
		return seedUserService(config)
	case "security":
		return seedSecurityService(config)
	case "all":
		for _, db := range []string{"auth", "user", "security"} {
			config.ServiceDB = db
			if err := run(config); err != nil {
				return errors.New("failed to seed " + db + " service db")
			}
		}
		return nil
	default:
		return errors.New("unknown service db specified: " + config.ServiceDB)
	}
}

func seedAuthService(config *Config) error {
	dbConfig := &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("AUTH_DB"),
	}

	db, err := connectDB(dbConfig)
	if err != nil {
		return errors.New("failed to connect to auth database")
	}
	defer func() {
		_ = db.Close()
	}()

	var seed AuthSeedFile
	if err := tryReadAndUnmarshalFile(config.DataPath, &seed); err != nil {
		return err
	}

	ctx := context.Background()
	for _, user := range seed.Users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("failed to hash password")
		}

		query := `INSERT INTO users (id, email, password_hash, created_at, updated_at) 
				  VALUES ($1, $2, $3, NOW(), NOW()) 
				  ON CONFLICT (id) DO NOTHING`

		_, err = db.ExecContext(ctx, query, uuid.MustParse(user.ID), user.Email, string(hashedPassword))
		if err != nil {
			return errors.New("failed to insert user")
		}

		log.Printf("Seeded in Auth DB: User %s with ID %s\n", user.Email, user.ID)
	}

	return nil
}

func seedUserService(config *Config) error {
	dbConfig := &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("USER_DB"),
	}

	db, err := connectDB(dbConfig)
	if err != nil {
		return errors.New("failed to connect to user database")
	}
	defer func() {
		_ = db.Close()
	}()

	var seed UserSeedFile
	if err := tryReadAndUnmarshalFile(config.DataPath, &seed); err != nil {
		return err
	}

	ctx := context.Background()
	for _, user := range seed.Users {
		query := `INSERT INTO user_profiles (id, first_name, last_name, created_at, updated_at) 
				  VALUES ($1, $2, $3, NOW(), NOW()) 
				  ON CONFLICT (id) DO NOTHING`

		_, err = db.ExecContext(ctx, query, uuid.MustParse(user.ID), user.FirstName, user.LastName)
		if err != nil {
			return errors.New("failed to insert user profile")
		}

		log.Printf("Seeded in User DB: Profile %s %s with ID %s\n", user.FirstName, user.LastName, user.ID)
	}

	return nil
}

func seedSecurityService(config *Config) error {
	dbConfig := &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("SECURITY_DB"),
	}

	db, err := connectDB(dbConfig)
	if err != nil {
		return errors.New("failed to connect to security database")
	}
	defer func() {
		_ = db.Close()
	}()

	var seed SecuritySeedFile
	if err := tryReadAndUnmarshalFile(config.DataPath, &seed); err != nil {
		return err
	}

	ctx := context.Background()
	for _, user := range seed.Users {
		query := `INSERT INTO users_roles (user_id, role_name) 
				  VALUES ($1, $2)
				  ON CONFLICT (user_id, role_name) DO NOTHING`

		_, err = db.ExecContext(ctx, query, uuid.MustParse(user.ID), user.RoleName)
		if err != nil {
			return errors.New("failed to insert user role")
		}

		log.Printf("Seeded in Security DB: User %s with role %s\n", user.ID, user.RoleName)
	}

	return nil
}

func tryReadAndUnmarshalFile(dataPath string, v interface{}) error {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return errors.New("failed to read seed file")
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return errors.New("failed to parse seed file")
	}

	return nil
}

func connectDB(config *DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
