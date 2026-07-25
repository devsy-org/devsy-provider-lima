package e2e

import (
	"testing"

	_ "github.com/devsy-org/devsy-provider-lima/e2e/tests/integration"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestRunE2ETests(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Devsy Lima Provider e2e suite")
}
