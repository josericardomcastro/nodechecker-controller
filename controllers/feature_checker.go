package controllers

import (
	"context"
	"fmt"
	"github.com/josericardomcastro/nodechecker-controller/config"
	nodecontrollerv1 "github.com/josericardomcastro/nodechecker-controller/pkg/apis/nodecontroller/v1"
	"github.com/josericardomcastro/nodechecker-controller/pkg/client"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
)

type FeatureChecker struct {
	node          *corev1.Node
	sourceNodes   *corev1.NodeList
	nodeChecker   nodecontrollerv1.NodeChecker
	eventNodeName string

	log     *logrus.Entry
	metrics *config.ServerMetrics
}

func CheckFeature(nc nodecontrollerv1.NodeChecker, log *logrus.Entry, metrics *config.ServerMetrics) {

	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}

	checker := &FeatureChecker{
		nodeChecker: nc,
		log:         log,
		metrics:     metrics,
	}

	checker.node, err = clientset.CoreV1().Nodes().Get(context.TODO(), os.Getenv("NODE_NAME"), metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	checker.eventNodeName = fmt.Sprintf("%s-%s", checker.node.Name, nc.Name)
	checker.sourceNodes = selectNodes(nc.Spec.SourceNodes.MatchLabels)

	for _, node := range checker.sourceNodes.Items {
		if checker.node.Name == node.Name {
			checker.executeCheck()
		}
	}

	if len(checker.sourceNodes.Items) == 0 {
		log.Infof("No nodes for matchLabel: %v", nc.Spec.SourceNodes.MatchLabels)
	}
}

func (a *FeatureChecker) executeCheck() {
	a.log = a.log.WithFields(logrus.Fields{"Rule": a.nodeChecker.Name, "Type": "FeatureChecker"})

	result := true
	cmd := exec.Command("sh", "-c", a.nodeChecker.Spec.FeatureCheck.Command)
	err := cmd.Run()
	if err != nil {
		a.log.Errorf("Feature check error: %s", err)
		result = false
		a.metrics.IncFeatureCheckError(a.node.Name, a.nodeChecker.Name, a.nodeChecker.Spec.Schedule)
		message := fmt.Sprintf("Failed to check feature %s - schedule %v", a.nodeChecker.Name, a.nodeChecker.Spec.Schedule)
		setNodeEvent(a.eventNodeName, a.node, a.nodeChecker, "FeatureCheckFailed", corev1.EventTypeWarning, message, a.log)
	} else {
		a.log.Info("Feature check success")
		a.metrics.SetFeatureCheckError(a.node.Name, a.nodeChecker.Name, a.nodeChecker.Spec.Schedule, 0)
	}

	if len(a.nodeChecker.Spec.FeatureCheck.SyncLabels) > 0 {
		setSyncLabels(a.nodeChecker.Spec.FeatureCheck.SyncLabels, result, a.log)
	}
}
