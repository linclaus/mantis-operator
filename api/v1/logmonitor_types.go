/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LogMonitorSpec defines the desired state of LogMonitor
type LogMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of LogMonitor. Edit LogMonitor_types.go to remove/update
	Labels   Label  `json:"labels,omitempty"`
	Dsl      string `json:"dsl,omitempty"`
	Promql   string `json:"promql,omitempty"`
	Duration string `json:"duration,omitempty"`
}

type Label struct {
	Application      string `json:"application,omitempty"`
	AlarmSource      string `json:"alarmSource,omitempty"`
	AlarmContent     string `json:"alarmContent,omitempty"`
	MetricName       string `json:"metricName,omitempty"`
	MetricInstanceId string `json:"metricInstanceId,omitempty"`
	StrategyName     string `json:"strategyName,omitempty"`
	StrategyId       string `json:"strategyId,omitempty"`
	Contact          string `json:"contact,omitempty"`
	ContainerName    string `json:"containerName,omitempty"`
}

// LogMonitorStatus defines the observed state of LogMonitor
type LogMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status     string `json:"created,omitempty"`
	RetryTimes int    `json:"retryTimes,omitempt"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true

// LogMonitor is the Schema for the logmonitors API
type LogMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogMonitorSpec   `json:"spec,omitempty"`
	Status LogMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LogMonitorList contains a list of LogMonitor
type LogMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogMonitor{}, &LogMonitorList{})
}
