package client

import (
	"flag"
	"os"
	"path/filepath"

	nodecontrollerClientset "github.com/josericardomcastro/nodechecker-controller/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig *string

func GetDefaultClient() (*kubernetes.Clientset, error) {

	var config *rest.Config

	if os.Getenv("CLUSTER_CONFIG") == "local" {
		config = getLocalConfig()
	} else {
		config = getClusterConfig()
	}

	return kubernetes.NewForConfig(config)
}

func GetNodeControllerClient() (*nodecontrollerClientset.Clientset, error) {

	var config *rest.Config

	if os.Getenv("CLUSTER_CONFIG") == "local" {
		config = getLocalConfig()
	} else {
		config = getClusterConfig()
	}

	return nodecontrollerClientset.NewForConfig(config)
}

func GetRestConfig() (*rest.Config, error) {
	var config *rest.Config

	if os.Getenv("CLUSTER_CONFIG") == "local" {
		config = getLocalConfig()
	} else {
		config = getClusterConfig()
	}
	return config, nil
}

func getClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return config
}

func getLocalConfig() *rest.Config {

	if kubeconfig == nil {
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
	}

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//log.Println("Accessing local config")

	return config
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
