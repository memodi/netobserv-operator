package e2etests

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	exutil "github.com/openshift/origin/test/extended/util"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"

	"golang.org/x/mod/semver"
)

var clusterVersion string

func GetOCPVersion(oc *exutil.CLI) (string, error) {
	if clusterVersion != "" {
		return clusterVersion, nil
	}

	var err error
	_ , clusterVersion, err = compat_otp.GetClusterVersion(oc)
	clusterVersion = semver.Canonical("v"+clusterVersion)
	fmt.Printf("Detected OCP version: %s\n", clusterVersion)
	return clusterVersion, err
}

// SkipIfOCPBelow skips test if cluster version is below requirement
func SkipIfOCPBelow(requiredVersion string) {
	if clusterVersion == "" {
		ginkgo.Fail("Cluster version not initialized")
	}

	requiredVersion = semver.Canonical(requiredVersion)
	if !semver.IsValid(requiredVersion) {
		ginkgo.Fail("Requested cluster version is invalid")
	}

	if semver.Compare(requiredVersion, clusterVersion) == -1 {
		ginkgo.Skip(fmt.Sprintf("Requires OCP %s+, cluster is %s", requiredVersion, clusterVersion))
	}
}
