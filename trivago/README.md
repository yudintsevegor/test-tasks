# February 2020

# Converting hotels from .csv to .xml and .json

### How to run?
Use `Makefile`.
 
 
``
    make all  - upd external pkgs and run a program.
``

### How to run tests?

``
    go test -v
``

### Files:

* `main.go` - func() convert(...) calls here
* `convert.go` - func() convert(...) is stored here
* `structs.go` - structs are stored here
* `main_test.go` - tests are stored here 
* `hotels_test.csv, hotels_test_expeted.json` - files for tests
 

### macOS

I use `macOS Catalina, v10.15.2` and `go version go1.13.8 darwin/amd64`
