package bind

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"reflect"
	"testing"
)

var _ Startup = (*LifecycleTest)(nil)
var _ Shutdown = (*LifecycleTest)(nil)

type LifecycleTest struct {
	Err error
}

func (t *LifecycleTest) Error() error {
	err := t.Err
	t.Err = nil
	return err
}
func (t *LifecycleTest) Startup() error  { return t.Error() }
func (t *LifecycleTest) Shutdown() error { return t.Error() }

func Test_buildItem(t *testing.T) {
	_, err := buildItem(nil)
	assert.ErrorContains(t, err, "cannot bind nil value")

	_, err = buildItem(0)
	assert.ErrorContains(t, err, "cannot bind non-pointer type int, all items must be pointers")

	// Test nil pointer
	var nilPtr *LifecycleTest
	_, err = buildItem(nilPtr)
	assert.ErrorContains(t, err, "cannot bind nil pointer of type *bind.LifecycleTest")

	ti := &LifecycleTest{}
	theItem, err := buildItem(ti)
	assert.NoError(t, err)
	assert.Equal(t, theItem.target, ti)
	assert.Equal(t, theItem.typeOf, reflect.TypeOf(ti))
	assert.Equal(t, theItem.valueOf, reflect.ValueOf(ti))
}

func Test_Lifecycle(t *testing.T) {
	ti := &LifecycleTest{}
	ti.Err = fmt.Errorf("startup error")

	bi, err := buildItem(ti)
	assert.NoError(t, err)
	assert.ErrorContains(t, bi.Startup(), "startup error")
	assert.NoError(t, bi.Startup())

	ti.Err = fmt.Errorf("shutdown error")
	assert.ErrorContains(t, bi.Shutdown(), "shutdown error")
	assert.NoError(t, bi.Shutdown())
	assert.NoError(t, bi.Shutdown())

	bi, err = buildItem(slog.New(slog.NewTextHandler(io.Discard, nil)))
	assert.NoError(t, err)
	assert.NoError(t, bi.Startup())
	assert.NoError(t, bi.Shutdown())

}
