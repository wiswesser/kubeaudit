package sarif

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Shopify/kubeaudit"
	"github.com/Shopify/kubeaudit/auditors/apparmor"
	"github.com/Shopify/kubeaudit/auditors/capabilities"
	"github.com/Shopify/kubeaudit/auditors/image"
	"github.com/Shopify/kubeaudit/auditors/limits"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWithResults(t *testing.T) {
	capabilitiesAuditable := capabilities.New(capabilities.Config{})
	apparmorAuditable := apparmor.New()
	imageAuditable := image.New(image.Config{Image: "scratch:1.5"})
	limitsAuditable, _ := limits.New(limits.Config{})

	cases := []struct {
		file               string
		auditors           []kubeaudit.Auditable
		expectedRule       string
		expectedErrorLevel string
		expectedMessage    string
		expectedURI        string
	}{
		{
			"apparmor-invalid.yaml",
			[]kubeaudit.Auditable{apparmorAuditable},
			apparmor.AppArmorInvalidAnnotation,
			"error",
			"AppArmor annotation key refers to a container that doesn't exist",
			"https://github.com/Shopify/kubeaudit/blob/main/docs/auditors/apparmor.md",
		},
		{
			"capabilities-added.yaml",
			[]kubeaudit.Auditable{capabilitiesAuditable},
			capabilities.CapabilityAdded,
			"error",
			"It should be removed from the capability add list",
			"https://github.com/Shopify/kubeaudit/blob/main/docs/auditors/capabilities.md",
		},
		{
			"image-tag-present.yaml",
			[]kubeaudit.Auditable{imageAuditable},
			image.ImageCorrect,
			"note",
			"Image tag is correct",
			"https://github.com/Shopify/kubeaudit/blob/main/docs/auditors/image.md",
		},
		{
			"limits-nil.yaml",
			[]kubeaudit.Auditable{limitsAuditable},
			limits.LimitsNotSet,
			"warning",
			"Resource limits not set",
			"https://github.com/Shopify/kubeaudit/blob/main/docs/auditors/limits.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			fixture := filepath.Join("fixtures", tc.file)
			auditor, err := kubeaudit.New(tc.auditors)
			require.NoError(t, err)

			manifest, openErr := os.Open(fixture)
			require.NoError(t, openErr)

			defer manifest.Close()

			kubeAuditReport, err := auditor.AuditManifest(fixture, manifest)
			require.NoError(t, err)

			sarifReport, err := Create(kubeAuditReport)
			require.NoError(t, err)

			assert.Equal(t, repoURL,
				*sarifReport.Runs[0].Tool.Driver.InformationURI)

			// verify that the rules have been added as per report findings
			assert.Equal(t, tc.expectedRule, sarifReport.Runs[0].Tool.Driver.Rules[0].ID)

			var ruleNames []string

			// check for rules occurrences
			for _, sarifRule := range sarifReport.Runs[0].Tool.Driver.Rules {
				assert.Equal(t, []string{
					"security",
					"kubernetes",
					"infrastructure",
				},
					sarifRule.Properties["tags"],
				)

				ruleNames = append(ruleNames, sarifRule.ID)

				assert.Contains(t, *sarifRule.Help.Text, tc.expectedURI)
			}

			for _, sarifResult := range sarifReport.Runs[0].Results {
				assert.Contains(t, ruleNames, *sarifResult.RuleID)
				assert.Equal(t, tc.expectedErrorLevel, *sarifResult.Level)
				assert.Contains(t, *sarifResult.Message.Text, tc.expectedMessage)
				assert.Contains(t, "sarif/fixtures/"+tc.file, *sarifResult.Locations[0].PhysicalLocation.ArtifactLocation.URI)
			}
		})
	}
}

func TestCreateWithNoResults(t *testing.T) {
	apparmorAuditable := apparmor.New()

	fixture := filepath.Join("fixtures", "apparmor-valid.yaml")
	auditor, err := kubeaudit.New([]kubeaudit.Auditable{apparmorAuditable})
	require.NoError(t, err)

	manifest, openErr := os.Open(fixture)
	require.NoError(t, openErr)

	defer manifest.Close()

	kubeAuditReport, err := auditor.AuditManifest(fixture, manifest)
	require.NoError(t, err)

	sarifReport, err := Create(kubeAuditReport)
	require.NoError(t, err)

	require.NotEmpty(t, *sarifReport.Runs[0])

	// verify that the rules are only added as per report findings
	assert.Len(t, sarifReport.Runs[0].Tool.Driver.Rules, 0)
}
