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

//import (
//	"github.com/banzaicloud/kafka-operator/api/v1beta1"
//	"github.com/banzaicloud/kafka-operator/pkg/resources/templates"
//	"github.com/go-logr/logr"
//	corev1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/util/intstr"
//)
//
//func (r *Reconciler) service(log logr.Logger, externalListenerConfig v1beta1.ExternalListenerConfig) runtime.Object {
//	return &corev1.Service{
//		ObjectMeta: templates.ObjectMeta(serviceName, map[string]string{}, r.KafkaCluster),
//		Spec: corev1.ServiceSpec{
//			Selector: selector,
//			Ports: []corev1.ServicePort{
//				{
//					Name:       "external",
//					Port:       externalListener.ContainerPort,
//					TargetPort: intstr.FromInt(int(externalListener.ContainerPort)),
//					Protocol:   corev1.ProtocolTCP,
//					NodePort:   nodePort,
//				},
//			},
//			Type: corev1.ServiceTypeNodePort,
//		},
//	}
//}
