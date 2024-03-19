# Kuro the simple JSON http client

Kuro is designed to be the simplest way possible to make http requests. It sends an HTTP request and unmarshals json from the response in just one function call.

```
package main

import (
	"fmt"
	"log"

	"github.com/duythinht/kuro"
)

type Product struct {
	kuro.Response // Markup that struct is a response

	ID                 int
	Title              string
	Description        string
	Price              int
	DiscountPercentage float64
	Rating             float64
	Stock              int
}

func main() {
	product, err := kuro.Get[Product](context.Background(), "https://dummyjson.com/products/1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", product)
}
```

## How to

### With Header

```
product, err := kuro.Get[Product](context.Background(), "https://dummyjson.com/products/1", kuro.WithHeader("content-type", "application/json"))
```

### Do Post

```go
type ProductRequest struct {
    Title       string
    Description string
}

type ProductResponse struct {
    kuro.Response
    ID          string
    Title       string
    Description string
}

product, err := kuro.Post[ProductResponse](
    context.Background(),
    "https://dummyjson.com/products",
    &ProductRequest{
        Title: "Test",
        Description: "Test product description",
    }, 
    kuro.WithHeader("content-type", "application/json"),
)
```

## Handle error

```go
_, err := kuro.Post[Body](
    context.Background(),
    "https://dummyjson.com/products",
    &Request{
        //.. data here
    }, 
    kuro.WithHeader("content-type", "application/json"),
)

switch {
case errors.Is(err, kuro.ErrClientMalfunction):
    // do with client malfunction
case errors.Is(err, kuro.Err4xx) || errors.Is(err, kuro.Err5xx):
    var kuroError *kuro.Error
    if errors.As(err, &kuroError) {
        // do with kuroError
    }
}
```