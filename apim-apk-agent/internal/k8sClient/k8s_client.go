/*
 *  Copyright (c) 2024, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

// Package k8sclient contains the common implementation methods to invoke k8s APIs in the agent
package k8sclient

import (
	"context"
	"crypto/sha1"
	"encoding/hex"

	dpv1alpha1 "github.com/wso2/apk/common-go-libs/apis/dp/v1alpha1"
	dpv1alpha2 "github.com/wso2/apk/common-go-libs/apis/dp/v1alpha2"
	"github.com/wso2/apk/common-go-libs/utils"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/config"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/internal/loggers"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/internal/logging"
	eventhubTypes "github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/eventhub/types"
	corev1 "k8s.io/api/core/v1"
	k8error "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// DeployAPICR applies the given API struct to the Kubernetes cluster.
func DeployAPICR(api *dpv1alpha2.API, k8sClient client.Client) {
	crAPI := &dpv1alpha2.API{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: api.ObjectMeta.Namespace, Name: api.Name}, crAPI); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get API CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), api); err != nil {
			loggers.LoggerXds.Error("Unable to create API CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("API CR created: " + api.Name)
		}
	} else {
		crAPI.Spec = api.Spec
		if err := k8sClient.Update(context.Background(), crAPI); err != nil {
			loggers.LoggerXds.Error("Unable to update API CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("API CR updated: " + api.Name)
		}
	}
}

// UndeployAPICR removes the API Custom Resource from the Kubernetes cluster based on API ID label.
func UndeployAPICR(apiID string, k8sClient client.Client) {
	conf, errReadConfig := config.ReadConfigs()
	if errReadConfig != nil {
		loggers.LoggerXds.Errorf("Error reading configurations: %v", errReadConfig)
	}
	api := &dpv1alpha2.API{}
	// Retrieve all API CRs from the Kubernetes cluster
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: apiID, Namespace: conf.DataPlane.Namespace}, api); err != nil {
		loggers.LoggerXds.Errorf("Unable to list API CRs: %v", err)
	}
	if err := k8sClient.Delete(context.Background(), api); err != nil {
		loggers.LoggerXds.Errorf("Unable to delete API CR: %v", err)
	}
	loggers.LoggerXds.Infof("Deleted API CR: %s", api.Name)
}

// contains checks if a string is present in a slice of strings
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// DeployConfigMapCR applies the given ConfigMap struct to the Kubernetes cluster.
func DeployConfigMapCR(configMap *corev1.ConfigMap, k8sClient client.Client) {
	crConfigMap := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: configMap.ObjectMeta.Namespace, Name: configMap.Name}, crConfigMap); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get ConfigMap CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), configMap); err != nil {
			loggers.LoggerXds.Error("Unable to create ConfigMap CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("ConfigMap CR created: " + configMap.Name)
		}
	} else {
		crConfigMap.Data = configMap.Data
		if err := k8sClient.Update(context.Background(), crConfigMap); err != nil {
			loggers.LoggerXds.Error("Unable to update ConfigMap CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("ConfigMap CR updated: " + configMap.Name)
		}
	}
}

// DeployHTTPRouteCR applies the given HttpRoute struct to the Kubernetes cluster.
func DeployHTTPRouteCR(httpRoute *gwapiv1b1.HTTPRoute, k8sClient client.Client) {
	crHTTPRoute := &gwapiv1b1.HTTPRoute{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: httpRoute.ObjectMeta.Namespace, Name: httpRoute.Name}, crHTTPRoute); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get HTTPRoute CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), httpRoute); err != nil {
			loggers.LoggerXds.Error("Unable to create HTTPRoute CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("HTTPRoute CR created: " + httpRoute.Name)
		}
	} else {
		crHTTPRoute.Spec = httpRoute.Spec
		if err := k8sClient.Update(context.Background(), crHTTPRoute); err != nil {
			loggers.LoggerXds.Error("Unable to update HTTPRoute CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("HTTPRoute CR updated: " + httpRoute.Name)
		}
	}
}

// DeploySecretCR applies the given Secret struct to the Kubernetes cluster.
func DeploySecretCR(secret *corev1.Secret, k8sClient client.Client) {
	crSecret := &corev1.Secret{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: secret.ObjectMeta.Namespace, Name: secret.Name}, crSecret); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get Secret CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), secret); err != nil {
			loggers.LoggerXds.Error("Unable to create Secret CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Secret CR created: " + secret.Name)
		}
	} else {
		crSecret.Data = secret.Data
		if err := k8sClient.Update(context.Background(), crSecret); err != nil {
			loggers.LoggerXds.Error("Unable to update Secret CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Secret CR updated: " + secret.Name)
		}
	}
}

// DeployAuthenticationCR applies the given Authentication struct to the Kubernetes cluster.
func DeployAuthenticationCR(authPolicy *dpv1alpha2.Authentication, k8sClient client.Client) {
	crAuthPolicy := &dpv1alpha2.Authentication{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: authPolicy.ObjectMeta.Namespace, Name: authPolicy.Name}, crAuthPolicy); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get Authentication CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), authPolicy); err != nil {
			loggers.LoggerXds.Error("Unable to create Authentication CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Authentication CR created: " + authPolicy.Name)
		}
	} else {
		crAuthPolicy.Spec = authPolicy.Spec
		if err := k8sClient.Update(context.Background(), crAuthPolicy); err != nil {
			loggers.LoggerXds.Error("Unable to update Authentication CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Authentication CR updated: " + authPolicy.Name)
		}
	}
}

// DeployBackendJWTCR applies the given BackendJWT struct to the Kubernetes cluster.
func DeployBackendJWTCR(backendJWT *dpv1alpha1.BackendJWT, k8sClient client.Client) {
	crBackendJWT := &dpv1alpha1.BackendJWT{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: backendJWT.ObjectMeta.Namespace, Name: backendJWT.Name}, crBackendJWT); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get BackendJWT CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), backendJWT); err != nil {
			loggers.LoggerXds.Error("Unable to create BackendJWT CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("BackendJWT CR created: " + backendJWT.Name)
		}
	} else {
		crBackendJWT.Spec = backendJWT.Spec
		if err := k8sClient.Update(context.Background(), crBackendJWT); err != nil {
			loggers.LoggerXds.Error("Unable to update BackendJWT CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("BackendJWT CR updated: " + backendJWT.Name)
		}
	}
}

// DeployAPIPolicyCR applies the given APIPolicies struct to the Kubernetes cluster.
func DeployAPIPolicyCR(apiPolicies *dpv1alpha2.APIPolicy, k8sClient client.Client) {
	crAPIPolicies := &dpv1alpha2.APIPolicy{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: apiPolicies.ObjectMeta.Namespace, Name: apiPolicies.Name}, crAPIPolicies); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get APIPolicies CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), apiPolicies); err != nil {
			loggers.LoggerXds.Error("Unable to create APIPolicies CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("APIPolicies CR created: " + apiPolicies.Name)
		}
	} else {
		crAPIPolicies.Spec = apiPolicies.Spec
		if err := k8sClient.Update(context.Background(), crAPIPolicies); err != nil {
			loggers.LoggerXds.Error("Unable to update APIPolicies CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("APIPolicies CR updated: " + apiPolicies.Name)
		}
	}
}

// DeployInterceptorServicesCR applies the given InterceptorServices struct to the Kubernetes cluster.
func DeployInterceptorServicesCR(interceptorServices *dpv1alpha1.InterceptorService, k8sClient client.Client) {
	crInterceptorServices := &dpv1alpha1.InterceptorService{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: interceptorServices.ObjectMeta.Namespace, Name: interceptorServices.Name}, crInterceptorServices); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get InterceptorServices CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), interceptorServices); err != nil {
			loggers.LoggerXds.Error("Unable to create InterceptorServices CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("InterceptorServices CR created: " + interceptorServices.Name)
		}
	} else {
		crInterceptorServices.Spec = interceptorServices.Spec
		if err := k8sClient.Update(context.Background(), crInterceptorServices); err != nil {
			loggers.LoggerXds.Error("Unable to update InterceptorServices CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("InterceptorServices CR updated: " + interceptorServices.Name)
		}
	}
}

// DeployScopeCR applies the given Scope struct to the Kubernetes cluster.
func DeployScopeCR(scope *dpv1alpha1.Scope, k8sClient client.Client) {
	crScope := &dpv1alpha1.Scope{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: scope.ObjectMeta.Namespace, Name: scope.Name}, crScope); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get Scope CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), scope); err != nil {
			loggers.LoggerXds.Error("Unable to create Scope CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Scope CR created: " + scope.Name)
		}
	} else {
		crScope.Spec = scope.Spec
		if err := k8sClient.Update(context.Background(), crScope); err != nil {
			loggers.LoggerXds.Error("Unable to update Scope CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Scope CR updated: " + scope.Name)
		}
	}
}

// DeployRateLimitPolicyCR applies the given RateLimitPolicies struct to the Kubernetes cluster.
func DeployRateLimitPolicyCR(rateLimitPolicies *dpv1alpha1.RateLimitPolicy, k8sClient client.Client) {
	crRateLimitPolicies := &dpv1alpha1.RateLimitPolicy{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: rateLimitPolicies.ObjectMeta.Namespace, Name: rateLimitPolicies.Name}, crRateLimitPolicies); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get RateLimitPolicies CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), rateLimitPolicies); err != nil {
			loggers.LoggerXds.Error("Unable to create RateLimitPolicies CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("RateLimitPolicies CR created: " + rateLimitPolicies.Name)
		}
	} else {
		crRateLimitPolicies.Spec = rateLimitPolicies.Spec
		if err := k8sClient.Update(context.Background(), crRateLimitPolicies); err != nil {
			loggers.LoggerXds.Error("Unable to update RateLimitPolicies CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("RateLimitPolicies CR updated: " + rateLimitPolicies.Name)
		}
	}
}

// DeployBackendCR applies the given Backends struct to the Kubernetes cluster.
func DeployBackendCR(backends *dpv1alpha1.Backend, k8sClient client.Client) {
	crBackends := &dpv1alpha1.Backend{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: backends.ObjectMeta.Namespace, Name: backends.Name}, crBackends); err != nil {
		if !k8error.IsNotFound(err) {
			loggers.LoggerXds.Error("Unable to get Backends CR: " + err.Error())
		}
		if err := k8sClient.Create(context.Background(), backends); err != nil {
			loggers.LoggerXds.Error("Unable to create Backends CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Backends CR created: " + backends.Name)
		}
	} else {
		crBackends.Spec = backends.Spec
		if err := k8sClient.Update(context.Background(), crBackends); err != nil {
			loggers.LoggerXds.Error("Unable to update Backends CR: " + err.Error())
		} else {
			loggers.LoggerXds.Info("Backends CR updated: " + backends.Name)
		}
	}
}

// CreateAndUpdateTokenIssuersCR applies the given TokenIssuers struct to the Kubernetes cluster.
func CreateAndUpdateTokenIssuersCR(keyManager eventhubTypes.ResolvedKeyManager, k8sClient client.Client) error {
	conf, _ := config.ReadConfigs()
	sha1ValueofKmName := getSha1Value(keyManager.Name)
	sha1ValueOfOrganization := getSha1Value(keyManager.Organization)
	labelMap := map[string]string{"name": sha1ValueofKmName, "organization": sha1ValueOfOrganization}

	tokenIssuer := dpv1alpha2.TokenIssuer{
		ObjectMeta: metav1.ObjectMeta{Name: keyManager.UUID,
			Namespace: conf.DataPlane.Namespace,
			Labels:    labelMap,
		},
		Spec: dpv1alpha2.TokenIssuerSpec{
			Name:                keyManager.Name,
			Organization:        keyManager.Organization,
			Issuer:              keyManager.KeyManagerConfig.Issuer,
			ClaimMappings:       marshalClaimMappings(keyManager.KeyManagerConfig.ClaimMappings),
			SignatureValidation: marshalSignatureValidation(keyManager.KeyManagerConfig),
			TargetRef:           &v1alpha2.PolicyTargetReference{Group: "gateway.networking.k8s.io", Kind: "Gateway", Name: "default"},
		},
	}
	tokenIssuer.Spec.ConsumerKeyClaim = "azp"
	if keyManager.KeyManagerConfig.ConsumerKeyClaim != "" {
		tokenIssuer.Spec.ConsumerKeyClaim = keyManager.KeyManagerConfig.ConsumerKeyClaim
	}
	keyManager.KeyManagerConfig.ScopesClaim = "scope"
	if keyManager.KeyManagerConfig.ScopesClaim != "" {
		tokenIssuer.Spec.ScopesClaim = keyManager.KeyManagerConfig.ScopesClaim
	}
	err := k8sClient.Create(context.Background(), &tokenIssuer)
	if err != nil {
		loggers.LoggerXds.Error("Unable to create TokenIssuer CR: " + err.Error())
		return err
	}
	loggers.LoggerXds.Debug("TokenIssuer CR created: " + tokenIssuer.Name)
	return nil
}

// DeleteTokenIssuersCR deletes the TokenIssuers struct from the Kubernetes cluster.
func DeleteTokenIssuersCR(k8sClient client.Client, keymanagerName string, tenantDomain string) error {
	conf, _ := config.ReadConfigs()
	sha1ValueofKmName := getSha1Value(keymanagerName)
	sha1ValueOfOrganization := getSha1Value(tenantDomain)
	labelMap := map[string]string{"name": sha1ValueofKmName, "organization": sha1ValueOfOrganization}
	// Create a list option with the label selector
	listOption := &client.ListOptions{
		Namespace:     conf.DataPlane.Namespace,
		LabelSelector: labels.SelectorFromSet(labelMap),
	}

	tokenIssuerList := &dpv1alpha2.TokenIssuerList{}
	err := k8sClient.List(context.Background(), tokenIssuerList, listOption)
	if err != nil {
		loggers.LoggerAPI.Error("Unable to list TokenIssuer CR: " + err.Error())
	}
	if len(tokenIssuerList.Items) == 0 {
		loggers.LoggerXds.Debug("No TokenIssuer CR found for deletion")
	}
	for _, tokenIssuer := range tokenIssuerList.Items {
		err := k8sClient.Delete(context.Background(), &tokenIssuer, &client.DeleteOptions{})
		if err != nil {
			loggers.LoggerAPI.Error("Unable to delete TokenIssuer CR: " + err.Error())
			return err
		}
		loggers.LoggerXds.Debug("TokenIssuer CR deleted: " + tokenIssuer.Name)
	}
	return nil
}

// UpdateTokenIssuersCR applies the given TokenIssuers struct to the Kubernetes cluster.
func UpdateTokenIssuersCR(keyManager eventhubTypes.ResolvedKeyManager, k8sClient client.Client) error {
	conf, _ := config.ReadConfigs()
	sha1ValueofKmName := getSha1Value(keyManager.Name)
	sha1ValueOfOrganization := getSha1Value(keyManager.Organization)
	labelMap := map[string]string{"name": sha1ValueofKmName, "organization": sha1ValueOfOrganization}
	tokenIssuer := &dpv1alpha2.TokenIssuer{}
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: keyManager.UUID, Namespace: conf.DataPlane.Namespace}, tokenIssuer)
	if err != nil {
		loggers.LoggerAPI.Error("Unable to get TokenIssuer CR: " + err.Error())
		return err
	}
	tokenIssuer.ObjectMeta.Labels = labelMap
	tokenIssuer.Spec.Name = keyManager.Name
	tokenIssuer.Spec.Organization = keyManager.Organization
	tokenIssuer.Spec.Issuer = keyManager.KeyManagerConfig.Issuer
	tokenIssuer.Spec.ClaimMappings = marshalClaimMappings(keyManager.KeyManagerConfig.ClaimMappings)
	tokenIssuer.Spec.SignatureValidation = marshalSignatureValidation(keyManager.KeyManagerConfig)
	tokenIssuer.Spec.TargetRef = &v1alpha2.PolicyTargetReference{Group: "gateway.networking.k8s.io", Kind: "Gateway", Name: "default"}
	if keyManager.KeyManagerConfig.ConsumerKeyClaim != "" {
		tokenIssuer.Spec.ConsumerKeyClaim = keyManager.KeyManagerConfig.ConsumerKeyClaim
	}
	if keyManager.KeyManagerConfig.ScopesClaim != "" {
		tokenIssuer.Spec.ScopesClaim = keyManager.KeyManagerConfig.ScopesClaim
	}
	err = k8sClient.Update(context.Background(), tokenIssuer)
	if err != nil {
		loggers.LoggerAPI.Error("Unable to update TokenIssuer CR: " + err.Error())
		return err
	}
	loggers.LoggerXds.Debug("TokenIssuer CR updated: " + tokenIssuer.Name)
	return nil
}
func marshalSignatureValidation(keyManagerConfig eventhubTypes.KeyManagerConfig) *dpv1alpha2.SignatureValidation {
	if keyManagerConfig.CertificateType != "" && keyManagerConfig.CertificateValue != "" {
		if keyManagerConfig.CertificateType == "JWKS" {
			return &dpv1alpha2.SignatureValidation{JWKS: &dpv1alpha2.JWKS{URL: keyManagerConfig.CertificateValue}}
		}
		return &dpv1alpha2.SignatureValidation{Certificate: &dpv1alpha2.CERTConfig{CertificateInline: &keyManagerConfig.CertificateValue}}
	}
	return nil
}

func marshalClaimMappings(claimMappings []eventhubTypes.Claim) *[]dpv1alpha2.ClaimMapping {
	resolvedClaimMappings := make([]dpv1alpha2.ClaimMapping, 0)
	for _, claim := range claimMappings {
		resolvedClaimMappings = append(resolvedClaimMappings, dpv1alpha2.ClaimMapping{RemoteClaim: claim.RemoteClaim, LocalClaim: claim.LocalClaim})
	}
	return &resolvedClaimMappings
}
func getSha1Value(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// RetrieveAllAPISFromK8s retrieves all the API CRs from the Kubernetes cluster
func RetrieveAllAPISFromK8s(k8sClient client.Client, nextToken string) ([]dpv1alpha2.API, string, error) {
	apiList := dpv1alpha2.APIList{}
	resolvedAPIList := make([]dpv1alpha2.API, 0)
	var err error
	if nextToken == "" {
		err = k8sClient.List(context.Background(), &apiList, &client.ListOptions{Namespace: utils.GetOperatorPodNamespace()})
	} else {
		err = k8sClient.List(context.Background(), &apiList, &client.ListOptions{Namespace: utils.GetOperatorPodNamespace(), Continue: nextToken})
	}
	if err != nil {
		loggers.LoggerSync.ErrorC(logging.PrintError(logging.Error1102, logging.CRITICAL, "Failed to get application from k8s %v", err.Error()))
		return nil, "", err
	}
	resolvedAPIList = append(resolvedAPIList, apiList.Items...)
	if apiList.Continue != "" {
		tempAPIList, _, err := RetrieveAllAPISFromK8s(k8sClient, apiList.Continue)
		if err != nil {
			return nil, "", err
		}
		resolvedAPIList = append(resolvedAPIList, tempAPIList...)
	}
	return resolvedAPIList, apiList.Continue, nil
}
