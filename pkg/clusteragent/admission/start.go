// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build kubeapiserver

package admission

import (
	"time"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/controllers/secret"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/controllers/webhook"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver/common"
	"github.com/DataDog/datadog-agent/pkg/util/log"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ControllerContext holds necessary context for the admission controllers
type ControllerContext struct {
	IsLeaderFunc func() bool
	Informers    informers.SharedInformerFactory
	Client       kubernetes.Interface
	StopCh       chan struct{}
}

// StartControllers starts the secret and webhook controllers
func StartControllers(ctx ControllerContext) error {
	if !config.Datadog.GetBool("admission_controller.enabled") {
		log.Info("Admission controller is disabled")
		return nil
	}

	secretConfig := secret.Config{
		Name: config.Datadog.GetString("admission_controller.certificate.secret_name"),
		Ns:   common.GetResourcesNamespace(),
		Svc:  config.Datadog.GetString("admission_controller.service_name"),
		Cert: secret.CertConfig{
			ExpirationThreshold: config.Datadog.GetDuration("admission_controller.certificate.expiration_threshold") * time.Hour,
			ValidityBound:       config.Datadog.GetDuration("admission_controller.certificate.validity_bound") * time.Hour,
		},
	}
	secretController := secret.NewController(
		ctx.Client,
		ctx.Informers.Core().V1().Secrets(),
		ctx.IsLeaderFunc,
		secretConfig,
	)
	go secretController.Run(ctx.StopCh)

	webhookConfig := webhook.Config{
		Name:       config.Datadog.GetString("admission_controller.webhook_name"),
		Templates:  generateWebhooks(),
		SecretName: config.Datadog.GetString("admission_controller.certificate.secret_name"),
		SecretNs:   common.GetResourcesNamespace(),
	}
	webhookController := webhook.NewController(
		ctx.Client,
		ctx.Informers.Core().V1().Secrets(),
		ctx.Informers.Admissionregistration().V1beta1().MutatingWebhookConfigurations(),
		ctx.IsLeaderFunc,
		webhookConfig,
	)
	go webhookController.Run(ctx.StopCh)

	ctx.Informers.Start(ctx.StopCh)

	return apiserver.SyncInformers(map[apiserver.InformerName]cache.SharedInformer{
		apiserver.SecretsInformer:  ctx.Informers.Core().V1().Secrets().Informer(),
		apiserver.WebhooksInformer: ctx.Informers.Admissionregistration().V1beta1().MutatingWebhookConfigurations().Informer(),
	})
}
