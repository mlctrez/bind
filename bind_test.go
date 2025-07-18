package bind

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
)

// Common test setup helper
func setupBinder() Binder {
	return New()
}

// Helper to create a logger for testing
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to capture logs for assertions
func bufferLog(t *testing.T, b Binder) *bytes.Buffer {
	buffer := &bytes.Buffer{}
	assert.NoError(t, b.Add(slog.New(slog.NewTextHandler(buffer, nil))))
	return buffer
}

func Test_setupLogger(t *testing.T) {
	b := setupBinder()
	defaultLogger := b.(*binder).Logger
	assert.NotNil(t, defaultLogger.Handler())
	dl := &defaultLogger
	nl := createTestLogger()
	assert.NoError(t, b.Add(nl))
	assert.True(t, &dl != &nl)
}

func Test_Add(t *testing.T) {
	b := setupBinder()

	// Test empty items slice
	assert.NoError(t, b.Add())

	// Test successful and failing additions
	good := &LifecycleTest{}
	bad := &LifecycleTest{Err: fmt.Errorf("add error")}
	assert.ErrorContains(t, b.Add(good, bad), "add error")

	// Test non-pointer addition (should fail)
	assert.ErrorContains(t, b.Add(1), "cannot bind non-pointer type int, all items must be pointers")
}

func Test_Shutdown(t *testing.T) {
	b := setupBinder()
	buffer := bufferLog(t, b)

	// Test successful shutdown
	good := &LifecycleTest{}
	assert.NoError(t, b.Add(good))
	b.Shutdown()

	// Test error during shutdown
	bad := &LifecycleTest{}
	assert.NoError(t, b.Add(bad))
	bad.Err = fmt.Errorf("shutdown error")
	b.Shutdown()
	assert.Contains(t, buffer.String(), `level=ERROR msg="shutdown error"`)
}

func Test_bindField(t *testing.T) {
	b := setupBinder()
	logger := createTestLogger()
	assert.NoError(t, b.Add(logger))

	// Group test cases by binding type
	t.Run("PrivateField", func(t *testing.T) {
		privateLog := &PrivateLog{}
		assert.NoError(t, b.Add(privateLog))
		assert.Nil(t, privateLog.log, "Private fields should not be bound")
	})

	t.Run("PublicField", func(t *testing.T) {
		publicLog := &PublicLog{}
		assert.NoError(t, b.Add(publicLog))
		assert.NotNil(t, publicLog.Log, "Public fields should be bound")
		assert.Equal(t, logger, publicLog.Log)
	})

	t.Run("EmbeddedField", func(t *testing.T) {
		inlineLog := &InlineLog{}
		assert.NoError(t, b.Add(inlineLog))
		assert.NotNil(t, inlineLog.Logger.Handler(), "Embedded fields should be bound")
	})
}

func Test_bindField_EdgeCases(t *testing.T) {
	b := setupBinder()

	t.Run("NoMatchingDependencies", func(t *testing.T) {
		noDeps := &NoDependencies{}
		assert.NoError(t, b.Add(noDeps))
		assert.Equal(t, "", noDeps.Name, "Fields with no matching dependencies should remain empty")
	})

	t.Run("MultipleMatches", func(t *testing.T) {
		service1 := &TestService{Name: "service1"}
		service2 := &TestService{Name: "service2"}
		consumer := &ServiceConsumer{}

		assert.NoError(t, b.Add(service1))
		assert.NoError(t, b.Add(consumer))
		assert.NoError(t, b.Add(service2))

		assert.NotNil(t, consumer.Service, "Service should be bound")
		assert.Equal(t, "service1", consumer.Service.Name, "First matching service should be bound")
	})

	t.Run("ReadOnlyField", func(t *testing.T) {
		readOnly := &ReadOnlyField{}
		assert.NoError(t, b.Add(readOnly), "Adding component with readonly field should succeed")
	})
}

// Interface binding tests
func Test_InterfaceBindings(t *testing.T) {
	t.Run("SingleInterface", func(t *testing.T) {
		b := setupBinder()
		repo := &UserRepository{Users: map[string]string{"1": "Alice", "2": "Bob"}}
		service := &UserInterfaceService{}

		assert.NoError(t, b.Add(repo, service))
		assert.NotNil(t, service.Repository, "Interface should be bound")

		name, err := service.GetUserName("1")
		assert.NoError(t, err)
		assert.Equal(t, "Alice", name, "Bound interface should work correctly")
	})

	t.Run("MultipleInterfaces", func(t *testing.T) {
		b := setupBinder()
		userRepo := &UserRepository{Users: map[string]string{"1": "Alice"}}
		productRepo := &ProductRepository{Products: map[string]string{"A": "Laptop"}}
		compositeService := &CompositeService{}

		assert.NoError(t, b.Add(userRepo, productRepo, compositeService))

		assert.NotNil(t, compositeService.UserRepo, "First interface should be bound")
		assert.NotNil(t, compositeService.ProductRepo, "Second interface should be bound")

		userName, err := compositeService.UserRepo.FindUser("1")
		assert.NoError(t, err)
		assert.Equal(t, "Alice", userName)

		productName, err := compositeService.ProductRepo.FindProduct("A")
		assert.NoError(t, err)
		assert.Equal(t, "Laptop", productName)
	})

	t.Run("SameInterfaceDifferentFields", func(t *testing.T) {
		b := setupBinder()
		memCache := &MemoryCache{Data: map[string][]byte{"key1": []byte("value1")}}
		diskCache := &DiskCache{FilePath: "/tmp/cache"}
		cacheService := &CacheService{}

		assert.NoError(t, b.Add(memCache, diskCache, cacheService))

		assert.NotNil(t, cacheService.PrimaryCache, "First cache field should be bound")
		assert.NotNil(t, cacheService.BackupCache, "Second cache field should be bound")

		_, ok := cacheService.PrimaryCache.(*MemoryCache)
		assert.True(t, ok, "PrimaryCache should be MemoryCache")

		_, ok = cacheService.BackupCache.(*MemoryCache)
		assert.True(t, ok, "BackupCache should be MemoryCache (first match behavior)")
	})

	t.Run("PointerImplementingInterface", func(t *testing.T) {
		b := setupBinder()
		formatter := &CustomFormatter{Prefix: "TEST: "}
		service := &FormatterService{}

		assert.NoError(t, b.Add(formatter, service))

		assert.NotNil(t, service.Formatter, "Formatter should be injected")

		result := service.FormatMessage("Hello")
		assert.Equal(t, "TEST: Hello", result, "Formatter should work correctly")

		_, ok := service.Formatter.(*CustomFormatter)
		assert.True(t, ok, "Formatter should be *CustomFormatter")
	})
}

// ==================== Test Types ====================

// Types for logging tests
type PrivateLog struct {
	log *slog.Logger
}

type PublicLog struct {
	Log *slog.Logger
}

type InlineLog struct {
	slog.Logger
}

// Types for simple dependency tests
type NoDependencies struct {
	Name string
}

type TestService struct {
	Name string
}

type ServiceConsumer struct {
	Service *TestService
}

type ReadOnlyField struct {
	readOnly string
}

// Repository interfaces and implementations
type Repository interface {
	FindUser(id string) (string, error)
}

type UserRepository struct {
	Users map[string]string
}

func (r *UserRepository) FindUser(id string) (string, error) {
	if name, ok := r.Users[id]; ok {
		return name, nil
	}
	return "", fmt.Errorf("user not found")
}

type UserInterfaceService struct {
	Repository Repository
}

func (s *UserInterfaceService) GetUserName(id string) (string, error) {
	return s.Repository.FindUser(id)
}

// Product interfaces and implementations
type ProductRepo interface {
	FindProduct(id string) (string, error)
}

type ProductRepository struct {
	Products map[string]string
}

func (r *ProductRepository) FindProduct(id string) (string, error) {
	if name, ok := r.Products[id]; ok {
		return name, nil
	}
	return "", fmt.Errorf("product not found")
}

// Service using multiple interfaces
type CompositeService struct {
	UserRepo    Repository
	ProductRepo ProductRepo
}

// Cache interfaces and implementations
type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

type MemoryCache struct {
	Data map[string][]byte
}

func (c *MemoryCache) Get(key string) ([]byte, error) {
	if data, ok := c.Data[key]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (c *MemoryCache) Set(key string, value []byte) error {
	c.Data[key] = value
	return nil
}

type DiskCache struct {
	FilePath string
}

func (c *DiskCache) Get(_ string) ([]byte, error) {
	return []byte("mock data"), nil
}

func (c *DiskCache) Set(_ string, _ []byte) error {
	return nil
}

type CacheService struct {
	PrimaryCache Cache
	BackupCache  Cache
}

// Formatter interfaces and implementations
type Formatter interface {
	Format(string) string
}

type CustomFormatter struct {
	Prefix string
}

func (f *CustomFormatter) Format(message string) string {
	return f.Prefix + message
}

type FormatterService struct {
	Formatter Formatter
}

func (s *FormatterService) FormatMessage(message string) string {
	if s.Formatter == nil {
		return "FORMATTER NOT INJECTED: " + message
	}
	return s.Formatter.Format(message)
}
