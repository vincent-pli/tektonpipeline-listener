/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ksvcClientset "github.com/knative/serving/pkg/client/clientset/versioned"

	eventbindingcontrollerImp "github.com/vincent-pli/tektonpipeline-listener/pkg/controller/eventbinding"
	templatecontrollerImp "github.com/vincent-pli/tektonpipeline-listener/pkg/controller/listenertemplate"
	listenerClientset "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/clientset/versioned"
	informers "github.com/vincent-pli/tektonpipeline-listener/pkg/generated/informers/externalversions"
	"github.com/vincent-pli/tektonpipeline-listener/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flagset := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(flagset)
	flagset.Set("logtostderr", "true")
	flagset.Set("v", "4")

	flag.Parse()
	klog.Info("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	listenerClient, err := listenerClientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building listener clientset: %s", err.Error())
	}

	ksvcClient, err := ksvcClientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building knative serving clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	listenerInformerFactory := informers.NewSharedInformerFactory(listenerClient, time.Second*30)

	templateController := templatecontrollerImp.NewController(kubeClient, listenerClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		listenerInformerFactory.Samplecontroller().V1alpha1().ListenerTemplates())

	eventbindingController := eventbindingcontrollerImp.NewController(kubeClient, listenerClient, ksvcClient
		kubeInformerFactory.Apps().V1().Deployments(),
		listenerInformerFactory.Samplecontroller().V1alpha1().EventBindings())
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	listenerInformerFactory.Start(stopCh)

	if err = templateController.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running templatecontroller: %s", err.Error())
	}

	if err = eventbindingController.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running eventbindiingcontroller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
