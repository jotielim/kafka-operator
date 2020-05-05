package k8sutil

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Available types are: `ExternalDNS`, `ExternalIP`, `InternalDNS`, `InternalIP` and `Hostname`.
// By default, the addresses will be used in the following order (the first one found will be used):
//	* `ExternalDNS`
//	* `ExternalIP`
//	* `InternalDNS`
//	* `InternalIP`
//	* `Hostname`
func getSpecificNodeAddress(nodeName string, client runtimeClient.Client, addressType corev1.NodeAddressType) (string, error) {
	node := &corev1.Node{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: nodeName, Namespace: ""}, node)
	if err != nil {
		return "", err
	}

	addressMap := map[corev1.NodeAddressType]string{}
	for _, address := range node.Status.Addresses {
		addressMap[address.Type] = address.Address
	}

	nodeAddress := ""
	addressTypes := []corev1.NodeAddressType{
		addressType,
		corev1.NodeExternalDNS,
		corev1.NodeExternalIP,
		corev1.NodeInternalDNS,
		corev1.NodeInternalIP,
		corev1.NodeHostName,
	}
	for _, addressType := range addressTypes {
		if address, ok := addressMap[addressType]; ok {
			nodeAddress = address
			break
		}
	}

	if nodeAddress == "" {
		return "", errors.New("unable to find node address")
	}

	return nodeAddress, nil
}
