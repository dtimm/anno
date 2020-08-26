package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dtimm/anno/proxy"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	config, err := getConfig()
	if err != nil {
		panic(fmt.Errorf("no kubeconfig available"))
	}

	f, err := createFetcher(config)
	if err != nil {
		panic(err)
	}

	p := proxy.NewProxy(proxy.Config{
		Fetcher: f,
		Port:    8080,
	})
	p.Start()
	defer p.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func createFetcher(config *rest.Config) (proxy.Fetcher, error) {
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return func() (*v1.PodList, error) {
		o := metav1.ListOptions{}
		return c.CoreV1().Pods("cf-workloads").List(context.TODO(), o)
	}, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getConfig() (*rest.Config, error) {
	k, err := rest.InClusterConfig()
	if err == nil {
		return k, err
	}

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}
