package main

import (
	"flag"
	"fmt"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/wardenlym/static-pod-controller/pkg/generated/clientset/versioned"
	informers "github.com/wardenlym/static-pod-controller/pkg/generated/informers/externalversions"
	"github.com/wardenlym/static-pod-controller/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	fmt.Println("static-pod-controller")
	flag.Parse()

	var err error
	var cfg *rest.Config
	if kubeconfig == "" {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Error get cluster kubeconfig: %s", err.Error())
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			klog.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	staticmacvlanClientSet, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	staticmacvlanInformerFactory := informers.NewSharedInformerFactory(staticmacvlanClientSet, time.Second*30)

	controller := NewController(kubeClient, staticmacvlanClientSet,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Pods(),
		staticmacvlanInformerFactory.Staticmacvlan().V1().StaticPods(),
		staticmacvlanInformerFactory.Staticmacvlan().V1().VLANSubnets())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	fmt.Println("informer start")
	kubeInformerFactory.Start(stopCh)
	staticmacvlanInformerFactory.Start(stopCh)
	fmt.Println("controller run")
	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
