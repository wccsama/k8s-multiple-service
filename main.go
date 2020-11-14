package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"time"

// 	clientset "github.com/k8s-multiple-service/pkg/client/clientset/versioned"
// 	informers "github.com/k8s-multiple-service/pkg/client/informers/externalversions"
// 	controller "github.com/k8s-multiple-service/pkg/controller"
// 	"github.com/spf13/pflag"
// 	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
// 	"k8s.io/apimachinery/pkg/util/wait"
// 	k8sinformers "k8s.io/client-go/informers"
// 	"k8s.io/client-go/kubernetes"
// 	restclient "k8s.io/client-go/rest"
// 	"k8s.io/client-go/tools/clientcmd"
// 	"k8s.io/klog"
// )

// var (
// 	controllerResyncPeriod int32
// 	informerDefaultResync  int32
// 	configQPS              float32
// 	configBurst            int
// 	workers                int
// 	fromKubeConfig         string
// 	toKubeConfig           string
// 	leaderElect            bool
// 	createCRD              bool
// )

// func addFlags(fs *pflag.FlagSet) {
// 	fs.Int32Var(&controllerResyncPeriod, "controllerResyncPeriod", 60, "controller resyncPeriod unit second")
// 	fs.Int32Var(&informerDefaultResync, "informerDefaultResync", 30, "informer resyncPeriod unit second")

// 	fs.Float32Var(&configQPS, "configQPS", 20, "api configQPS")
// 	fs.IntVar(&configBurst, "configBurst", 50, "api configBurst")

// 	fs.IntVar(&workers, "workers", 1, "wokers' num in controller")
// 	fs.StringVar(&fromKubeConfig, "fromKubeConfig", "/root/.kube/config", "export service from this cluster")
// 	fs.StringVar(&toKubeConfig, "toKubeConfig", "/root/.kube/config", "export service to this cluster")
// 	// TODO: multiple
// 	fs.BoolVar(&leaderElect, "leaderElect", false, "open leaderElection or not")
// 	fs.BoolVar(&createCRD, "create-crd", true, "Create ServiceExport CRD if it does not exist")

// }

// func main() {
// 	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
// 	addFlags(pflag.CommandLine)
// 	pflag.Parse()

// 	if err := checkCommandParameters(); err != nil {
// 		klog.Fatalf("checkCommandParameters err: %s", err.Error())
// 	}

// 	fromClientConfig, err := newK8SClientFromKubeConfig(fromKubeConfig)
// 	if err != nil {
// 		klog.Fatalf("Error getting kubeconfig: %s", err.Error())
// 	}

// 	toClientConfig, err := newK8SClientFromKubeConfig(toKubeConfig)
// 	if err != nil {
// 		klog.Fatalf("Error getting kubeconfig: %s", err.Error())
// 	}

// 	fromKubeClient, err := kubernetes.NewForConfig(fromClientConfig)
// 	if err != nil {
// 		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
// 	}

// 	toKubeClient, err := kubernetes.NewForConfig(toClientConfig)
// 	if err != nil {
// 		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
// 	}

// 	// for reconcile serviceExport
// 	serviceExportClient, err := clientset.NewForConfig(fromClientConfig)
// 	if err != nil {
// 		klog.Fatalf("Error building cronHpa clientset: %s", err.Error())
// 	}

// 	// for ensuring crd in export service to cluster
// 	extensionsClient, err := apiextensionsclient.NewForConfig(toClientConfig)
// 	if err != nil {
// 		klog.Fatalf("Error instantiating apiextensions client: %s", err.Error())
// 	}

// 	// from cluster
// 	informerServiceExportFactory := informers.NewSharedInformerFactory(serviceExportClient, time.Duration(informerDefaultResync)*time.Second)
// 	fromInformerFactory := k8sinformers.NewSharedInformerFactory(fromKubeClient, time.Duration(informerDefaultResync)*time.Second)
// 	toInformerFactory := k8sinformers.NewSharedInformerFactory(toKubeClient, time.Duration(informerDefaultResync)*time.Second)

// 	serviceExportInformer := informerServiceExportFactory.Wcc().V1().ServiceExports()

// 	msController, err := controller.NewMSController(fromKubeClient,
// 		toKubeClient,

// 		fromInformerFactory,
// 		toInformerFactory,
// 		serviceExportInformer,
// 		time.Duration(controllerResyncPeriod),
// 	)

// 	if err != nil {
// 		klog.Fatalf("Error NewMSController err: %s", err.Error())
// 	}

// 	run := func() {
// 		if createCRD {
// 			wait.PollUntil(time.Second*5, func() (bool, error) { return controller.EnsureCRDCreated(context.Background(), extensionsClient) }, context.Background().Done())
// 		}

// 		stopCh := make(chan struct{})
// 		go informerServiceExportFactory.Start(stopCh)
// 		go fromInformerFactory.Start(stopCh)
// 		go toInformerFactory.Start(stopCh)
// 		msController.Run(workers, stopCh)
// 	}

// 	run()

// }

// func checkCommandParameters() error {
// 	if controllerResyncPeriod < 0 || controllerResyncPeriod > 60 {
// 		return fmt.Errorf("controllerResyncPeriod is out of range 0~60. get: %v", controllerResyncPeriod)
// 	}

// 	if informerDefaultResync < 0 || informerDefaultResync > 60 {
// 		return fmt.Errorf("informerDefaultResync is out of range 0~60. get: %v", informerDefaultResync)
// 	}

// 	if workers < 0 || workers > 60 {
// 		return fmt.Errorf("workers is out of range 0~60. get: %v", workers)
// 	}

// 	if _, err := os.Stat(fromKubeConfig); err != nil {
// 		klog.Errorf("os.Stat fromKubeConfig failed, path: %v, err: %v ", fromKubeConfig, err)
// 		return err
// 	}

// 	if _, err := os.Stat(toKubeConfig); err != nil {
// 		klog.Errorf("os.Stat toKubeConfig failed, path: %v, err: %v ", toKubeConfig, err)
// 		return err
// 	}

// 	return nil
// }

// // newK8SClientFromKubeConfig for Config
// func newK8SClientFromKubeConfig(kubeConfigPath string) (*restclient.Config, error) {
// 	data, err := ioutil.ReadFile(kubeConfigPath)
// 	if err != nil {
// 		klog.Errorf("failed to read kubeconfig %v: %v", kubeConfigPath, err)
// 		return nil, err
// 	}

// 	config, err := clientcmd.RESTConfigFromKubeConfig(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	config.Burst = configBurst
// 	config.QPS = configQPS
// 	// 先检查集群的api连通性 3s内是否有返回
// 	config.Timeout = time.Second * time.Duration(3)

// 	return config, nil
// }
