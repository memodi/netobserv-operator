# NetObserv Backend Testing 

## Prerequisites
- Ginkgo CLI (installed via `go install github.com/onsi/ginkgo/v2/ginkgo`)

## Running Tests

### Run all tests
```bash
ginkgo
```
or to make sure the version of cli and packages is matching

```bash
go run github.com/onsi/ginkgo/v2/ginkgo
```

### Run tests with custom focus with verbose output

```bash
go run github.com/onsi/ginkgo/v2/ginkgo --focus="87145"
```

## Writing New Tests

### Create a new test file
```bash
ginkgo generate <test_name>
```

### Add labels to your tests
When writing tests, use labels to categorize them:

```go
var _ = Describe("My Integration Test", Label("integration"), func() {
    // Your integration test code
})
```
## Additional Resources

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
