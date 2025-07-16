package bind

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
)

func Test_setupLogger(t *testing.T) {
	b := New()
	defaultLogger := b.(*binder).Logger
	assert.NotNil(t, defaultLogger.Handler())
	dl := &defaultLogger
	nl := slog.New(slog.NewTextHandler(io.Discard, nil))
	assert.NoError(t, b.Add(nl))
	assert.True(t, &dl != &nl)
}

func Test_Add(t *testing.T) {
	b := New()

	// Test empty items slice
	assert.NoError(t, b.Add())

	good := &LifecycleTest{}
	bad := &LifecycleTest{Err: fmt.Errorf("add error")}
	assert.ErrorContains(t, b.Add(good, bad), "add error")

	assert.ErrorContains(t, b.Add(1), "cannot bind non-pointer type int, all items must be pointers")
}

func Test_Shutdown(t *testing.T) {
	b := New()
	buffer := bufferLog(t, b)
	good := &LifecycleTest{}
	assert.NoError(t, b.Add(good))
	b.Shutdown()

	bad := &LifecycleTest{}
	assert.NoError(t, b.Add(bad))

	bad.Err = fmt.Errorf("shutdown error")
	b.Shutdown()
	assert.Contains(t, buffer.String(), `level=ERROR msg="shutdown error"`)
}

func bufferLog(t *testing.T, b Binder) *bytes.Buffer {
	buffer := &bytes.Buffer{}
	assert.NoError(t, b.Add(slog.New(slog.NewTextHandler(buffer, nil))))
	return buffer
}

func Test_bindField(t *testing.T) {
	b := New()

	// Add a logger that will be injected
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	assert.NoError(t, b.Add(logger))

	// Test private field binding (pointer to pointer)
	privateLog := &PrivateLog{}
	assert.NoError(t, b.Add(privateLog))
	// Verify the private field was NOT bound (private fields can't be set)
	assert.Nil(t, privateLog.log)

	// Test public field binding (pointer to pointer)
	publicLog := &PublicLog{}
	assert.NoError(t, b.Add(publicLog))
	// Verify the public field was bound
	assert.NotNil(t, publicLog.Log)
	assert.Equal(t, logger, publicLog.Log)

	// Test embedded field binding (pointer to value)
	inlineLog := &InlineLog{}
	assert.NoError(t, b.Add(inlineLog))
	// Verify the embedded field was bound
	assert.NotNil(t, inlineLog.Logger.Handler())
}

func Test_bindField_EdgeCases(t *testing.T) {
	b := New()

	// Test binding with no dependencies
	noDeps := &NoDependencies{}
	assert.NoError(t, b.Add(noDeps))
	assert.Equal(t, "", noDeps.Name) // Should remain empty

	// Test binding with multiple potential matches
	service1 := &TestService{Name: "service1"}
	service2 := &TestService{Name: "service2"}
	consumer := &ServiceConsumer{}

	assert.NoError(t, b.Add(service1))
	assert.NoError(t, b.Add(consumer))
	assert.NoError(t, b.Add(service2))

	// Should bind to the first matching service (service1)
	assert.NotNil(t, consumer.Service)
	assert.Equal(t, "service1", consumer.Service.Name)

	// Test readonly field (should be skipped)
	readOnly := &ReadOnlyField{}
	assert.NoError(t, b.Add(readOnly))
}

type PrivateLog struct {
	log *slog.Logger
}

type PublicLog struct {
	Log *slog.Logger
}

type InlineLog struct {
	slog.Logger
}

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
