package bind_test

import (
	"fmt"

	"github.com/mlctrez/bind"
)

// Database represents a database connection with lifecycle management
type Database struct {
	ConnectionString string
	connected        bool
}

func (db *Database) Startup() error {
	fmt.Printf("Connecting to database: %s\n", db.ConnectionString)
	db.connected = true
	return nil
}

func (db *Database) Shutdown() error {
	fmt.Println("Disconnecting from database")
	db.connected = false
	return nil
}

type User struct {
	Name string
}

func (db *Database) GetUser(id int) (*User, error) {
	// a real db would do some sql
	return &User{Name: fmt.Sprintf("User %d", id)}, nil
}

// UserService depends on Database
type UserService struct {
	DB *Database // Will be automatically injected
}

func (us *UserService) Startup() error {
	fmt.Println("UserService starting up")
	return nil
}

func (us *UserService) GetUser(id int) string {
	user, err := us.DB.GetUser(id)
	if err != nil {
		return ""
	}
	return user.Name
}

// Example demonstrates basic usage of the bind dependency injection container
func Example() {
	// Create a new binder
	binder := bind.New()

	// Create and configure dependencies
	db := &Database{ConnectionString: "postgres://localhost/mydb"}
	userService := &UserService{}

	// Add all components to the binder
	// Dependencies will be automatically injected based on type matching
	if err := binder.Add(db, userService); err != nil {
		panic(err)
	}

	// Use the service (dependencies are already injected and started up)
	user := userService.GetUser(123)
	fmt.Printf("Retrieved: %s\n", user)

	// Clean shutdown - calls Shutdown() on all components in reverse order
	binder.Shutdown()

	// Output:
	// Connecting to database: postgres://localhost/mydb
	// UserService starting up
	// Retrieved: User 123
	// Disconnecting from database
}

// ExampleBinder_Add demonstrates adding multiple items at once
func ExampleBinder_Add() {
	binder := bind.New()

	// Create components
	db := &Database{ConnectionString: "sqlite://memory"}
	service := &UserService{}

	// Add multiple items at once
	err := binder.Add(db, service)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Components added successfully")
	binder.Shutdown()

	// Output:
	// Connecting to database: sqlite://memory
	// UserService starting up
	// Components added successfully
	// Disconnecting from database
}

// ExampleBinder_Shutdown demonstrates the shutdown lifecycle
func ExampleBinder_Shutdown() {
	binder := bind.New()

	// Add components in order
	db := &Database{ConnectionString: "test://db"}
	service := &UserService{}

	binder.Add(db, service)

	fmt.Println("Before shutdown")
	// Shutdown calls Shutdown() on components in reverse order
	binder.Shutdown()
	fmt.Println("After shutdown")

	// Output:
	// Connecting to database: test://db
	// UserService starting up
	// Before shutdown
	// Disconnecting from database
	// After shutdown
}
