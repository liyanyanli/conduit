package k8s

import (
	"fmt"
	"net/url"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubernetesDeployments            = "deployments"
	KubernetesNamespaces             = "namespaces"
	KubernetesPods                   = "pods"
	KubernetesReplicationControllers = "replicationcontrollers"
	KubernetesServices               = "services"
)

func generateKubernetesApiBaseUrlFor(schemeHostAndPort string, namespace string, extraPathStartingWithSlash string) (*url.URL, error) {
	if string(extraPathStartingWithSlash[0]) != "/" {
		return nil, fmt.Errorf("Path must start with a [/], was [%s]", extraPathStartingWithSlash)
	}

	baseURL, err := generateBaseKubernetesApiUrl(schemeHostAndPort)
	if err != nil {
		return nil, err
	}

	urlString := fmt.Sprintf("%snamespaces/%s%s", baseURL.String(), namespace, extraPathStartingWithSlash)
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("error generating namespace URL for Kubernetes API from [%s]", urlString)
	}

	return url, nil
}

func generateBaseKubernetesApiUrl(schemeHostAndPort string) (*url.URL, error) {
	urlString := fmt.Sprintf("%s/api/v1/", schemeHostAndPort)
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("error generating base URL for Kubernetes API from [%s]", urlString)
	}
	return url, nil
}

func getConfig(fpath string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if fpath != "" {
		rules.ExplicitPath = fpath
	}
	overrides := &clientcmd.ConfigOverrides{}
	return clientcmd.
		NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).
		ClientConfig()
}

// CanonicalKubernetesNameFromFriendlyName returns a canonical name from common shorthands used in command line tools.
// This works based on https://github.com/kubernetes/kubernetes/blob/63ffb1995b292be0a1e9ebde6216b83fc79dd988/pkg/kubectl/kubectl.go#L39
func CanonicalKubernetesNameFromFriendlyName(friendlyName string) (string, error) {
	switch friendlyName {
	case "deploy", "deployment", "deployments":
		return KubernetesDeployments, nil
	case "ns", "namespace", "namespaces":
		return KubernetesNamespaces, nil
	case "po", "pod", "pods":
		return KubernetesPods, nil
	case "rc", "replicationcontroller", "replicationcontrollers":
		return KubernetesReplicationControllers, nil
	case "svc", "service", "services":
		return KubernetesServices, nil
	}

	return "", fmt.Errorf("cannot find Kubernetes canonical name from friendly name [%s]", friendlyName)
}

// GetSelectorFromObject returns a label selector based on the Kubernetes object provided.
func GetSelectorFromObject(obj runtime.Object) (labels.Selector, error) {
	switch typed := obj.(type) {
	case *apiv1.Namespace:
		return labels.Everything(), nil

	case *appsv1beta2.Deployment:
		return labels.Set(typed.Spec.Selector.MatchLabels).AsSelector(), nil

	case *apiv1.ReplicationController:
		return labels.Set(typed.Spec.Selector).AsSelector(), nil

	case *apiv1.Service:
		return labels.Set(typed.Spec.Selector).AsSelector(), nil

	default:
		return nil, fmt.Errorf("Cannot get object selector: %v", obj)
	}
}
