# bind

A minimal dependency injection library using only golang standard libraries.

See [example](https://github.com/mlctrez/bind/blob/cf0c8d0fe0e648eafbb1c39ddc1fe97a99633ddc/example_test.go#L55) for
basic usage.

```go
package main

import (
	"fmt"
	"github.com/mlctrez/bind"
)

type DepOne struct{}

type DepTwo struct {
	DepOne *DepOne
}

func main() {
	binder := bind.New()

	depOne := &DepOne{}
	depTwo := &DepTwo{}

	err := binder.Add(depOne, depTwo)
	if err != nil {
		// handle errors that may occur on startup
	}

	fmt.Printf("dependency injected = %t", depTwo.DepOne != nil)
}

```

[![Go Report Card](https://goreportcard.com/badge/github.com/mlctrez/bind)](https://goreportcard.com/report/github.com/mlctrez/bind)

created by [tigwen](https://github.com/mlctrez/tigwen)
