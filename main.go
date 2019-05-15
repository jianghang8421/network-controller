package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ehazlett/simplelog"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/cnrancher/network-controller/controller"
	clientset "github.com/cnrancher/network-controller/pkg/generated/clientset/versioned"
	informers "github.com/cnrancher/network-controller/pkg/generated/informers/externalversions"
	"github.com/cnrancher/network-controller/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	fmt.Println("network-controller")
	flag.Parse()

	var err error
	var cfg *rest.Config
	if kubeconfig == "" {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error get cluster kubeconfig: %s", err.Error())
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	dynamicKubeClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	macvlanClientSet, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	macvlanInformerFactory := informers.NewSharedInformerFactory(macvlanClientSet, time.Second*30)

	c := controller.NewController(
		kubeClient,
		dynamicKubeClient,
		macvlanClientSet,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Namespaces(),
		kubeInformerFactory.Core().V1().Pods(),
		macvlanInformerFactory.Macvlan().V1().MacvlanIPs(),
		macvlanInformerFactory.Macvlan().V1().MacvlanSubnets(),
		stopCh)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	fmt.Println("informer start")
	kubeInformerFactory.Start(stopCh)
	macvlanInformerFactory.Start(stopCh)
	fmt.Println("controller run")

	if err = c.Run(1, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {

	log.SetFormatter(&simplelog.SimpleFormatter{})
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
