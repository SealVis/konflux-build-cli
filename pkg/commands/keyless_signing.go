package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/konflux-ci/konflux-build-cli/pkg/cliwrappers"
	"github.com/konflux-ci/konflux-build-cli/pkg/config"
	"github.com/konflux-ci/konflux-build-cli/pkg/logger"
	"github.com/spf13/cobra"
)

type KeylessSigning struct {
	CosignCli cliwrappers.CosignCliInterface
}

type KeylessSigningResult struct {
	Enabled    bool   `json:"enabled"`
	FulcioURL  string `json:"fulcio-url,omitempty"`
	RekorURL   string `json:"rekor-url,omitempty"`
	OIDCIssuer string `json:"oidc-issuer,omitempty"`
}

func NewKeylessSigning(cmd *cobra.Command) (*KeylessSigning, error) {
	executor := cliwrappers.NewCliExecutor()
	cosignCli, err := cliwrappers.NewCosignCli(executor)
	if err != nil {
		return nil, err
	}

	return &KeylessSigning{CosignCli: cosignCli}, nil
}

func (k *KeylessSigningResult) printJSON() error {
	data, err := json.Marshal(k)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (k *KeylessSigning) Run() error {
	konfluxConfig, err := config.GetKonfluxConfig()
	if err != nil {
		return err
	}

	result := &KeylessSigningResult{Enabled: false}

	config := konfluxConfig.KeylessSigning
	if !config.Enabled {
		logger.Logger.Warn("Keyless signing is not enabled")
		return result.printJSON()
	}

	if !config.IsValid() {
		logger.Logger.Warn("Keyless signing is not configured correctly")
		return result.printJSON()
	}

	fulcioURL := config.GetFulcioURL()
	logger.Logger.Debugf("Using Fulcio URL: %s", fulcioURL)
	rekorURL := config.GetRekorURL()
	logger.Logger.Debugf("Using Rekor URL: %s", rekorURL)
	tufURL := strings.TrimSuffix(config.GetTufURL(), "/")
	logger.Logger.Debugf("Using TUF URL: %s", tufURL)
	oidcIssuer := config.OIDCIssuer
	logger.Logger.Debugf("Using OIDC Issuer: %s", oidcIssuer)

	args := &cliwrappers.CosignInitializeArgs{
		Root:   fmt.Sprintf("%s/root.json", tufURL),
		Mirror: tufURL,
	}
	if err := k.CosignCli.Initialize(args); err != nil {
		return err
	}

	result.Enabled = true
	result.OIDCIssuer = oidcIssuer
	result.RekorURL = rekorURL
	result.FulcioURL = fulcioURL
	return result.printJSON()
}
