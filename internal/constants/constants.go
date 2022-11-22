// Copyright 2022 Linkall Inc.
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

// Package constants defines some global constants
package constants

const (
	// StoreContainerName is the name of store container
	StoreContainerName = "store"

	// TriggerContainerName is the name of tigger container
	TriggerContainerName = "tigger"

	// TimerContainerName is the name of timer container
	TimerContainerName = "timer"

	// GatewayContainerName is the name of gateway container
	GatewayContainerName = "gateway"

	// BrokerConfigName is the name of mounted configuration file
	ControllerConfigName = "vanus-controller.yaml"

	// ConfigMountPath is the directory of Vanus configd files
	ConfigMountPath = "/vanus/config"

	// StoreMountPath is the directory of Store data files
	StoreMountPath = "/data"

	// StorageModeStorageClass is the name of StorageClass storage mode
	StorageModeStorageClass = "StorageClass"

	// StorageModeEmptyDir is the name of EmptyDir storage mode
	StorageModeEmptyDir = "EmptyDir"

	// StorageModeHostPath is the name pf HostPath storage mode
	StorageModeHostPath = "HostPath"

	// ControllerContainerName is the name of Controller container
	ControllerContainerName = "controller"

	// the container environment variable name of controller pod ip
	EnvControllerPodIP = "POD_IP"

	// the container environment variable name of controller pod name
	EnvControllerPodName = "POD_NAME"

	// the container environment variable name of controller log level
	EnvControllerLogLevel = "VANUS_LOG_LEVEL"

	// RequeueIntervalInSecond is an universal interval of the reconcile function
	RequeueIntervalInSecond = 6
)

var (
	// GroupNum is the number of broker group
	GroupNum = 0

	// NameServersStr is the name server list
	NameServersStr = ""

	// IsNameServersStrUpdated is whether the name server list is updated
	IsNameServersStrUpdated = false

	// IsNameServersStrInitialized is whether the name server list is initialized
	IsNameServersStrInitialized = false

	// BrokerClusterName is the broker cluster name
	BrokerClusterName = ""

	// svc of controller for brokers
	ControllerAccessPoint = ""
)
