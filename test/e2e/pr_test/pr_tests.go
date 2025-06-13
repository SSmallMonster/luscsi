package E2eTest

import (
	"context"
	"fmt"
	"github.com/luskits/luscsi/test/e2e/utils"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var _ = ginkgo.Describe("pr test ", ginkgo.Ordered, ginkgo.Label("pr-e2e"), func() {

	ginkgo.Context("Configure the base environment", func() {
		err := utils.K8sReadyCheck()
		gomega.Expect(err).To(gomega.BeNil())
		err = utils.InstallLuscsi()
		gomega.Expect(err).To(gomega.BeNil())

	})
	ginkgo.Context("Checking Component Status", func() {
		logrus.Infof("waiting for luscsi ready")
		time.Sleep(180 * time.Second)

		config, err := clientcmd.BuildConfigFromFlags("", "/home/github-runner/.kube/config")
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err := kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		podList, _ := kubeClient.CoreV1().Pods("luscsi").List(context.TODO(), metav1.ListOptions{})
		deploymentList, _ := kubeClient.AppsV1().Deployments("luscsi").List(context.TODO(), metav1.ListOptions{})

		for _, pod := range podList.Items {
			for {
				onePod, err := kubeClient.CoreV1().Pods("luscsi").Get(context.Background(), pod.Name, metav1.GetOptions{})
				// check pod status should be running
				onePodStatus := string(onePod.Status.Phase)
				ginkgo.GinkgoWriter.Printf("***** wait for pod[%s] status: %s\n", onePod.Name, onePodStatus)
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
				if onePodStatus == "Running" || onePodStatus == "Failed" || onePodStatus == "Succeeded" || onePodStatus == "Pending" {
					ginkgo.It(fmt.Sprintf("* wait for pod[%s] status: %s\n", pod.Name, onePodStatus), func() {
						// gomega.Expect(onePodStatus).To(gomega.Equal("Running"))
						gomega.Expect(onePodStatus).To(gomega.Or(gomega.MatchRegexp("Running"), gomega.MatchRegexp("Succeeded")))
					})
					break
				}
				time.Sleep(10 * time.Second)
			}
		}
		ginkgo.It(fmt.Sprintf("ns: luscsi deployment should be ready"), func() {
			for _, dm := range deploymentList.Items {
				fmt.Println(dm.Name, "======>", dm.Status.Replicas)
				gomega.Expect(dm.Status.ReadyReplicas).To(gomega.Equal(dm.Status.AvailableReplicas))
			}
		})
	})

})
