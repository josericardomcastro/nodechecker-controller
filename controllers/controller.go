package controllers

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/josericardomcastro/nodechecker-controller/config"

	nodecontrollerClientset "github.com/josericardomcastro/nodechecker-controller/pkg/generated/clientset/versioned"

	nodecontrollerv1 "github.com/josericardomcastro/nodechecker-controller/pkg/apis/nodecontroller/v1"
	"github.com/josericardomcastro/nodechecker-controller/pkg/client"
	informers "github.com/josericardomcastro/nodechecker-controller/pkg/generated/informers/externalversions"
	nodecontrollerInformers "github.com/josericardomcastro/nodechecker-controller/pkg/generated/informers/externalversions/nodecontroller/v1"
	nodecontrollerListers "github.com/josericardomcastro/nodechecker-controller/pkg/generated/listers/nodecontroller/v1"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
)

type NodeCheckerController struct {
	clientset nodecontrollerClientset.Interface

	lister  nodecontrollerListers.NodeCheckerLister
	synced  cache.InformerSynced
	log     *log.Entry
	metrics *config.ServerMetrics

	listCronjob map[string]cron.EntryID
	cronjob     *cron.Cron
}

func StartNodeCheckerController(metrics *config.ServerMetrics) {
	var wg sync.WaitGroup

	nodecontrollerClientset, err := client.GetNodeControllerClient()
	if err != nil {
		log.Errorf("Error to create a kubernetes clientset: %v", err)
	}

	log := log.WithFields(log.Fields{"Node": os.Getenv("NODE_NAME")})

	stopCh := make(chan struct{})
	defer close(stopCh)

	kubeInformerFactory := informers.NewSharedInformerFactory(nodecontrollerClientset, time.Second*30)
	controller := newNodeCheckerController(nodecontrollerClientset, kubeInformerFactory.Nodecontroller().V1().NodeCheckers(), log, metrics)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)

	if err = controller.run(stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
	select {}

	wg.Add(1)
	wg.Wait()
}

func (n *NodeCheckerController) run(stopCh <-chan struct{}) error {
	n.log.Infoln("Starting NodeChecker controller")

	// wait for the initial synchronization of the local cache.
	if ok := cache.WaitForCacheSync(stopCh, n.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	return nil
}

func newNodeCheckerController(nodecontrollerClientset nodecontrollerClientset.Interface, nodecontrollerInformer nodecontrollerInformers.NodeCheckerInformer, log *log.Entry, metrics *config.ServerMetrics) *NodeCheckerController {
	c := &NodeCheckerController{
		clientset:   nodecontrollerClientset,
		lister:      nodecontrollerInformer.Lister(),
		synced:      nodecontrollerInformer.Informer().HasSynced,
		listCronjob: make(map[string]cron.EntryID),
		log:         log,
		metrics:     metrics,
		cronjob:     cron.New(),
	}

	c.cronjob.Start()

	nodecontrollerInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.addHandle,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: func(old, new interface{}) {
				newObj := new.(*nodecontrollerv1.NodeChecker)
				oldObj := old.(*nodecontrollerv1.NodeChecker)
				if newObj.ResourceVersion == oldObj.ResourceVersion {
					// Periodic resync will send update events for all known NodeChecker.
					// Two different versions of the same NodeChecker will always have different RVs.
					return
				}
				c.updateHandle(oldObj, newObj)
			},
			// Called on resource deletion.
			DeleteFunc: c.deleteHandle,
		},
	)
	return c
}

func (n *NodeCheckerController) addHandle(obj interface{}) {
	nodeChecker := obj.(*nodecontrollerv1.NodeChecker)

	// Create new nodechecker rules
	n.createCronjobCheck(nodeChecker)
}

func (n *NodeCheckerController) updateHandle(old, new interface{}) {
	oldNodeChecker := old.(*nodecontrollerv1.NodeChecker)
	nodeChecker := new.(*nodecontrollerv1.NodeChecker)
	n.log.Infof("NodeChecker updated: %s/%s - %s", nodeChecker.Namespace, nodeChecker.Name, oldNodeChecker.Name)

	// Remove old rules from modified nodeCheker
	n.removeCronjobCheck(oldNodeChecker, false)
	// Create new rules
	n.createCronjobCheck(nodeChecker)
}

func (n *NodeCheckerController) deleteHandle(obj interface{}) {
	nodeChecker := obj.(*nodecontrollerv1.NodeChecker)
	n.removeCronjobCheck(nodeChecker, true)
	n.log.Infof("NodeChecker deleted: %s/%s", nodeChecker.Namespace, nodeChecker.Name)
}

func (n *NodeCheckerController) createCronjobCheck(nc *nodecontrollerv1.NodeChecker) {
	cronjobName := nc.Name
	n.log = n.log.WithFields(log.Fields{"Name": nc.Name, "Schedule": nc.Spec.Schedule, "Type": ""})
	if _, ok := n.listCronjob[cronjobName]; !ok {
		entryId, err := n.startCronjobRule(*nc)
		if err != nil {
			n.log.Warning("Unable to create cronjob")
		}
		n.listCronjob[cronjobName] = entryId
		n.log.Info("Create cronjob rule")
	}
}

func (n *NodeCheckerController) removeCronjobCheck(nc *nodecontrollerv1.NodeChecker, removeLabels bool) {
	cronjobName := nc.Name
	n.cronjob.Remove(n.listCronjob[cronjobName])
	delete(n.listCronjob, cronjobName)
	n.log.Infof("Remove cronjob entryId=%v %s", n.listCronjob[cronjobName], cronjobName)

	// Remove node labels from deleted rules
	if removeLabels {
		if nc.Spec.Type == "connection" {
			for _, i := range nc.Spec.ConnectionCheck.External {
				setSyncLabels(i.SyncLabels, false, n.log)
			}

			for _, i := range nc.Spec.ConnectionCheck.ClusterNodes {
				setSyncLabels(i.SyncLabels, false, n.log)
			}

			for _, i := range nc.Spec.ConnectionCheck.ClusterEndpoints {
				setSyncLabels(i.SyncLabels, false, n.log)
			}
		} else {
			setSyncLabels(nc.Spec.FeatureCheck.SyncLabels, false, n.log)
		}
		n.log.Infof("Removed sync labels")
	}
}

func (n *NodeCheckerController) startCronjobRule(nc nodecontrollerv1.NodeChecker) (cron.EntryID, error) {
	return n.cronjob.AddFunc(nc.Spec.Schedule, func() {
		if nc.Spec.Type == "connection" {
			CheckConnection(nc, n.log, n.metrics)
		}
		if nc.Spec.Type == "feature" {
			CheckFeature(nc, n.log, n.metrics)
		}
	})
}
