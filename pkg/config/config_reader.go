package config

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/konflux-ci/konflux-build-cli/pkg/clients"
	"github.com/konflux-ci/konflux-build-cli/pkg/logger"
	"gopkg.in/ini.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var rawConfigLogger = logger.Logger.WithField("logger", "KonfluxConfig")

var NewConfigReader func() (ConfigReader, error) = ConfigReaderFactory

// KonfluxRawConfig holds unmodified data that was read from the config source.
type KonfluxRawConfig struct {
	AllowCacheProxy                    string
	HttpProxy                          string
	NoProxy                            string
	HermetoBundlerProxy                string
	HermetoCargoProxy                  string
	HermetoGomodProxy                  string
	HermetoNpmProxy                    string
	HermetoPipProxy                    string
	HermetoPnpmProxy                   string
	HermetoYarnProxy                   string
	HermetoPackageRegistryProxyAllowed string
	FulcioInternalURL                  string
	FulcioExternalURL                  string
	RekorInternalURL                   string
	RekorExternalURL                   string
	TufInternalURL                     string
	TufExternalURL                     string
	DefaultOIDCIssuer                  string
	EnableKeylessSigning               string
}

func (k *KonfluxRawConfig) LogData() {
	v := reflect.ValueOf(k).Elem()
	for _, field := range reflect.VisibleFields(v.Type()) {
		rawConfigLogger.Infof("%s: %s", field.Name, v.FieldByName(field.Name).String())
	}
}

// ConfigReader defines the interface for reading config data.
type ConfigReader interface {
	ReadConfigData() (*KonfluxRawConfig, error)
}

// ConfigReaderFactory returns config reader according to the configured config source.
func ConfigReaderFactory() (ConfigReader, error) {
	platformConfigFile := os.Getenv("PLATFORM_CONFIG_FILE")
	if platformConfigFile != "" {
		return &IniFileReader{FilePath: platformConfigFile}, nil
	} else {
		clientset, err := clients.NewKubeClientSet()
		if err != nil {
			return nil, err
		}
		return &K8sConfigMapReader{Name: "cluster-config", Namespace: "konflux-info", Clientset: clientset}, nil
	}
}

// K8sConfigMapReader reads configuration from a Kubernetes cluster.
type K8sConfigMapReader struct {
	Name      string
	Namespace string
	Clientset kubernetes.Interface
}

// ReadConfigData reads the config from the ConfigMap data of a Kubernetes cluster.
func (k *K8sConfigMapReader) ReadConfigData() (*KonfluxRawConfig, error) {
	ctx := context.Background()
	configMap, err := k.Clientset.CoreV1().ConfigMaps(k.Namespace).Get(ctx, k.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", k.Namespace, k.Name, err)
	}

	rawConfig := &KonfluxRawConfig{
		AllowCacheProxy:                    configMap.Data["allow-cache-proxy"],
		HttpProxy:                          configMap.Data["http-proxy"],
		NoProxy:                            configMap.Data["no-proxy"],
		HermetoBundlerProxy:                configMap.Data["package-registry-proxy-bundler-url"],
		HermetoCargoProxy:                  configMap.Data["package-registry-proxy-cargo-url"],
		HermetoGomodProxy:                  configMap.Data["package-registry-proxy-gomod-url"],
		HermetoNpmProxy:                    configMap.Data["package-registry-proxy-npm-url"],
		HermetoPipProxy:                    configMap.Data["package-registry-proxy-pip-url"],
		HermetoPnpmProxy:                   configMap.Data["package-registry-proxy-pnpm-url"],
		HermetoYarnProxy:                   configMap.Data["package-registry-proxy-yarn-url"],
		HermetoPackageRegistryProxyAllowed: configMap.Data["allow-package-registry-proxy"],
		FulcioInternalURL:                  configMap.Data["fulcioInternalUrl"],
		FulcioExternalURL:                  configMap.Data["fulcioExternalUrl"],
		RekorInternalURL:                   configMap.Data["rekorInternalUrl"],
		RekorExternalURL:                   configMap.Data["rekorExternalUrl"],
		TufInternalURL:                     configMap.Data["tufInternalUrl"],
		TufExternalURL:                     configMap.Data["tufExternalUrl"],
		DefaultOIDCIssuer:                  configMap.Data["defaultOIDCIssuer"],
		EnableKeylessSigning:               configMap.Data["enableKeylessSigning"],
	}
	rawConfig.LogData()
	return rawConfig, nil
}

// INIFileReader reads configuration from a local INI file.
type IniFileReader struct {
	FilePath string
}

// ReadConfigData reads platform config from the INI file
func (y *IniFileReader) ReadConfigData() (*KonfluxRawConfig, error) {
	cfg, err := ini.Load(y.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ini file: %v", err)
	}

	cacheProxy := cfg.Section("cache-proxy")
	artifactRegistry := cfg.Section("artifact-registry")
	keylessSigning := cfg.Section("keyless-signing")

	rawConfig := &KonfluxRawConfig{
		AllowCacheProxy:                    cacheProxy.Key("allow-cache-proxy").String(),
		HttpProxy:                          cacheProxy.Key("http-proxy").String(),
		NoProxy:                            cacheProxy.Key("no-proxy").String(),
		HermetoBundlerProxy:                artifactRegistry.Key("package-registry-proxy-bundler-url").String(),
		HermetoCargoProxy:                  artifactRegistry.Key("package-registry-proxy-cargo-url").String(),
		HermetoGomodProxy:                  artifactRegistry.Key("package-registry-proxy-gomod-url").String(),
		HermetoNpmProxy:                    artifactRegistry.Key("package-registry-proxy-npm-url").String(),
		HermetoPipProxy:                    artifactRegistry.Key("package-registry-proxy-pip-url").String(),
		HermetoPnpmProxy:                   artifactRegistry.Key("package-registry-proxy-pnpm-url").String(),
		HermetoYarnProxy:                   artifactRegistry.Key("package-registry-proxy-yarn-url").String(),
		HermetoPackageRegistryProxyAllowed: artifactRegistry.Key("allow-package-registry-proxy").String(),
		FulcioInternalURL:                  keylessSigning.Key("fulcioInternalUrl").String(),
		FulcioExternalURL:                  keylessSigning.Key("fulcioExternalUrl").String(),
		RekorInternalURL:                   keylessSigning.Key("rekorInternalUrl").String(),
		RekorExternalURL:                   keylessSigning.Key("rekorExternalUrl").String(),
		TufInternalURL:                     keylessSigning.Key("tufInternalUrl").String(),
		TufExternalURL:                     keylessSigning.Key("tufExternalUrl").String(),
		DefaultOIDCIssuer:                  keylessSigning.Key("defaultOIDCIssuer").String(),
		EnableKeylessSigning:               keylessSigning.Key("enableKeylessSigning").String(),
	}
	rawConfig.LogData()
	return rawConfig, nil
}
