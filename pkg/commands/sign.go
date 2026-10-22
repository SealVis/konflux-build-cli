package commands

import (
	"github.com/spf13/cobra"

	"reflect"

	"github.com/konflux-ci/konflux-build-cli/pkg/cliwrappers"
	"github.com/konflux-ci/konflux-build-cli/pkg/common"
)

var SignParamsConfig = map[string]common.Parameter{
	"image": {
		Name:       "image",
		TypeKind:   reflect.String,
		EnvVarName: "KBC_SIGN_IMAGE",
		Usage:      "Container image to sign",
		Required:   true,
	},
	"rekor-url": {
		Name:       "rekor-url",
		TypeKind:   reflect.String,
		EnvVarName: "KBC_SIGN_REKOR_URL",
		Usage:      "Rekor transparency log URL",
		Required:   true,
	},
	"fulcio-url": {
		Name:       "fulcio-url",
		TypeKind:   reflect.String,
		EnvVarName: "KBC_SIGN_FULCIO_URL",
		Usage:      "Fulcio certificate authority URL",
		Required:   true,
	},
	"oidc-issuer": {
		Name:       "oidc-issuer",
		TypeKind:   reflect.String,
		EnvVarName: "KBC_SIGN_OIDC_ISSUER",
		Usage:      "OIDC issuer",
		Required:   true,
	},
	"oidc-token-path": {
		Name:       "oidc-token-path",
		TypeKind:   reflect.String,
		EnvVarName: "KBC_SIGN_OIDC_TOKEN_PATH",
		Usage:      "OIDC token path",
		Required:   true,
	},
}

type SignParams struct {
	Image         string `paramName:"image"`
	RekorURL      string `paramName:"rekor-url"`
	FulcioURL     string `paramName:"fulcio-url"`
	OIDCIssuer    string `paramName:"oidc-issuer"`
	OIDCTokenPath string `paramName:"oidc-token-path"`
}

type Sign struct {
	Params    *SignParams
	CosignCli *cliwrappers.CosignCli
}

func NewSignCommand(cmd *cobra.Command) (*Sign, error) {
	params := SignParams{}
	if err := common.ParseParameters(cmd, SignParamsConfig, &params); err != nil {
		return nil, err
	}

	executor := cliwrappers.NewCliExecutor()
	cosignCli, err := cliwrappers.NewCosignCli(executor)
	if err != nil {
		return nil, err
	}

	return &Sign{Params: &params, CosignCli: cosignCli}, nil
}

func (s *Sign) Run() error {
	return nil
}
