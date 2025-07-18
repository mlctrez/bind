package bind

import (
	"io"
	"log/slog"
	"reflect"
)

// Binder provides dependency injection with lifecycle management.
// Items are registered as pointers and dependencies are automatically injected
// into struct fields based on type matching.
type Binder interface {
	// Add registers one or more items for dependency injection.
	// Items must be pointers to structs or other types.
	Add(items ...any) error
	// Shutdown calls Shutdown() on all registered items in reverse order.
	Shutdown()
}

var _ Binder = (*binder)(nil)

// New creates a new dependency injection binder with a default logger
// that discards output. Replace the logger by adding a *slog.Logger instance.
func New() Binder {
	b := &binder{items: make([]*item, 0)}
	b.Logger = *slog.New(slog.NewTextHandler(io.Discard, nil))
	return b
}

type binder struct {
	slog.Logger
	items []*item
}

func (b *binder) Add(items ...any) error {
	if len(items) == 0 {
		return nil
	}

	for _, it := range items {
		if err := b.add(it); err != nil {
			return err
		}
	}
	return nil
}

func (b *binder) Shutdown() {
	for i := len(b.items) - 1; i >= 0; i-- {
		theItem := b.items[i]
		if theItem != nil {
			if _, ok := theItem.target.(Shutdown); ok {
				b.Info("shutting down", "item", theItem.typeOf)
			}
			if err := theItem.Shutdown(); err != nil {
				b.Error("shutdown error", "type", theItem.typeOf, "error", err)
			}
		}
	}
	b.items = make([]*item, 0)
}

func (b *binder) add(i any) error {
	it, err := buildItem(i)
	if err != nil {
		return err
	}

	// Handle logger replacement
	if logger, ok := i.(*slog.Logger); ok {
		b.Logger = *logger
	}

	// Only bind fields if it's a struct pointer
	if it.typeOf.Kind() == reflect.Ptr && it.typeOf.Elem().Kind() == reflect.Struct {
		elemType := it.typeOf.Elem()
		for f := 0; f < elemType.NumField(); f++ {
			b.bindField(elemType.Field(f), it.valueOf.Elem().Field(f))
		}
	}

	// Call startup if implemented
	if _, ok := it.target.(Startup); ok {
		b.Info("starting", "item", it.typeOf)
	}
	if err = it.Startup(); err != nil {
		return err
	}

	b.items = append(b.items, it)
	return nil
}

func (b *binder) bindField(field reflect.StructField, fieldValue reflect.Value) {
	for _, dep := range b.items {
		if !fieldValue.CanSet() {
			continue
		}
		// Match pointer to pointer
		if dep.typeOf == fieldValue.Type() {
			fieldValue.Set(dep.valueOf)
			break
		}
		// Match a pointer to value (inject a pointer into the value field)
		if dep.typeOf.Kind() == reflect.Ptr && dep.typeOf.Elem() == field.Type {
			fieldValue.Set(dep.valueOf.Elem())
			break
		}
		// Match interface to implementing type
		// If field is an interface type and the dependency implements it
		if field.Type.Kind() == reflect.Interface && dep.typeOf.Implements(field.Type) {
			fieldValue.Set(dep.valueOf)
			break
		}
	}
}
