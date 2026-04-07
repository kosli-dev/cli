package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/charmbracelet/lipgloss"
	"github.com/kosli-dev/cli/internal/requests"
)

// ---------------------------------------------------------------------------
// Wizard context (API data)
// ---------------------------------------------------------------------------

type wizardContext struct {
	flowNames         []string
	customAttestTypes []string
}

func (wctx *wizardContext) fetchFromAPI() {
	wctx.flowNames = fetchFlowNames()
	wctx.customAttestTypes = fetchCustomAttestationTypes()
}

func fetchFlowNames() []string {
	u, err := url.JoinPath(global.Host, "api/v2/flows", global.Org)
	if err != nil {
		logger.Debug("failed to build flows URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch flows: %v", err)
		return nil
	}

	var flows []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &flows); err != nil {
		logger.Debug("failed to parse flows response: %v", err)
		return nil
	}

	names := make([]string, 0, len(flows))
	for _, flow := range flows {
		if name, ok := flow["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func fetchCustomAttestationTypes() []string {
	u, err := url.JoinPath(global.Host, "api/v2/custom-attestation-types", global.Org)
	if err != nil {
		logger.Debug("failed to build attestation types URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch attestation types: %v", err)
		return nil
	}

	var types []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &types); err != nil {
		logger.Debug("failed to parse attestation types response: %v", err)
		return nil
	}

	names := make([]string, 0, len(types))
	for _, t := range types {
		if name, ok := t["name"].(string); ok {
			names = append(names, "custom:"+name)
		}
	}
	return names
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

type wizardStyles struct {
	base        lipgloss.Style
	title       lipgloss.Style
	preview     lipgloss.Style
	previewText lipgloss.Style
	footer      lipgloss.Style
	accent      lipgloss.Style
}

func newWizardStyles() wizardStyles {
	accent := lipgloss.Color("#7571F9")
	green := lipgloss.Color("#02BF87")
	return wizardStyles{
		base: lipgloss.NewStyle().Padding(1, 2),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Padding(0, 1),
		preview: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2),
		previewText: lipgloss.NewStyle().
			Foreground(green),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 1, 0, 1),
		accent: lipgloss.NewStyle().
			Foreground(accent),
	}
}
