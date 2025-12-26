package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	dbURL          string
	migrationsPath string
	command        string
	steps          int
)

func init() {
	flag.StringVar(&dbURL, "database-url", os.Getenv("DATABASE_URL"), "Database connection URL (or set DATABASE_URL env)")
	flag.StringVar(&migrationsPath, "path", "file://migrations", "Path to migrations directory")
	flag.StringVar(&command, "command", "up", "Migration command: up, down, force, version, drop")
	flag.IntVar(&steps, "steps", 0, "Number of steps for up/down (0 = all)")
}

func main() {
	flag.Parse()

	// Default database URL if not provided
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/employee_service?sslmode=disable"
		log.Printf("Using default database URL: %s", dbURL)
	}

	// Open database connection
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Execute command
	switch command {
	case "up":
		if steps > 0 {
			log.Printf("Applying %d migration(s) up...", steps)
			if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration failed: %v", err)
			}
		} else {
			log.Println("Applying all migrations up...")
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration failed: %v", err)
			}
		}
		log.Println("Migrations applied successfully")

	case "down":
		if steps > 0 {
			log.Printf("Rolling back %d migration(s)...", steps)
			if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration rollback failed: %v", err)
			}
		} else {
			log.Println("Rolling back all migrations...")
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration rollback failed: %v", err)
			}
		}
		log.Println("Migrations rolled back successfully")

	case "force":
		if steps == 0 {
			log.Fatal("Must specify -steps for force command")
		}
		log.Printf("Forcing version to %d...", steps)
		if err := m.Force(steps); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		log.Println("Version forced successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			log.Fatalf("Failed to get version: %v", err)
		}
		if err == migrate.ErrNilVersion {
			log.Println("No migrations applied yet")
		} else {
			fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)
		}

	case "drop":
		log.Println("WARNING: This will drop all tables!")
		log.Println("Press Ctrl+C to cancel, or wait 5 seconds to continue...")
		if err := m.Drop(); err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		log.Println("Database dropped successfully")

	default:
		log.Fatalf("Unknown command: %s (available: up, down, force, version, drop)", command)
	}
}
