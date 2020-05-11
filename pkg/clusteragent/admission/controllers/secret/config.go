// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build kubeapiserver

package secret

import "time"

type CertConfig struct {
	ExpirationThreshold time.Duration
	ValidityBound       time.Duration
}

type Config struct {
	Name string
	Ns   string
	Svc  string
	Cert CertConfig
}

func (s *Config) GetName() string                     { return s.Name }
func (s *Config) GetNs() string                       { return s.Ns }
func (s *Config) GetSvc() string                      { return s.Svc }
func (s *Config) GetCertExpiration() time.Duration    { return s.Cert.ExpirationThreshold }
func (s *Config) GetCertValidityBound() time.Duration { return s.Cert.ValidityBound }
