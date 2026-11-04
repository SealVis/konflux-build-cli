package config

import (
	"fmt"
	"strconv"

	"github.com/konflux-ci/konflux-build-cli/pkg/logger"
)

// https://github.com/konflux-ci/konflux-ci/blob/main/integrations/sigstore/install.sh
type KeylessSigningConfig struct {
	Enabled           bool
	FulcioInternalURL string
	FulcioExternalURL string
	RekorInternalURL  string
	RekorExternalURL  string
	TufInternalURL    string
	TufExternalURL    string
	OIDCIssuer        string
}

var _ KonfluxConfigPart[*KeylessSigningConfig] = (*KeylessSigningConfig)(nil)

func (c *KeylessSigningConfig) GetRekorURL() string {
	if c.RekorInternalURL != "" {
		return c.RekorInternalURL
	}
	return c.RekorExternalURL
}

func (c *KeylessSigningConfig) GetFulcioURL() string {
	if c.FulcioInternalURL != "" {
		return c.FulcioInternalURL
	}
	return c.FulcioExternalURL
}

func (c *KeylessSigningConfig) GetTufURL() string {
	if c.TufInternalURL != "" {
		return c.TufInternalURL
	}
	return c.TufExternalURL
}

func (c *KeylessSigningConfig) IsValid() bool {
	hasFulcioURL := c.GetFulcioURL() != ""
	hasRekorURL := c.GetRekorURL() != ""
	hasTufURL := c.GetTufURL() != ""
	hasOIDCIssuer := c.OIDCIssuer != ""
	return hasFulcioURL && hasRekorURL && hasTufURL && hasOIDCIssuer
}

func NewKeylessSigningConfig(rawConfig KonfluxRawConfig) *KeylessSigningConfig {
	enabled, err := strconv.ParseBool(rawConfig.EnableKeylessSigning)
	if err != nil {
		logger.Logger.Errorf("Failed to parse boolean value for enableKeylessSigning: %v", err)
		enabled = false
	}
	return &KeylessSigningConfig{
		Enabled:           enabled,
		FulcioInternalURL: rawConfig.FulcioInternalURL,
		FulcioExternalURL: rawConfig.FulcioExternalURL,
		RekorInternalURL:  rawConfig.RekorInternalURL,
		RekorExternalURL:  rawConfig.RekorExternalURL,
		TufInternalURL:    rawConfig.TufInternalURL,
		TufExternalURL:    rawConfig.TufExternalURL,
		OIDCIssuer:        rawConfig.DefaultOIDCIssuer,
	}
}

func (c *KeylessSigningConfig) DeepCopy() *KeylessSigningConfig {
	return &KeylessSigningConfig{
		Enabled:           c.Enabled,
		FulcioInternalURL: c.FulcioInternalURL,
		FulcioExternalURL: c.FulcioExternalURL,
		RekorInternalURL:  c.RekorInternalURL,
		RekorExternalURL:  c.RekorExternalURL,
		TufInternalURL:    c.TufInternalURL,
		TufExternalURL:    c.TufExternalURL,
		OIDCIssuer:        c.OIDCIssuer,
	}
}

func (c *KeylessSigningConfig) ToString() string {
	return fmt.Sprintf("%+v", c)
}
