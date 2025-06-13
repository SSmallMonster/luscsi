package utils

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func Int32Ptr(i int32) *int32 { return &i }

func BoolPter(i bool) *bool { return &i }

func RunInLinux(cmd string) (string, error) {
	result, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logrus.Printf("ERROR:%+v ", err)
	}
	return string(result), err
}

func InstallLuscsi() error {
	logrus.Infof("helm install luscsi")
	_, err := RunInLinux("helm install luscsi -n luscsi --create-namespace ../../deploy/luscsi/")
	return err
}

func K8sReadyCheck() error {

	err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		output, err := RunInLinux("kubectl get pod -A | grep -v Running | wc -l")
		if err != nil {
			logrus.WithError(err).Warn("Failed to execute command")
			return false, nil
		}

		cleanedOutput := strings.TrimSpace(output)
		count, parseErr := strconv.Atoi(cleanedOutput)
		if parseErr != nil {
			logrus.Warnf("Failed to parse output: %q", cleanedOutput)
			return false, nil
		}

		if count == 1 {
			logrus.Info("k8s ready")
			return true, nil
		}

		logrus.Debugf("Waiting for k8s readiness, current not-ready pod count: %d", count)
		return false, nil
	})

	if err != nil {
		logrus.WithError(err).Error("Kubernetes readiness check timed out")
	}
	return err

}

func CheckingComponentStatus() {

}
