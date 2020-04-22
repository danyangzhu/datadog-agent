// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build kubeapiserver

package apiserver

const (
	autoscalerNowHandleMsgEvent = "Autoscaler is now handled by the Cluster-Agent"
)

// controllerName represents the cluster agent controller names
type controllerName string

const (
	metadataController    controllerName = "metadata"
	autoscalersController controllerName = "autoscalers"
	servicesController    controllerName = "services"
	endpointsController   controllerName = "endpoints"
	secretsController     controllerName = "secrets"
)

// InformerName represents the kubernetes informer names
// TODO: make it private
type InformerName string

const (
	nodesInformer     InformerName = "nodes"
	endpointsInformer InformerName = "endpoints"
	wpaInformer       InformerName = "wpa"
	hpaInformer       InformerName = "hpa"
	servicesInformer  InformerName = "services"
	secretsInformer   InformerName = "secrets"
	// PodsInformer name, TODO: make private
	PodsInformer InformerName = "pods"
	// DeploysInformer name, TODO: make private
	DeploysInformer InformerName = "deploys"
	// ReplicasetsInformer name, TODO: make it private
	ReplicasetsInformer InformerName = "replicasets"
)
