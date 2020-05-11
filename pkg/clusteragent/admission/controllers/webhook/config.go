// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build kubeapiserver

package webhook

import admiv1beta1 "k8s.io/api/admissionregistration/v1beta1"

type Config struct {
	Name       string
	SecretName string
	SecretNs   string
	Templates  []admiv1beta1.MutatingWebhook
}

func (w *Config) GetName() string                             { return w.Name }
func (w *Config) GetSecretName() string                       { return w.SecretName }
func (w *Config) GetSecretNs() string                         { return w.SecretNs }
func (w *Config) GetTemplates() []admiv1beta1.MutatingWebhook { return w.Templates }
