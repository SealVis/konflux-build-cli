package commands

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/konflux-ci/konflux-build-cli/pkg/cliwrappers"
	"github.com/konflux-ci/konflux-build-cli/pkg/config"

	. "github.com/onsi/gomega"
)

func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out), err
}

func TestKeylessSigningResult(t *testing.T) {
	g := NewWithT(t)

	t.Run("omits empty fields when disabled", func(t *testing.T) {
		data, err := json.Marshal(&KeylessSigningResult{Enabled: false})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(data).To(MatchJSON(`{"enabled":false}`))
	})

	t.Run("includes configured fields when enabled", func(t *testing.T) {
		data, err := json.Marshal(&KeylessSigningResult{
			Enabled:    true,
			FulcioURL:  "https://fulcio.example.com",
			RekorURL:   "https://rekor.example.com",
			OIDCIssuer: "https://oidc.example.com",
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(data).To(MatchJSON(`{
			"enabled": true,
			"fulcio-url": "https://fulcio.example.com",
			"rekor-url": "https://rekor.example.com",
			"oidc-issuer": "https://oidc.example.com"
		}`))
	})
}

func TestKeylessSigningRun(t *testing.T) {
	g := NewWithT(t)

	const (
		fulcioURL  = "https://fulcio.example.com"
		rekorURL   = "https://rekor.example.com"
		tufURL     = "https://tuf.example.com"
		oidcIssuer = "https://oidc.example.com"
	)

	originalGetKonfluxConfig := config.GetKonfluxConfig
	t.Cleanup(func() { config.GetKonfluxConfig = originalGetKonfluxConfig })

	t.Run("prints disabled when keyless signing is not enabled", func(t *testing.T) {
		config.GetKonfluxConfig = func() (*config.KonfluxConfig, error) {
			return &config.KonfluxConfig{
				KeylessSigning: &config.KeylessSigningConfig{
					Enabled:           false,
					FulcioInternalURL: fulcioURL,
					RekorInternalURL:  rekorURL,
					TufInternalURL:    tufURL,
					OIDCIssuer:        oidcIssuer,
				},
			}, nil
		}

		initializeCalled := false
		k := &KeylessSigning{
			CosignCli: &mockCosignCli{
				InitializeFunc: func(args *cliwrappers.CosignInitializeArgs) error {
					initializeCalled = true
					return nil
				},
			},
		}

		stdout, err := captureStdout(k.Run)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(initializeCalled).To(BeFalse())
		g.Expect(stdout).To(MatchJSON(`{"enabled":false}`))
	})

	t.Run("prints disabled when keyless signing config is incomplete", func(t *testing.T) {
		config.GetKonfluxConfig = func() (*config.KonfluxConfig, error) {
			return &config.KonfluxConfig{
				KeylessSigning: &config.KeylessSigningConfig{
					Enabled:        true,
					TufInternalURL: tufURL,
				},
			}, nil
		}

		initializeCalled := false
		k := &KeylessSigning{
			CosignCli: &mockCosignCli{
				InitializeFunc: func(args *cliwrappers.CosignInitializeArgs) error {
					initializeCalled = true
					return nil
				},
			},
		}

		stdout, err := captureStdout(k.Run)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(initializeCalled).To(BeFalse())
		g.Expect(stdout).To(MatchJSON(`{"enabled":false}`))
	})

	t.Run("prints enabled result when keyless signing config is complete", func(t *testing.T) {
		config.GetKonfluxConfig = func() (*config.KonfluxConfig, error) {
			return &config.KonfluxConfig{
				KeylessSigning: &config.KeylessSigningConfig{
					Enabled:           true,
					FulcioInternalURL: fulcioURL,
					RekorInternalURL:  rekorURL,
					TufInternalURL:    tufURL + "/",
					OIDCIssuer:        oidcIssuer,
				},
			}, nil
		}

		var cosignArgs *cliwrappers.CosignInitializeArgs
		k := &KeylessSigning{
			CosignCli: &mockCosignCli{
				InitializeFunc: func(args *cliwrappers.CosignInitializeArgs) error {
					cosignArgs = args
					return nil
				},
			},
		}

		stdout, err := captureStdout(k.Run)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(cosignArgs).ToNot(BeNil())
		g.Expect(cosignArgs.Root).To(Equal(tufURL + "/root.json"))
		g.Expect(cosignArgs.Mirror).To(Equal(tufURL))

		var result KeylessSigningResult
		g.Expect(json.Unmarshal([]byte(stdout), &result)).To(Succeed())
		g.Expect(result.Enabled).To(BeTrue())
		g.Expect(result.FulcioURL).To(Equal(fulcioURL))
		g.Expect(result.RekorURL).To(Equal(rekorURL))
		g.Expect(result.OIDCIssuer).To(Equal(oidcIssuer))
	})

	t.Run("returns error when config cannot be read", func(t *testing.T) {
		config.GetKonfluxConfig = func() (*config.KonfluxConfig, error) {
			return nil, errors.New("oc error")
		}

		initializeCalled := false
		k := &KeylessSigning{
			CosignCli: &mockCosignCli{
				InitializeFunc: func(args *cliwrappers.CosignInitializeArgs) error {
					initializeCalled = true
					return nil
				},
			},
		}

		stdout, err := captureStdout(k.Run)
		g.Expect(err).To(MatchError("oc error"))
		g.Expect(initializeCalled).To(BeFalse())
		g.Expect(stdout).To(BeEmpty())
	})

	t.Run("returns error when cosign initialize fails", func(t *testing.T) {
		config.GetKonfluxConfig = func() (*config.KonfluxConfig, error) {
			return &config.KonfluxConfig{
				KeylessSigning: &config.KeylessSigningConfig{
					Enabled:           true,
					FulcioInternalURL: fulcioURL,
					RekorInternalURL:  rekorURL,
					TufInternalURL:    tufURL,
					OIDCIssuer:        oidcIssuer,
				},
			}, nil
		}

		k := &KeylessSigning{
			CosignCli: &mockCosignCli{
				InitializeFunc: func(args *cliwrappers.CosignInitializeArgs) error {
					return errors.New("cosign initialize error")
				},
			},
		}

		stdout, err := captureStdout(k.Run)
		g.Expect(err).To(MatchError("cosign initialize error"))
		g.Expect(stdout).To(BeEmpty())
	})
}
