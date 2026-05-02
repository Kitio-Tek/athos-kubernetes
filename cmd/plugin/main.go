/*
Copyright 2026.

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

// Command kubectl-vigil is a kubectl plugin for interacting with
// Vigil-managed PostgreSQL clusters.
//
// Install:
//
//	go install github.com/Kitio-Tek/vigil-kubernetes/cmd/plugin@latest
//	mv $(go env GOPATH)/bin/plugin $(go env GOPATH)/bin/kubectl-vigil
//
// Usage:
//
//	kubectl vigil status <cluster-name> [-n namespace]
//	kubectl vigil failover <cluster-name> --to <pod-name> [-n namespace]
//	kubectl vigil backup <cluster-name> [-n namespace]
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace  string
	kubeconfig string
)

func init() {
	flag.StringVar(&namespace, "n", "default", "Kubernetes namespace")
	flag.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Path to kubeconfig file")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "status":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: kubectl vigil status <cluster-name>")
			os.Exit(1)
		}
		if err := runStatus(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		runVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func runStatus(clusterName string) error {
	client, err := newDynamicClient()
	if err != nil {
		return fmt.Errorf("building Kubernetes client: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "pg.vigil.io",
		Version:  "v1alpha1",
		Resource: "postgresclusters",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	obj, err := client.Resource(gvr).Namespace(namespace).Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting cluster %q in namespace %q: %w", clusterName, namespace, err)
	}

	spec, _, _ := nestedMap(obj.Object, "spec")
	status, _, _ := nestedMap(obj.Object, "status")

	phase, _, _ := nestedString(status, "phase")
	primary, _, _ := nestedString(status, "currentPrimary")
	ready, _, _ := nestedInt64(status, "readyInstances")
	instances, _, _ := nestedInt64(spec, "instances")
	pgVersion, _, _ := nestedInt64(spec, "postgresVersion")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\t%s\n", clusterName)
	fmt.Fprintf(w, "NAMESPACE\t%s\n", namespace)
	fmt.Fprintf(w, "PHASE\t%s\n", phase)
	fmt.Fprintf(w, "PRIMARY\t%s\n", primary)
	fmt.Fprintf(w, "READY\t%d/%d\n", ready, instances)
	fmt.Fprintf(w, "POSTGRES\t%d\n", pgVersion)
	w.Flush()

	return nil
}

func runVersion() {
	fmt.Printf("kubectl-vigil version 0.1.0\n")
	fmt.Printf("  PostgreSQL operator: Vigil Kubernetes\n")
	fmt.Printf("  API group:           pg.vigil.io/v1alpha1\n")
}

func printUsage() {
	fmt.Println("kubectl-vigil - kubectl plugin for Vigil Kubernetes operator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  kubectl vigil [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  status <cluster>   Show cluster status and instance counts")
	fmt.Println("  version            Print plugin and operator version")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -n, --namespace    Kubernetes namespace (default: default)")
	fmt.Println("      --kubeconfig   Path to kubeconfig (default: $KUBECONFIG)")
}

func newDynamicClient() (dynamic.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}
	return dynamic.NewForConfig(config)
}

func nestedMap(obj map[string]interface{}, field string) (map[string]interface{}, bool, error) {
	v, ok := obj[field]
	if !ok {
		return nil, false, nil
	}
	m, ok := v.(map[string]interface{})
	return m, ok, nil
}

func nestedString(obj map[string]interface{}, field string) (string, bool, error) {
	v, ok := obj[field]
	if !ok {
		return "", false, nil
	}
	s, ok := v.(string)
	return s, ok, nil
}

func nestedInt64(obj map[string]interface{}, field string) (int64, bool, error) {
	v, ok := obj[field]
	if !ok {
		return 0, false, nil
	}
	switch n := v.(type) {
	case int64:
		return n, true, nil
	case float64:
		return int64(n), true, nil
	case int32:
		return int64(n), true, nil
	}
	return 0, false, nil
}

