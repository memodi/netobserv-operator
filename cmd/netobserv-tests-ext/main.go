package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	oteextension "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	oteginkgo "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	e2eframework "k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/klog/v2"
)

func main() {
	command, err := newOperatorTestCommand()
	if err != nil {
		klog.Fatal(err)
	}
	code := cli.Run(command)
	os.Exit(code)
}

func newOperatorTestCommand() (*cobra.Command, error) {
	registry, err := prepareOperatorTestsRegistry()
	if err != nil {
		return nil, err
	}

	root := &cobra.Command{
		Use:   "netobserv-tests-ext",
		Short: "NetObserv operator test extension for openshift-tests",
		Long: `Network Observability operator test extension binary that integrates
with the OpenShift tests framework for automated testing.`,
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)
	return root, nil
}

func prepareOperatorTestsRegistry() (*oteextension.Registry, error) {
	registry := oteextension.NewRegistry()

	// Create extension for NetObserv layered operator
	ext := oteextension.NewExtension("layered", "netobserv", "netobserv-operator")

	// All tests suite - includes everything
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/all",
		Description: "All Network Observability operator tests",
		Qualifiers: []string{
			`name.contains("[sig-netobserv]")`,
		},
	})

	// Parallel suite - non-serial tests
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/parallel",
		Description: "Network Observability operator tests that can run in parallel",
		Qualifiers: []string{
			`name.contains("[sig-netobserv]")`,
			`!name.contains("[Serial]")`,
			`!name.contains("[qe-only]")`,
		},
	})

	// Serial suite - tests that must run one at a time
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/serial",
		Description: "Network Observability operator tests that must run serially",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("Serial")`},
	})

	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/disruptive",
		Description: "Network Observability operator tests that must run serially",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("Disruptive")`,
		},
	})

	// Smoke suite - fast non-serial tests for quick validation
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/sanity",
		Description: "Network Observability operator sanity tests",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("Sanity")`,
		},
	})

	// tests with Kafka
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/kafka",
		Description: "Network Observability operator sanity tests",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("Kafka")`,
		},
	})

	// tests with Loki
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/loki",
		Description: "Network Observability operator sanity tests",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("Loki") && !name.contains("Kafka")`,
		},
	})

	// tests for virtualization.
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/VMs",
		Description: "Network Observability operator sanity tests",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("VMs")`,
		},
	})

	// tests for exporters.
	ext.AddSuite(oteextension.Suite{
		Name:        "layered/netobserv/exporters",
		Description: "Network Observability operator sanity tests",
		Parallelism: 1,
		Qualifiers: []string{
			`name.contains("[sig-netobserv]") && name.contains("exporters")`,
		},
	})

	// Build test specs from the Ginkgo suite
	// Tests are registered via init() in dependencymagnet.go
	specs, err := oteginkgo.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		return nil, fmt.Errorf("couldn't build extension test specs from ginkgo: %w", err)
	}

	// Initialize framework before running tests.
	// The extension framework clears BeforeSuite/AfterSuite nodes and never calls
	// TestBackend(), so exutil.WithCleanup() (which sets the unexported testsStarted
	// flag) is never invoked. Without it, NewCLI's BeforeEach panics at
	// requiresTestStart(). We call WithCleanup here to set that flag and perform
	// all necessary framework initialization before any test BeforeEach runs.
	specs.AddBeforeAll(func() {
		exutil.WithCleanup(func() {
			flag.Parse()

			kubeconfig := os.Getenv("KUBECONFIG")
			if kubeconfig == "" {
				panic("KUBECONFIG environment variable must be set")
			}
			exutil.TestContext.KubeConfig = kubeconfig

			e2eframework.AfterReadingAllFlags(exutil.TestContext)
		})
	})

	ext.AddSpecs(specs)
	registry.Register(ext)

	return registry, nil
}
