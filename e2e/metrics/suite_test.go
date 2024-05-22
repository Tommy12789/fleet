package metrics_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rancher/fleet/e2e/metrics"
	"github.com/rancher/fleet/e2e/testenv"
	"github.com/rancher/fleet/e2e/testenv/kubectl"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite for metrics")
}

var (
	env *testenv.Env
	// k is the kubectl command for the cluster registration namespace
	k     kubectl.Command
	et    metrics.ExporterTest
	shard string
)

type ServiceData struct {
	Name           string
	Port           int64
	IsDefaultShard bool
	Shard          string
}

// setupLoadBalancer creates a load balancer service for the fleet controller.
// If shard is empty, it creates a service for the default (unsharded)
// controller.
func setupLoadBalancer(shard string) (metricsURL string) {
	rs := rand.NewSource(time.Now().UnixNano())
	port := rs.Int63()%1000 + 30000
	loadBalancerName := testenv.AddRandomSuffix("fleetcontroller", rs)

	ks := k.Namespace("cattle-fleet-system")
	err := testenv.ApplyTemplate(
		ks,
		testenv.AssetPath("metrics/fleetcontroller_service.yaml"),
		ServiceData{
			Name:           loadBalancerName,
			Port:           port,
			IsDefaultShard: shard == "",
			Shard:          shard,
		},
	)
	Expect(err).ToNot(HaveOccurred())

	Eventually(func() (string, error) {
		ip, err := ks.Get(
			"service", loadBalancerName,
			"-o", "jsonpath={.status.loadBalancer.ingress[0].ip}",
		)
		metricsURL = fmt.Sprintf("http://%s:%d/metrics", ip, port)
		return ip, err
	}).ShouldNot(BeEmpty())

	DeferCleanup(func() {
		ks := k.Namespace("cattle-fleet-system")
		_, _ = ks.Delete("service", loadBalancerName)
	})

	return metricsURL
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(testenv.Timeout)
	SetDefaultEventuallyPollingInterval(time.Second)
	testenv.SetRoot("../..")

	if os.Getenv("SHARD") != "" {
		shard = os.Getenv("SHARD")
	}

	// Enable passing the metrics URL via environment solely for debugging
	// purposes, e.g. when a fleetcontroller is run outside the cluster. This is
	// not intended for regular use.
	var metricsURL string
	if os.Getenv("METRICS_URL") != "" {
		metricsURL = os.Getenv("METRICS_URL")
	} else {
		metricsURL = setupLoadBalancer(shard)
	}
	et = metrics.NewExporterTest(metricsURL)

	env = testenv.New()
})
