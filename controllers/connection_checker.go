package controllers

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/josericardomcastro/nodechecker-controller/config"

	nodecontrollerv1 "github.com/josericardomcastro/nodechecker-controller/pkg/apis/nodecontroller/v1"
	"github.com/josericardomcastro/nodechecker-controller/pkg/client"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConnectionChecker struct {
	node          *corev1.Node
	sourceNodes   *corev1.NodeList
	nodeChecker   nodecontrollerv1.NodeChecker
	eventNodeName string

	log     *logrus.Entry
	metrics *config.ServerMetrics
}

func CheckConnection(nc nodecontrollerv1.NodeChecker, log *logrus.Entry, metrics *config.ServerMetrics) {
	connectionCheck := &ConnectionChecker{
		nodeChecker: nc,
		log:         log,
		metrics:     metrics,
	}

	connectionCheck.log = connectionCheck.log.WithFields(logrus.Fields{"Type": "ConnectionCheck"})

	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}

	connectionCheck.node, err = clientset.CoreV1().Nodes().Get(context.TODO(), os.Getenv("NODE_NAME"), metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	connectionCheck.sourceNodes = selectNodes(nc.Spec.SourceNodes.MatchLabels)

	for _, node := range connectionCheck.sourceNodes.Items {
		if connectionCheck.node.Name == node.Name {
			connectionCheck.executeCheck()
		}
	}
}

func (a *ConnectionChecker) executeCheck() {
	for _, i := range a.nodeChecker.Spec.ConnectionCheck.External {
		a.log = a.log.WithFields(logrus.Fields{"Rule": a.nodeChecker.Name, "Name": i.Name, "Target": "External", "Host": i.Host, "Port": i.Port, "Protocolo": i.Protocol})
		a.log.Info("Connection check start")
		a.eventNodeName = fmt.Sprintf("%s-%s-%s", a.node.Name, a.nodeChecker.Name, i.Name)
		a.checkTargetExternal(i)
	}

	for _, i := range a.nodeChecker.Spec.ConnectionCheck.ClusterNodes {
		a.log = a.log.WithFields(logrus.Fields{"Rule": a.nodeChecker.Name, "Name": i.Name, "Target": "ClusterNode", "Port": i.Port, "Protocolo": i.Protocol})
		a.log.Info("Connection check start")
		a.eventNodeName = fmt.Sprintf("%s-%s-%s", a.node.Name, a.nodeChecker.Name, i.Name)
		a.checkTargetClusterNodes(i)
	}

	for _, i := range a.nodeChecker.Spec.ConnectionCheck.ClusterEndpoints {
		a.log = a.log.WithFields(logrus.Fields{"Rule": a.nodeChecker.Name, "Name": i.Name, "Target": "ClusterEndpoint", "Endpoint": fmt.Sprintf("%s/%s", i.Endpoint.Namespace, i.Endpoint.Name)})
		a.log.Info("Connection check start")
		a.eventNodeName = fmt.Sprintf("%s-%s-%s", a.node.Name, a.nodeChecker.Name, i.Name)
		a.checkTargetClusterEndpoints(i)
	}
}

func (a *ConnectionChecker) checkTargetExternal(t nodecontrollerv1.TargetExternal) {
	result := true
	c, err := net.DialTimeout(strings.ToLower(t.Protocol), fmt.Sprintf("%v:%v", t.Host, t.Port), 5*time.Second)
	if err != nil {
		a.log.Errorf("Result: %s", err.Error())
		a.metrics.IncConnectionCheckError(a.node.Name, a.nodeChecker.Name, "External", fmt.Sprintf("%v:%v", t.Host, t.Port), a.nodeChecker.Spec.Schedule)
		result = false
		message := fmt.Sprintf("Failed to connect to %s:%d with protocol %s - schedule %v", t.Host, t.Port, t.Protocol, a.nodeChecker.Spec.Schedule)
		setNodeEvent(a.eventNodeName, a.node, a.nodeChecker, "ConnectionCheckFailed", corev1.EventTypeWarning, message, a.log)
	} else {
		a.metrics.SetConnectionCheckError(a.node.Name, a.nodeChecker.Name, "External", fmt.Sprintf("%v:%v", t.Host, t.Port), a.nodeChecker.Spec.Schedule, 0)
		a.log.Info("Connection check success")
		c.Close()
	}

	if len(t.SyncLabels) != 0 {
		setSyncLabels(t.SyncLabels, result, a.log)
	}
}

func (a *ConnectionChecker) checkTargetClusterNodes(t nodecontrollerv1.TargetClusterNodes) {
	result := true
	nodeList := selectNodes(t.MatchLabels)

	for _, node := range nodeList.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				host := address.Address

				c, err := net.DialTimeout(strings.ToLower(t.Protocol), fmt.Sprintf("%v:%v", host, t.Port), 5*time.Second)
				if err != nil {
					a.log.Errorf("Result: %s", err.Error())
					a.metrics.IncConnectionCheckError(a.node.Name, a.nodeChecker.Name, "ClusterNodes", fmt.Sprintf("%v:%v", host, t.Port), a.nodeChecker.Spec.Schedule)
					result = false
					message := fmt.Sprintf("Failed to connect to node %s - %s:%d with protocol %s - schedule %v", node.Name, host, t.Port, t.Protocol, a.nodeChecker.Spec.Schedule)
					setNodeEvent(a.eventNodeName, a.node, a.nodeChecker, "ConnectionCheckFailed", corev1.EventTypeWarning, message, a.log)
				} else {
					a.metrics.SetConnectionCheckError(a.node.Name, a.nodeChecker.Name, "ClusterNodes", fmt.Sprintf("%v:%v", host, t.Port), a.nodeChecker.Spec.Schedule, 0)
					a.log.Info("Connection check success")
					c.Close()
				}
			}
		}
	}

	if len(t.SyncLabels) > 0 && len(nodeList.Items) > 0 {
		setSyncLabels(t.SyncLabels, result, a.log)
	}
}

func (a *ConnectionChecker) checkTargetClusterEndpoints(t nodecontrollerv1.TargetClusterEndpoints) {
	result := true
	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}

	endpoint, err := clientset.CoreV1().Endpoints(t.Endpoint.Namespace).Get(context.TODO(), t.Endpoint.Name, metav1.GetOptions{})
	if err != nil {
		a.log.Errorf("Result: %v", err.Error())
	}

	for _, subset := range endpoint.Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				endpointProtocol := "TCP"
				if port.Protocol == corev1.ProtocolUDP {
					endpointProtocol = "UDP"
				}

				c, err := net.DialTimeout(strings.ToLower(endpointProtocol), fmt.Sprintf("%v:%v", address.IP, port.Port), 5*time.Second)
				if err != nil {
					a.log.Errorf("Result: %s", err.Error())
					result = false
					a.metrics.IncConnectionCheckError(a.node.Name, a.nodeChecker.Name, "ClusterEndpoint", fmt.Sprintf("%v:%v", t.Endpoint.Name, t.Endpoint.Namespace), a.nodeChecker.Spec.Schedule)
					message := fmt.Sprintf("Failed to connect to endpoint %s/%s - %s:%d with protocol %s - schedule %v", endpoint.Namespace, endpoint.Name, address.IP, port.Port, endpointProtocol, a.nodeChecker.Spec.Schedule)
					setNodeEvent(a.eventNodeName, a.node, a.nodeChecker, "ConnectionCheckFailed", corev1.EventTypeWarning, message, a.log)
				} else {
					a.metrics.SetConnectionCheckError(a.node.Name, a.nodeChecker.Name, "ClusterEndpoint", fmt.Sprintf("%v:%v", t.Endpoint.Name, t.Endpoint.Namespace), a.nodeChecker.Spec.Schedule, 0)
					a.log.Info("Connection check success")
					c.Close()
				}
			}
		}
	}

	if len(t.SyncLabels) != 0 && len(endpoint.Subsets) > 0 {
		setSyncLabels(t.SyncLabels, result, a.log)
	}
}
