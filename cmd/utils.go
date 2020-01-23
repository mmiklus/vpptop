/*
 * Copyright (c) 2019 PANTHEON.tech.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/PantheonTechnologies/vpptop/client"
	"github.com/PantheonTechnologies/vpptop/gui"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// startClient is a blocking call that starts
// the terminal frontend for displaying VPP metrics.
func startClient(socket, raddr string, file io.Writer) error {
	var lightTheme bool
	if _, lightTheme = os.LookupEnv("VPPTOP_THEME_LIGHT"); lightTheme {
		gui.SetLightTheme()
	}

	app := client.NewApp(lightTheme)
	if err := app.Init(socket, raddr); err != nil {
		return fmt.Errorf("error occurred during client init: %v", err)
	}

	log.SetOutput(file)
	app.Run()
	return nil
}

// resolveNode resolves an ip address from a given nodeName/ip-addr.
func resolveNode(kubeconfig string, name string) (string, bool) {
	if ip := net.ParseIP(name); ip != nil {
		return name, true
	}

	node, found := findNode(getNodes(kubeconfig), name)
	if !found {
		return "", false
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeExternalIP || addr.Type == v1.NodeInternalIP {
			return addr.Address, true
		}
	}

	return "", false
}

// findNode finds the specified node in the node list.
func findNode(nodes []v1.Node, name string) (v1.Node, bool) {
	for _, node := range nodes {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeHostName && addr.Address == name {
				return node, true
			}
		}
	}

	return v1.Node{}, false
}

// getNodes returns all k8s nodes in the cluster.
func getNodes(kubeconfig string) []v1.Node {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil
	}
	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return nodeList.Items
}

func homeDir() string {
	return os.Getenv("HOME")
}