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

package nodePort

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
	componentName = "nodePort"
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
		for _, externalListener := range r.KafkaCluster.Spec.ListenersConfig.ExternalListeners {
			for _, broker := range r.KafkaCluster.Spec.Brokers {
				o := r.nodePort(log, externalListener, broker, r.KafkaCluster.Name)
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

	log.V(1).Info("Reconciled")

	return nil
}

func (r *Reconciler) nodePort(log logr.Logger, externalListener v1beta1.ExternalListenerConfig,
	broker v1beta1.Broker, clusterName string) runtime.Object {
	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(fmt.Sprintf("%s-%d-svc", clusterName, broker.Id), map[string]string{}, r.KafkaCluster),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":      "kafka",
				"brokerId": strconv.Itoa(int(broker.Id)),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       fmt.Sprintf("broker-%d", broker.Id),
					Port:       externalListener.ContainerPort,
					TargetPort: intstr.FromInt(int(externalListener.ContainerPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			//Type: corev1.ServiceTypeNodePort,
		},
	}
}
