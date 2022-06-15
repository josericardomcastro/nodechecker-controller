package controllers

import (
	"context"
	nodecontrollerv1 "github.com/josericardomcastro/nodechecker-controller/pkg/apis/nodecontroller/v1"
	"github.com/josericardomcastro/nodechecker-controller/pkg/client"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"time"
)

func setSyncLabels(syncLabels map[string]string, executeResult bool, log *logrus.Entry) *corev1.Node {
	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}
	nodeUpdated, err := clientset.CoreV1().Nodes().Get(context.TODO(), os.Getenv("NODE_NAME"), metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	newNodeLabel := nodeUpdated.Labels
	for key, value := range syncLabels {
		_, found := newNodeLabel[key]
		if executeResult {
			if !found {
				newNodeLabel[key] = value
			}
		} else {
			if found {
				delete(newNodeLabel, key)
			}
		}
	}

	// Update node labels
	nodeUpdated.SetLabels(newNodeLabel)
	nodeClient := clientset.CoreV1().Nodes()
	_, err = nodeClient.Update(context.TODO(), nodeUpdated, metav1.UpdateOptions{})
	if err != nil {
		log.Warningf("Erro to update node label: %v", err.Error())
	} else {
		log.Info("Labels synced")
	}
	return nodeUpdated
}

func setNodeEvent(eventNodeName string, node *corev1.Node, nodeChecker nodecontrollerv1.NodeChecker, reason string, typeEvent string, message string, log *logrus.Entry) {
	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}

	eventFound, errNode := clientset.CoreV1().Events("default").Get(context.TODO(), eventNodeName, metav1.GetOptions{})

	if errNode != nil && errors.IsNotFound(errNode) {
		eventNode := &corev1.Event{}
		eventNode.SetName(eventNodeName)
		eventNode.SetNamespace("default")
		eventNode.InvolvedObject = corev1.ObjectReference{Kind: "Node", Name: node.Name, UID: types.UID(node.Name), APIVersion: node.APIVersion}
		eventNode.Related = &corev1.ObjectReference{Kind: "NodeChecker", Name: nodeChecker.Name, Namespace: nodeChecker.Namespace, APIVersion: nodeChecker.APIVersion}
		eventNode.Reason = reason
		eventNode.Message = message
		eventNode.ReportingController = "NodeCheckerController"
		eventNode.Source = corev1.EventSource{Component: "NodeCheckerController"}
		eventNode.Type = typeEvent
		eventNode.Count = 1
		eventNode.FirstTimestamp = metav1.Now()
		eventNode.LastTimestamp = metav1.Now()

		_, err = clientset.CoreV1().Events("default").Create(context.TODO(), eventNode, metav1.CreateOptions{})
		if err != nil {
			log.Warningf("Erro to create a node event: %s", err.Error())
		} else {
			log.Info("Node event created")
		}
	} else if err != nil {
		panic(err.Error())
	} else {
		eventFound.LastTimestamp = metav1.Time{time.Now()}
		eventFound.Message = message
		eventFound.Count = eventFound.Count + 1
		_, err = clientset.CoreV1().Events("default").Update(context.TODO(), eventFound, metav1.UpdateOptions{})
		if err != nil {
			log.Warningf("Erro to update a node event: %s", err.Error())
		} else {
			log.Info("Node event updated")
		}
	}
}

func selectNodes(matchLabel map[string]string) *corev1.NodeList {
	clientset, err := client.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabel}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:         500,
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), listOptions)
	if err != nil {
		panic(err.Error())
	}

	return nodes
}
