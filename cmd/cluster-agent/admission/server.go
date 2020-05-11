// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017-2020 Datadog, Inc.

// +build kubeapiserver

package admission

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/mutate"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver/common"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/certificate"
	"github.com/DataDog/datadog-agent/pkg/util/log"

	admiv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
)

const jsonContentType = "application/json"

// RunServer creates and start a k8s admission webhook server
func RunServer(mainCtx context.Context, client kubernetes.Interface) error {
	mux := http.NewServeMux()
	mux.HandleFunc(config.Datadog.GetString("admission_controller.inject_config.endpoint"), func(w http.ResponseWriter, r *http.Request) {
		mutateHandler(w, r, mutate.InjectConfig)
	})
	mux.HandleFunc(config.Datadog.GetString("admission_controller.inject_tags.endpoint"), func(w http.ResponseWriter, r *http.Request) {
		mutateHandler(w, r, mutate.InjectTags)
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Datadog.GetInt("admission_controller.port")),
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				secretNs := common.GetResourcesNamespace()
				secretName := config.Datadog.GetString("admission_controller.certificate.secret_name")
				cert, err := certificate.GetCertificateFromSecret(secretNs, secretName, client)
				if err != nil {
					log.Errorf("Couldn't fetch certificate: %v", err)
				}
				return cert, nil
			},
		},
	}
	go func() error {
		return log.Error(server.ListenAndServeTLS("", ""))
	}()

	<-mainCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func mutateHandler(w http.ResponseWriter, r *http.Request, mutateFunc func(*admiv1beta1.AdmissionRequest) (*admiv1beta1.AdmissionResponse, error)) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Warnf("Invalid method %s, only POST requests are allowed", r.Method)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Warnf("Could not read request body: %v", err)
		return
	}
	defer r.Body.Close()

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		log.Warnf("Unsupported content type %s, only %s is supported", contentType, jsonContentType)
		return
	}

	var admissionReviewReq admiv1beta1.AdmissionReview
	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Warnf("Could not deserialize request: %v", err)
		return
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Warn("Malformed admission review: request is nil")
		return
	}

	var admissionReviewResp admiv1beta1.AdmissionReview
	resp, err := mutateFunc(admissionReviewReq.Request)
	if err != nil {
		log.Warnf("Failed to mutate: %v", err)
		admissionReviewResp = admiv1beta1.AdmissionReview{
			Response: &admiv1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
				Allowed: false,
			},
		}
	} else {
		admissionReviewResp = admiv1beta1.AdmissionReview{
			Response: resp,
		}
	}
	admissionReviewResp.Response.UID = admissionReviewReq.Request.UID

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&admissionReviewResp)
	if err != nil {
		log.Warnf("Failed to encode the response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
