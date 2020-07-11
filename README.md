# GenderBR

Simple lib that returns the gender for a given name based on [Nomes no Brasil 2010](https://censo2010.ibge.gov.br/nomes/#/search) IBGE's research.

### Install
```shell script
$ go get github.com/herlon214/genderBR
```

### Usage
```go
package main
import (
    "fmt"
    "github.com/herlon214/genderBR"
)


func main() {
    names := []string{"João"}
    results := genderBR.For(names)
    
    for _, result := range results {
        fmt.Println(result.Name, "=", result.Gender)
    }
}
```