package cmd

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/signal"

	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	hyperp "github.com/swadhinbiswas/ghost/internal/agent/hyper"
	"github.com/swadhinbiswas/ghost/internal/config"
	"github.com/swadhinbiswas/ghost/internal/oauth"
	"github.com/swadhinbiswas/ghost/internal/oauth/copilot"
	"github.com/swadhinbiswas/ghost/internal/oauth/hyper"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Aliases: []string{"auth"},
	Use:     "login [platform]",
	Short:   "Login Ghost to a platform or save an API key",
	Long: `Login Ghost to a specified platform or save an API key for any AI provider.
The platform or provider ID should be provided as an argument.
If the provider uses OAuth (like hyper or copilot), it will initiate the OAuth flow.
Otherwise, it will securely prompt you for your API key and save it to your local configuration store.`,
	Example: `
# Authenticate with Charm Hyper
Ghost login

# Authenticate with GitHub Copilot
Ghost login copilot
  `,
	ValidArgs: []cobra.Completion{
		"hyper",
		"copilot",
		"github-copilot",
		"nvidia-nim",
		"openai",
		"anthropic",
		"google",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setupAppWithProgressBar(cmd)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		provider := "hyper"
		if len(args) > 0 {
			provider = args[0]
		}
		switch provider {
		case "hyper":
			return loginHyper(app.Store())
		case "copilot", "github", "github-copilot":
			return loginCopilot(app.Store())
		default:
			return loginAPIKey(app.Store(), provider)
		}
	},
}

func loginHyper(cfg *config.ConfigStore) error {
	if !hyperp.Enabled() {
		return fmt.Errorf("hyper not enabled")
	}
	ctx := getLoginContext()

	resp, err := hyper.InitiateDeviceAuth(ctx)
	if err != nil {
		return err
	}

	if clipboard.WriteAll(resp.UserCode) == nil {
		fmt.Println("The following code should be on clipboard already:")
	} else {
		fmt.Println("Copy the following code:")
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render(resp.UserCode))
	fmt.Println()
	fmt.Println("Press enter to open this URL, and then paste it there:")
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Hyperlink(resp.VerificationURL, "id=hyper").Render(resp.VerificationURL))
	fmt.Println()
	waitEnter()
	if err := browser.OpenURL(resp.VerificationURL); err != nil {
		fmt.Println("Could not open the URL. You'll need to manually open the URL in your browser.")
	}

	fmt.Println("Exchanging authorization code...")
	refreshToken, err := hyper.PollForToken(ctx, resp.DeviceCode, resp.ExpiresIn)
	if err != nil {
		return err
	}

	fmt.Println("Exchanging refresh token for access token...")
	token, err := hyper.ExchangeToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	fmt.Println("Verifying access token...")
	introspect, err := hyper.IntrospectToken(ctx, token.AccessToken)
	if err != nil {
		return fmt.Errorf("token introspection failed: %w", err)
	}
	if !introspect.Active {
		return fmt.Errorf("access token is not active")
	}

	if err := cmp.Or(
		cfg.SetConfigField(config.ScopeGlobal, "providers.hyper.api_key", token.AccessToken),
		cfg.SetConfigField(config.ScopeGlobal, "providers.hyper.oauth", token),
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with Hyper!")
	return nil
}

func loginCopilot(cfg *config.ConfigStore) error {
	ctx := getLoginContext()

	if cfg.HasConfigField(config.ScopeGlobal, "providers.copilot.oauth") {
		fmt.Println("You are already logged in to GitHub Copilot.")
		return nil
	}

	diskToken, hasDiskToken := copilot.RefreshTokenFromDisk()
	var token *oauth.Token

	switch {
	case hasDiskToken:
		fmt.Println("Found existing GitHub Copilot token on disk. Using it to authenticate...")

		t, err := copilot.RefreshToken(ctx, diskToken)
		if err != nil {
			return fmt.Errorf("unable to refresh token from disk: %w", err)
		}
		token = t
	default:
		fmt.Println("Requesting device code from GitHub...")
		dc, err := copilot.RequestDeviceCode(ctx)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println("Open the following URL and follow the instructions to authenticate with GitHub Copilot:")
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Hyperlink(dc.VerificationURI, "id=copilot").Render(dc.VerificationURI))
		fmt.Println()
		fmt.Println("Code:", lipgloss.NewStyle().Bold(true).Render(dc.UserCode))
		fmt.Println()
		fmt.Println("Waiting for authorization...")

		t, err := copilot.PollForToken(ctx, dc)
		if err == copilot.ErrNotAvailable {
			fmt.Println()
			fmt.Println("GitHub Copilot is unavailable for this account. To signup, go to the following page:")
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Hyperlink(copilot.SignupURL, "id=copilot-signup").Render(copilot.SignupURL))
			fmt.Println()
			fmt.Println("You may be able to request free access if eligible. For more information, see:")
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Hyperlink(copilot.FreeURL, "id=copilot-free").Render(copilot.FreeURL))
		}
		if err != nil {
			return err
		}
		token = t
	}

	if err := cmp.Or(
		cfg.SetConfigField(config.ScopeGlobal, "providers.copilot.api_key", token.AccessToken),
		cfg.SetConfigField(config.ScopeGlobal, "providers.copilot.oauth", token),
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with GitHub Copilot!")
	return nil
}

func getLoginContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	go func() {
		<-ctx.Done()
		cancel()
		os.Exit(1)
	}()
	return ctx
}

func waitEnter() {
	_, _ = fmt.Scanln()
}

func loginAPIKey(cfg *config.ConfigStore, provider string) error {
	fmt.Printf("Enter API key for %s: ", provider)
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println()

	apiKey := string(keyBytes)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	err = cfg.SetProviderAPIKey(config.ScopeGlobal, provider, apiKey)
	if err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	fmt.Printf("Successfully saved API key for provider '%s' to configuration store.\n", provider)
	return nil
}
