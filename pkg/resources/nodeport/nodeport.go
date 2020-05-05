// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodeport

import (
	"fmt"
	"strconv"

	"github.com/banzaicloud/kafka-operator/api/v1beta1"
	"github.com/banzaicloud/kafka-operator/pkg/k8sutil"
	"github.com/banzaicloud/kafka-operator/pkg/resources"
	"github.com/banzaicloud/kafka-operator/pkg/resources/templates"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName = "nodeport"
)

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

// New creates a new reconciler for Envoy
func New(client client.Client, cluster *v1beta1.KafkaCluster) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:       client,
			KafkaCluster: cluster,
		},
	}
}

// Reconcile implements the reconcile logic for Envoy
func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.V(1).Info("Reconciling")
	if r.KafkaCluster.Spec.ListenersConfig.ExternalListeners != nil {

		for _, externalListenerConfig := range r.KafkaCluster.Spec.ListenersConfig.ExternalListeners {
			if externalListenerConfig.ServiceType == string(corev1.ServiceTypeNodePort) {
				// Create kafka-external-bootstrap svc
				serviceName := fmt.Sprintf("%s-external-bootstrap", r.KafkaCluster.Name)
				selector := map[string]string{
					"app":      "kafka",
					"kafka_cr": r.KafkaCluster.Name,
				}
				svc := r.nodePort(log, externalListenerConfig, serviceName, selector, externalListenerConfig.ExternalStartingPort)
				if err := k8sutil.Reconcile(log, r.Client, svc, r.KafkaCluster); err != nil {
					return emperror.WrapWith(
						err,
						"failed to reconcile resource",
						"resource",
						svc.GetObjectKind().GroupVersionKind(),
					)
				}

				// Create svc for each broker
				for _, broker := range r.KafkaCluster.Spec.Brokers {
					serviceName := fmt.Sprintf("%s-%d-svc", r.KafkaCluster.Name, broker.Id)
					selector := map[string]string{
						"app":      "kafka",
						"kafka_cr": r.KafkaCluster.Name,
						"brokerId": strconv.Itoa(int(broker.Id)),
					}
					nodePort := getBrokerNodePort(externalListenerConfig, broker.Id)
					o := r.nodePort(log, externalListenerConfig, serviceName, selector, nodePort)
					if err := k8sutil.Reconcile(log, r.Client, o, r.KafkaCluster); err != nil {
						return emperror.WrapWith(
							err,
							"failed to reconcile resource",
							"resource",
							o.GetObjectKind().GroupVersionKind(),
						)
					}
				}
			}
		}
	}

	log.V(1).Info("Reconciled")

	return nil
}

func (r *Reconciler) nodePort(
	log logr.Logger,
	externalListenerConfig v1beta1.ExternalListenerConfig,
	serviceName string,
	selector map[string]string,
	nodePort int32,
) runtime.Object {
	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, map[string]string{}, r.KafkaCluster),
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "external",
					Port:       externalListenerConfig.ContainerPort,
					TargetPort: intstr.FromInt(int(externalListenerConfig.ContainerPort)),
					Protocol:   corev1.ProtocolTCP,
					NodePort:   nodePort,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
}

func getBrokerNodePort(externalListenerConfig v1beta1.ExternalListenerConfig, brokerId int32) (nodePort int32) {
	if externalListenerConfig.Overrides.Brokers == nil {
		return 0
	}
	for _, broker := range externalListenerConfig.Overrides.Brokers {
		if broker.Id == brokerId {
			return broker.NodePort
		}
	}
	return 0
}
