package securewebproxy_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/securewebproxy"
	"github.com/navikt/nada-backend/pkg/securewebproxy/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()

	e := emulator.New(log)
	url := e.Run()

	client := securewebproxy.New(url, true)

	id := &securewebproxy.URLListIdentifier{
		Project:  "test",
		Location: "europe-north1",
		Slug:     "myurllist",
	}

	urls := []string{"example.com", "github.com/navikt"}

	t.Run("Get URL list that does not exist", func(t *testing.T) {
		_, err := client.GetURLList(ctx, id)
		require.Error(t, err)
		assert.Equal(t, securewebproxy.ErrNotExist, err)
	})

	t.Run("Create URL list", func(t *testing.T) {
		err := client.CreateURLList(ctx, &securewebproxy.URLListCreateOpts{
			ID:          id,
			Description: "My URL list",
			URLS:        urls,
		})
		require.NoError(t, err)
	})

	t.Run("Create URL list that already exists", func(t *testing.T) {
		err := client.CreateURLList(ctx, &securewebproxy.URLListCreateOpts{
			ID:          id,
			Description: "My URL list",
			URLS:        urls,
		})
		require.Error(t, err)
		assert.Equal(t, securewebproxy.ErrExist, err)
	})

	t.Run("Get URL list that exists", func(t *testing.T) {
		got, err := client.GetURLList(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, urls, got)
	})

	updatedURLs := []string{"other.com", "github.com/other"}

	t.Run("Update URL list", func(t *testing.T) {
		err := client.UpdateURLList(ctx, &securewebproxy.URLListCreateOpts{
			ID:          id,
			Description: "My URL list",
			URLS:        updatedURLs,
		})
		require.NoError(t, err)
	})

	t.Run("Get updated URL list", func(t *testing.T) {
		got, err := client.GetURLList(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, updatedURLs, got)
	})

	t.Run("Delete URL list", func(t *testing.T) {
		err := client.DeleteURLList(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Delete URL list that does not exist", func(t *testing.T) {
		err := client.DeleteURLList(ctx, id)
		require.NoError(t, err)
	})
}

func TestPolicyRules(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()

	e := emulator.New(log)
	url := e.Run()

	client := securewebproxy.New(url, true)

	policyRuleID := &securewebproxy.PolicyRuleIdentifier{
		Project:  "test",
		Location: "europe-north1",
		Policy:   "myPolicy",
		Slug:     "myRule",
	}

	policyRule := &securewebproxy.GatewaySecurityPolicyRule{
		SessionMatcher:       "source.matchServiceAccount('my-email-at-something@test.iam.gserviceaccount.com')",
		ApplicationMatcher:   "inUrlList(request.url(), 'projects/test/locations/europe-north1/urlLists/mylist')",
		BasicProfile:         "ALLOW",
		Description:          "My policy rule",
		Enabled:              true,
		Name:                 policyRuleID.FullyQualifiedName(),
		Priority:             1,
		TlsInspectionEnabled: true,
	}

	t.Run("Get policy rule that does not exist", func(t *testing.T) {
		_, err := client.GetSecurityPolicyRule(ctx, policyRuleID)
		require.Error(t, err)
	})

	t.Run("Create policy rule", func(t *testing.T) {
		err := client.CreateSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleCreateOpts{
			ID:   policyRuleID,
			Rule: policyRule,
		})
		require.NoError(t, err)
	})

	t.Run("Create policy rule that exist", func(t *testing.T) {
		err := client.CreateSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleCreateOpts{
			ID:   policyRuleID,
			Rule: policyRule,
		})
		require.Error(t, err)
	})

	t.Run("Get policy rule", func(t *testing.T) {
		got, err := client.GetSecurityPolicyRule(ctx, policyRuleID)
		require.NoError(t, err)

		assert.NotNil(t, got.CreateTime)
		diff := cmp.Diff(policyRule, got, cmpopts.IgnoreFields(securewebproxy.GatewaySecurityPolicyRule{}, "CreateTime"))
		assert.Empty(t, diff)
	})

	updatedPolicyRule := &securewebproxy.GatewaySecurityPolicyRule{
		SessionMatcher:       "source.matchServiceAccount('another-email@test.iam.gserviceaccount.com')",
		ApplicationMatcher:   "inUrlList(request.url(), 'projects/test/locations/europe-north1/urlLists/anotherList')",
		BasicProfile:         "ALLOW",
		Description:          "My policy rule",
		Enabled:              true,
		Name:                 policyRuleID.FullyQualifiedName(),
		Priority:             1,
		TlsInspectionEnabled: true,
	}

	t.Run("Update policy rule", func(t *testing.T) {
		err := client.UpdateSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleCreateOpts{
			ID:   policyRuleID,
			Rule: updatedPolicyRule,
		})
		require.NoError(t, err)
	})

	t.Run("Get policy rule", func(t *testing.T) {
		got, err := client.GetSecurityPolicyRule(ctx, policyRuleID)
		require.NoError(t, err)

		assert.NotNil(t, got.UpdateTime)
		diff := cmp.Diff(updatedPolicyRule, got, cmpopts.IgnoreFields(securewebproxy.GatewaySecurityPolicyRule{}, "CreateTime", "UpdateTime"))
		assert.Empty(t, diff)
	})

	t.Run("Delete policy rule", func(t *testing.T) {
		err := client.DeleteSecurityPolicyRule(ctx, policyRuleID)
		require.NoError(t, err)
	})
}
