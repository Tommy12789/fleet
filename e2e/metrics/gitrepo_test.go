package metrics_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rancher/fleet/e2e/metrics"
	"github.com/rancher/fleet/e2e/testenv"
	"github.com/rancher/fleet/e2e/testenv/kubectl"
)

var _ = Describe("GitRepo Metrics", Label("gitrepo"), func() {

	var (
		// kw is the kubectl command for namespace the workload is deployed to
		kw        kubectl.Command
		namespace string
		objName   = "metrics"
		branch    = "master"
	)

	BeforeEach(func() {
		k = env.Kubectl.Namespace(env.Namespace)
		namespace = testenv.NewNamespaceName(
			objName,
			rand.New(rand.NewSource(time.Now().UnixNano())),
		)
		kw = k.Namespace(namespace)

		out, err := k.Create("ns", namespace)
		Expect(err).ToNot(HaveOccurred(), out)

		err = testenv.CreateGitRepo(
			kw,
			namespace,
			objName,
			branch,
			shard,
			"simple-manifest",
		)
		Expect(err).ToNot(HaveOccurred())

		DeferCleanup(func() {
			out, err = k.Delete("ns", namespace)
			Expect(err).ToNot(HaveOccurred(), out)
		})
	})

	When("testing GitRepo metrics", func() {
		gitrepoMetricNames := []string{
			"fleet_gitrepo_desired_ready_clusters",
			"fleet_gitrepo_ready_clusters",
			"fleet_gitrepo_resources_desired_ready",
			"fleet_gitrepo_resources_missing",
			"fleet_gitrepo_resources_modified",
			"fleet_gitrepo_resources_not_ready",
			"fleet_gitrepo_resources_orphaned",
			"fleet_gitrepo_resources_ready",
			"fleet_gitrepo_resources_unknown",
			"fleet_gitrepo_resources_wait_applied",
		}

		It("should have exactly one metric of each type for the gitrepo", func() {
			Eventually(func() error {
				metrics, err := et.Get()
				Expect(err).ToNot(HaveOccurred())
				for _, metricName := range gitrepoMetricNames {
					metric, err := et.FindOneMetric(
						metrics,
						metricName,
						map[string]string{
							"name":      objName,
							"namespace": namespace,
						},
					)
					if err != nil {
						return err
					}
					Expect(metric.Gauge.GetValue()).To(Equal(float64(0)))
				}
				return nil
			}).ShouldNot(HaveOccurred())
		})

		When("the GitRepo is changed", func() {
			It("it should not duplicate metrics", func() {
				out, err := kw.Patch(
					"gitrepo", objName,
					"--type=json",
					"-p", `[{"op": "replace", "path": "/spec/paths", "value": ["simple-chart"]}]`,
				)
				Expect(err).ToNot(HaveOccurred(), out)
				Expect(out).To(ContainSubstring("gitrepo.fleet.cattle.io/metrics patched"))

				// Wait for it to be changed and fetched.
				Eventually(func() (string, error) {
					return kw.Get("gitrepo", objName, "-o", "jsonpath={.status.commit}")
				}).ShouldNot(BeEmpty())

				var metric *metrics.Metric
				// Expect still no metrics to be duplicated.
				Eventually(func() error {
					metrics, err := et.Get()
					Expect(err).ToNot(HaveOccurred())
					for _, metricName := range gitrepoMetricNames {
						metric, err = et.FindOneMetric(
							metrics,
							metricName,
							map[string]string{
								"name":      objName,
								"namespace": namespace,
							},
						)
						if err != nil {
							return err
						}
						if metric.LabelValue("paths") != "simple-chart" {
							return fmt.Errorf("path for metric %s unchanged", metricName)
						}
					}
					return nil
				}).ShouldNot(HaveOccurred())
			})

			It("should not keep metrics if GitRepo is deleted", Label("gitrepo-delete"), func() {
				out, err := kw.Delete("gitrepo", objName)
				Expect(err).ToNot(HaveOccurred(), out)

				Eventually(func() error {
					metrics, err := et.Get()
					Expect(err).ToNot(HaveOccurred())
					for _, metricName := range gitrepoMetricNames {
						_, err := et.FindOneMetric(
							metrics,
							metricName,
							map[string]string{
								"name":      objName,
								"namespace": namespace,
							},
						)
						if err == nil {
							return fmt.Errorf("metric %s found", metricName)
						}
					}
					return nil
				}).ShouldNot(HaveOccurred())
			})
		})
	})
})
