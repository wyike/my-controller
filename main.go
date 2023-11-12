package main

import (
	"flag"
	"fmt"
	"time"
	
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/Users/yikew/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s building config from flags\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s, getting inclusterconfig\n", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}

	ch := make(chan struct{})
	factory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	c := newController(clientset, factory.Apps().V1().Deployments())
	factory.Start(ch)
	c.run(ch)
}
