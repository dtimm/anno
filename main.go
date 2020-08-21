package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		panic(err)
	}

	f, err := createFetcher(config)
	if err != nil {
		panic(err)
	}

	pods, err := f()
	if err != nil {
		panic(err)
	}

	for _, p := range pods.Items {
		fmt.Printf("%s\t%s\n", p.GetName(), p.Status.PodIP)
		for k, v := range p.GetAnnotations() {
			if strings.Contains(k, "prometheus.io/path") {
				fmt.Printf("\t%s: %s\n", k, v)
			}
		}
	}
}

type fetcher func() (*v1.PodList, error)

func createFetcher(config *rest.Config) (fetcher, error) {

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
