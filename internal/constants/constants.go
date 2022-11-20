/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package constants defines some global constants
package constants

const (
	// TODO: get paths by container ENV or actual value instead of fixed constant
	DataPath = "/root"

	// StoreContainerName is the name of store container
	StoreContainerName = "store"

	// TriggerContainerName is the name of tigger container
	TriggerContainerName = "tigger"

	// TimerContainerName is the name of timer container
	TimerContainerName = "timer"

	// GatewayContainerName is the name of gateway container
	GatewayContainerName = "gateway"

	// BasicCommand is basic command of exec function
	BasicCommand = "sh"

	// AdminToolDir is the RocketMQ Admin directory in operator image
	AdminToolDir = "/home/rocketmq/operator/bin/mqadmin"

	// StoreConfigDir is the directory of config file
	StoreConfigDir = DataPath + "/store/config"

	// TopicJsonDir is the directory of topics.json
	TopicJsonDir = StoreConfigDir + "/topics.json"

	// SubscriptionGroupJsonDir is the directory of subscriptionGroup.json
	SubscriptionGroupJsonDir = StoreConfigDir + "/subscriptionGroup.json"

	// BrokerConfigDir is the directory of the mounted config file
	BrokerConfigPath = DataPath + "/rocketmq/broker/conf"

	// BrokerConfigName is the name of mounted configuration file
	BrokerConfigName = "broker-common.conf"

	// UpdateBrokerConfig is update broker config command
	UpdateBrokerConfig = "updateBrokerConfig"

	// ParamNameServiceAddress is the name of name server list parameter
	ParamNameServiceAddress = "namesrvAddr"

	// EnvNameServiceAddress is the container environment variable name of name server list
	EnvNameServiceAddress = "NAMESRV_ADDR"

	// EnvBrokerId is the container environment variable name of broker id
	EnvBrokerId = "BROKER_ID"

	// EnvBrokerClusterName is the container environment variable name of broker cluster name
	EnvBrokerClusterName = "BROKER_CLUSTER_NAME"

	// EnvBrokerName is the container environment variable name of broker name
	EnvBrokerName = "BROKER_NAME"

	// ConfigMountPath is the directory of Vanus configd files
	ConfigMountPath = "/vanus/config"

	// StoreMountPath is the directory of Store data files
	StoreMountPath = "/data"

	// NameServiceMainContainerPort is the main port number of name server container
	NameServiceMainContainerPort = 9876

	// NameServiceMainContainerPortName is the main port name of name server container
	NameServiceMainContainerPortName = "main"

	// BrokerVipContainerPort is the VIP port number of broker container
	BrokerVipContainerPort = 10909

	// BrokerVipContainerPortName is the VIP port name of broker container
	BrokerVipContainerPortName = "vip"

	// BrokerMainContainerPort is the main port number of broker container
	BrokerMainContainerPort = 10911

	// BrokerMainContainerPortName is the main port name of broker container
	BrokerMainContainerPortName = "main"

	// BrokerHighAvailabilityContainerPort is the high availability port number of broker container
	BrokerHighAvailabilityContainerPort = 10912

	// BrokerHighAvailabilityContainerPortName is the high availability port name of broker container
	BrokerHighAvailabilityContainerPortName = "ha"

	// ConsoleContainerPort is the port number of RocketMQ Console container
	ConsoleContainerPort = 8080

	// ConsoleContainerPortName is the port name of RocketMQ Console container
	ConsoleContainerPortName = "console"

	// StorageModeStorageClass is the name of StorageClass storage mode
	StorageModeStorageClass = "StorageClass"

	// StorageModeEmptyDir is the name of EmptyDir storage mode
	StorageModeEmptyDir = "EmptyDir"

	// StorageModeHostPath is the name pf HostPath storage mode
	StorageModeHostPath = "HostPath"

	// RestartBrokerPodIntervalInSecond is restart broker pod interval in second
	RestartBrokerPodIntervalInSecond = 30

	// WaitForNameServerReadyInSecond is the time broker sleep for waiting nameserver ready in second
	WaitForNameServerReadyInSecond = 1

	// MinMetadataJsonFileSize is the threshold value if file length is lower than this will be considered as invalid
	MinMetadataJsonFileSize = 5

	// MinIpListLength is the threshold value if the name server list parameter length is shorter than this will be considered as invalid
	MinIpListLength = 8

	// CheckConsumeFinishIntervalInSecond is the interval of checking whether the consumption process is finished in second
	CheckConsumeFinishIntervalInSecond = 5

	// RequeueIntervalInSecond is an universal interval of the reconcile function
	RequeueIntervalInSecond = 6

	// Topic is the topic field index of the output when using command check consume progress
	Topic = 0

	// BrokerName is the broker name field index of the output when using command check consume progress
	BrokerName = 1

	// Diff is the diff field index of the output when using command check consume progress
	Diff = 6

	// TopicListTopic is the topic field index of the output when using command check topic list
	TopicListTopic = 1

	// TopicListConsumerGroup is the consumer group field index of the output when using command check topic list
	TopicListConsumerGroup = 2

	// ControllerContainerName is the name of Controller container
	ControllerContainerName = "controller"

	// the container environment variable name of controller pod ip
	EnvControllerPodIP = "POD_IP"

	// the container environment variable name of controller pod name
	EnvControllerPodName = "POD_NAME"

	// the container environment variable name of controller log level
	EnvControllerLogLevel = "VANUS_LOG_LEVEL"
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
