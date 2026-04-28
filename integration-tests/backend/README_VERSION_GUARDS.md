# OCP Version Guards in Ginkgo Integration Tests

## Overview
Integration tests in `integration-tests/backend/` can be conditionally skipped based on OpenShift cluster version using runtime checks.

## Setup (One-time)

### 1. Version Helper (`integration-tests/backend/version_helper.go`)
```go
package e2etests

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    
    . "github.com/onsi/ginkgo/v2"
    configv1 "github.com/openshift/api/config/v1"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type OCPVersion struct {
    Major int
    Minor int
}

var clusterVersion *OCPVersion

func GetOCPVersion(ctx context.Context, k8sClient client.Client) (*OCPVersion, error) {
    if clusterVersion != nil {
        return clusterVersion, nil
    }
    
    cv := &configv1.ClusterVersion{}
    err := k8sClient.Get(ctx, types.NamespacedName{Name: "version"}, cv)
    if err != nil {
        return nil, err
    }
    
    version := cv.Status.Desired.Version
    parts := strings.Split(version, ".")
    if len(parts) < 2 {
        return nil, fmt.Errorf("invalid version: %s", version)
    }
    
    major, _ := strconv.Atoi(parts[0])
    minor, _ := strconv.Atoi(parts[1])
    
    clusterVersion = &OCPVersion{Major: major, Minor: minor}
    return clusterVersion, nil
}

func (v *OCPVersion) AtLeast(major, minor int) bool {
    if v.Major > major {
        return true
    }
    return v.Major == major && v.Minor >= minor
}

func (v *OCPVersion) String() string {
    return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// SkipIfOCPBelow skips test if cluster version is below requirement
func SkipIfOCPBelow(major, minor int) {
    if clusterVersion == nil {
        Fail("Cluster version not initialized")
    }
    if !clusterVersion.AtLeast(major, minor) {
        Skip(fmt.Sprintf("Requires OCP %d.%d+, cluster is %s", major, minor, clusterVersion))
    }
}
```

### 2. Initialize in Suite (`integration-tests/backend/backend_suite_test.go`)
```go
var _ = BeforeSuite(func() {
    // ... existing setup code ...
    
    // Cache cluster version
    var err error
    clusterVersion, err = GetOCPVersion(context.Background(), k8sClient)
    Expect(err).NotTo(HaveOccurred())
    GinkgoWriter.Printf("Running tests against OCP %s\n", clusterVersion.String())
})
```

## Usage in Tests

### Basic pattern:
```go
It("feature requiring OCP 4.16+", func(ctx SpecContext) {
    SkipIfOCPBelow(4, 16)
    
    // Test code here - only runs on OCP 4.16+
})
```

### Multiple tests with same requirement:
```go
var _ = Describe("New Features in 4.16", func() {
    BeforeEach(func() {
        SkipIfOCPBelow(4, 16)
    })
    
    It("feature A", func(ctx SpecContext) {
        // Runs on 4.16+
    })
    
    It("feature B", func(ctx SpecContext) {
        // Runs on 4.16+
    })
})
```

### Mixed version requirements:
```go
var _ = Describe("FlowCollector", func() {
    It("basic test - no version requirement", func(ctx SpecContext) {
        // Runs on all OCP versions
    })
    
    It("advanced feature - 4.15+", func(ctx SpecContext) {
        SkipIfOCPBelow(4, 15)
        // Runs on 4.15+
    })
    
    It("newest feature - 4.17+", func(ctx SpecContext) {
        SkipIfOCPBelow(4, 17)
        // Runs on 4.17+
    })
})
```

## Running Tests

```bash
# Run all tests (version checks happen automatically at runtime)
cd integration-tests/backend
ginkgo -v

# From repository root
ginkgo -r integration-tests/backend/

# Output shows skipped tests with reason:
# S [SKIPPED] in 0.001 seconds
# Requires OCP 4.16+, cluster is 4.15
```

## Best Practices

1. **Put version check first** in test body for clarity
2. **Use BeforeEach** for multiple tests with same requirement
3. **Always specify why** - message shows required vs actual version
4. **Cache version** - retrieved once per suite, not per test

## Advanced Patterns

### Skip on specific versions only:
```go
It("workaround for 4.15 bug", func(ctx SpecContext) {
    if clusterVersion.Major == 4 && clusterVersion.Minor == 15 {
        Skip("Known issue in OCP 4.15 - fixed in 4.16")
    }
    // Test code
})
```

### Version-specific test logic:
```go
It("handles API changes", func(ctx SpecContext) {
    if clusterVersion.AtLeast(4, 16) {
        // Use new API
    } else {
        // Use legacy API
    }
})
```

### Skip above certain version:
```go
It("deprecated feature removed in 4.17", func(ctx SpecContext) {
    if clusterVersion.AtLeast(4, 17) {
        Skip("Feature removed in OCP 4.17+")
    }
    // Test deprecated feature
})
```

## How It Works

1. **BeforeSuite** queries the OpenShift cluster version API once
2. Version is **cached** in `clusterVersion` package variable
3. Each test calls `SkipIfOCPBelow(major, minor)` as needed
4. Ginkgo marks test as **SKIPPED** with clear reason if version too old
5. Test results show exactly which tests ran vs skipped and why

## Example Test Output

```
Running Suite: Backend Test Suite
===================================
Running tests against OCP 4.15

• [PASSED] [0.123 seconds]
FlowCollector basic test - no version requirement

S [SKIPPED] [0.001 seconds]
FlowCollector newest feature - 4.17+
  Requires OCP 4.17+, cluster is 4.15

• [PASSED] [2.456 seconds]
FlowCollector advanced feature - 4.15+

Ran 2 of 3 Specs in 2.580 seconds
SUCCESS! -- 2 Passed | 0 Failed | 0 Pending | 1 Skipped
```
