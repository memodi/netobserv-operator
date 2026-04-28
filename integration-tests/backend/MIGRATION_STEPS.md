# copy test files.

```
~/Repos/openshift-tests-private/test/extended$ cp netobserv/* ~/Repos/network-observability-operator/integration-tests/backend/
```
# copy dependencies until compilation succesful

copy the files
```
~/Repos/openshift-tests-private/test/extended$ cp -r util ~/Repos/network-observability-operator/integration-tests/backend/
~/Repos/openshift-tests-private/test/extended$ cp -r scheme ~/Repos/network-observability-operator/integration-tests/backend/
~/Repos/openshift-tests-private/test/extended$ cp -r testdata ~/Repos/network-observability-operator/integration-tests/backend/
```
add to go.mod replace commands, use the origin go.mod as a teamplate
```go
replace (
	github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2 v2.6.1-0.20240806135314-3946b2b7b2a8
	bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	github.com/jteeuwen/go-bindata => github.com/jteeuwen/go-bindata v3.0.8-0.20151023091102-a0ff2567cfb7+incompatible
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0
#and more...
)
```

When executing go mod tidy `export GOPRIVATE="github.com/openshift/*"` is requied.

The following addition to go.mod:
```go
exclude (
  // Exclude old unified containerd to avoid conflicts with split API module
  github.com/containerd/containerd v1.7.18
)
```
fixes `go mod tidy` issues.

# adjust the execution of tests to ignore the kubernetes tests
As k8s.io/kubernetes/test/e2e/framework is imported, it triggters tests inside of that package -> the test suite need to be adjusted to ignore these by default.

# file loading
File loading need to be changed to use the filepath go package instead of the `compat_otp.FixturePath` as that requires the bindata generation. Using the golang filepath package allows for just refferencing the files.

# package declaration
Change the package declaration in all test files from whatever it was in openshift-tests-private to:
```go
package e2etests
```

# filepath usage pattern
Replace all file loading patterns with filepath-based approach. Import filepath with alias:
```go
import (
    filePath "path/filepath"
)
```

Then use this pattern for loading templates:
```go
baseDir, _ := filepath.Abs("./testdata")
templatePath := filepath.Join(baseDir, "subdirectory", "template-file.yaml")
```

Examples of the changes:
- `operator.go`: Uses `filePath.Abs("./testdata")` and `filePath.Join()` for namespace templates
- `test_flowcollector.go`: Uses `filepath.Abs("./testdata")` and `filepath.Join()` for all template paths
- `loki_storage.go`: Uses `filepath.Abs("./testdata")` and `filepath.Join()` for ODF and minIO templates
- `multitenants.go`: Uses `filePath.Abs("./testdata")` and `filePath.Join()` for CRB templates

# CLI initialization
Use the compat_otp CLI initialization with kubeconfig:
```go
oc = compat_otp.NewCLI("netobserv", compat_otp.KubeConfigPath())
```

# custom test suite (backend_suite_test.go)
Create a custom test suite file with:

1. **Init function** for framework flags:
```go
func init() {
    // Initialize framework flags - must be done before flag.Parse()
    exutil.InitStandardFlags()
}
```

2. **BeforeSuite** hook for test initialization:
```go
var _ = BeforeSuite(func() {
    // Parse flags
    flag.Parse()
    
    // Set up provider config after parsing flags
    e2eframework.AfterReadingAllFlags(exutil.TestContext)
    
    // Initialize test
    Expect(exutil.InitTest(false)).NotTo(HaveOccurred())
    
    // Additional setup if needed (e.g., CHECK_NOO_EXISTS environment variable check)
})
```

3. **TestBackend function** with WithCleanup wrapper and focus filter:
```go
func TestBackend(t *testing.T) {
    exutil.WithCleanup(func() {
        RegisterFailHandler(Fail)
        
        suiteConfig, reporterConfig := GinkgoConfiguration()
        
        // Apply focus filter to only run sig-netobserv tests
        if len(suiteConfig.FocusStrings) > 0 {
            combinedFocus := make([]string, len(suiteConfig.FocusStrings))
            for i, userFocus := range suiteConfig.FocusStrings {
                combinedFocus[i] = "sig-netobserv.*" + userFocus
            }
            suiteConfig.FocusStrings = combinedFocus
        } else {
            suiteConfig.FocusStrings = []string{"sig-netobserv"}
        }
        
        // Configure reporter settings
        suiteConfig.EmitSpecProgress = true
        suiteConfig.OutputInterceptorMode = "none"
        reporterConfig.SilenceSkips = true
        reporterConfig.NoColor = true
        reporterConfig.Succinct = true
        reporterConfig.Verbose = false
        
        RunSpecs(t, "Backend Suite", suiteConfig, reporterConfig)
    })
}
```

4. **Custom reporting hooks** for better test output:
   - `ReportBeforeSuite`: Print suite information
   - `ReportAfterEach`: Print individual test results with timing
   - `ReportAfterSuite`: Print summary statistics

See `backend_suite_test.go` for the complete implementation.

# remove temporary directories
The `util/` and `scheme/` directories that were copied during migration can be removed - their dependencies are now resolved through the proper imports in go.mod.

# update go.mod replace directives
Ensure the go.mod has all necessary replace directives, particularly:
- `github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2` (with appropriate version)
- All k8s.io/* packages pointing to github.com/openshift/kubernetes/staging/src/k8s.io/*
- See the main go.mod in the repo root for the complete list

The `exclude` directive for containerd may or may not be needed depending on your dependency resolution - test with `go mod tidy` first.
