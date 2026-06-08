package sohacli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	sohaapi "github.com/opensoha/soha-contracts/gen/go/sohaapi"
)

const userAgent = "soha/0.1"

type APIClient struct {
	ServerURL string
	Token     string
	Client    *http.Client
}

type loginResponse = sohaapi.AuthResultEnvelope

type manifestResponse = sohaapi.AIGatewayManifestEnvelope

type invokeResponse = sohaapi.ToolInvocationResultEnvelope

type resourceReadResponse = sohaapi.ResourceReadResultEnvelope

type promptGetResponse = sohaapi.PromptGetResultEnvelope

type itemsResponse[T any] struct {
	Items []T `json:"items"`
}

type itemResponse[T any] struct {
	Data T `json:"data"`
}

type Manifest = sohaapi.AIGatewayManifest

type ToolCapability = sohaapi.ToolCapability

type ResourceCapability = sohaapi.ResourceCapability

type PromptCapability = sohaapi.PromptCapability

type SkillCapability = sohaapi.SkillCapability

type ToolInvocationResult = sohaapi.ToolInvocationResult

type ResourceReadResult = sohaapi.ResourceReadResult

type PromptMessage = sohaapi.PromptMessage

type PromptGetResult = sohaapi.PromptGetResult

type PersonalAccessToken = sohaapi.PersonalAccessToken

type CreatedPersonalAccessToken = sohaapi.CreatedPersonalAccessToken

type ServiceAccount = sohaapi.ServiceAccount

type ServiceAccountToken = sohaapi.ServiceAccountToken

type CreatedServiceAccountToken = sohaapi.CreatedServiceAccountToken

type AuditLog = sohaapi.AuditLog

type ApprovalTrace = sohaapi.ApprovalTrace

type ApprovalDecisionTrace = sohaapi.ApprovalDecisionTrace

type ApprovalStageTrace = sohaapi.ApprovalStageTrace

type ApprovalTimelineEvent = sohaapi.ApprovalTimelineEvent

type ApprovalRequest = sohaapi.ApprovalRequest

type ApprovalDecisionResult = sohaapi.ApprovalDecisionResult

type ApprovalTimeline = sohaapi.ApprovalTimeline

type GovernanceStatus = sohaapi.GovernanceStatus

type GovernanceHealth = sohaapi.GovernanceHealth

type GovernanceHealthCheck = sohaapi.GovernanceHealthCheck

type GovernanceMetrics = sohaapi.GovernanceMetrics

type GovernanceMetricCount = sohaapi.GovernanceMetricCount

type GovernanceRedactionSummary = sohaapi.GovernanceRedactionSummary

type GovernanceTokenSummary = sohaapi.GovernanceTokenSummary

type GovernanceTokenCounts = sohaapi.GovernanceTokenCounts

type GovernanceTokenFinding = sohaapi.GovernanceTokenFinding

type GovernanceClientSummary = sohaapi.GovernanceClientSummary

type GovernanceApprovalSummary = sohaapi.GovernanceApprovalSummary

type GovernancePolicyCoverage = sohaapi.GovernancePolicyCoverage

type GovernanceFinding = sohaapi.GovernanceFinding

type GovernanceRecommendationAction = sohaapi.GovernanceRecommendationAction

type PluginManifest = sohaapi.PluginManifest

type MarketplacePlugin = sohaapi.MarketplacePlugin

type InstalledPlugin = sohaapi.InstalledPlugin

type PluginInstallRequest = sohaapi.PluginInstallRequest

type PluginConfigRequest = sohaapi.PluginConfigRequest

func (c APIClient) Login(ctx context.Context, login, password string) (loginResponse, error) {
	var out loginResponse
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/auth/login", "", nil, map[string]string{
		"login":    login,
		"password": password,
	}, &out)
	return out, err
}

func (c APIClient) Capabilities(ctx context.Context, headers map[string]string) (Manifest, error) {
	var out manifestResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/ai-gateway/capabilities", c.Token, headers, nil, &out); err != nil {
		return Manifest{}, err
	}
	return out.Data, nil
}

func (c APIClient) InvokeTool(ctx context.Context, toolName string, input map[string]any, headers map[string]string) (ToolInvocationResult, error) {
	var out invokeResponse
	path := "/api/v1/ai-gateway/tools/" + url.PathEscape(toolName) + "/invoke"
	payload := map[string]any{"input": emptyInput(input)}
	if err := c.doJSON(ctx, http.MethodPost, path, c.Token, headers, payload, &out); err != nil {
		return ToolInvocationResult{}, err
	}
	return out.Data, nil
}

func (c APIClient) ReadResource(ctx context.Context, uri string, contextValues map[string]any, headers map[string]string) (ResourceReadResult, error) {
	var out resourceReadResponse
	payload := map[string]any{"uri": strings.TrimSpace(uri), "context": emptyInput(contextValues)}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/ai-gateway/resources/read", c.Token, headers, payload, &out); err != nil {
		return ResourceReadResult{}, err
	}
	return out.Data, nil
}

func (c APIClient) GetPrompt(ctx context.Context, name string, arguments map[string]any, contextValues map[string]any, headers map[string]string) (PromptGetResult, error) {
	var out promptGetResponse
	payload := map[string]any{"name": strings.TrimSpace(name), "arguments": emptyInput(arguments), "context": emptyInput(contextValues)}
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/ai-gateway/prompts/get", c.Token, headers, payload, &out); err != nil {
		return PromptGetResult{}, err
	}
	return out.Data, nil
}

func (c APIClient) ListPersonalAccessTokens(ctx context.Context, headers map[string]string) ([]PersonalAccessToken, error) {
	var out itemsResponse[PersonalAccessToken]
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/ai-gateway/personal-access-tokens", c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) CreatePersonalAccessToken(ctx context.Context, input map[string]any, headers map[string]string) (CreatedPersonalAccessToken, error) {
	var out itemResponse[CreatedPersonalAccessToken]
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/ai-gateway/personal-access-tokens", c.Token, headers, input, &out); err != nil {
		return CreatedPersonalAccessToken{}, err
	}
	return out.Data, nil
}

func (c APIClient) RevokePersonalAccessToken(ctx context.Context, tokenID string, headers map[string]string) error {
	path := "/api/v1/ai-gateway/personal-access-tokens/" + url.PathEscape(tokenID) + "/revoke"
	return c.doJSON(ctx, http.MethodPost, path, c.Token, headers, nil, nil)
}

func (c APIClient) ListServiceAccounts(ctx context.Context, headers map[string]string) ([]ServiceAccount, error) {
	var out itemsResponse[ServiceAccount]
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/ai-gateway/service-accounts", c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) ListServiceAccountTokens(ctx context.Context, headers map[string]string) ([]ServiceAccountToken, error) {
	var out itemsResponse[ServiceAccountToken]
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/ai-gateway/service-account-tokens", c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) CreateServiceAccount(ctx context.Context, input map[string]any, headers map[string]string) (ServiceAccount, error) {
	var out itemResponse[ServiceAccount]
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/ai-gateway/service-accounts", c.Token, headers, input, &out); err != nil {
		return ServiceAccount{}, err
	}
	return out.Data, nil
}

func (c APIClient) CreateServiceAccountToken(ctx context.Context, serviceAccountID string, input map[string]any, headers map[string]string) (CreatedServiceAccountToken, error) {
	var out itemResponse[CreatedServiceAccountToken]
	path := "/api/v1/ai-gateway/service-accounts/" + url.PathEscape(serviceAccountID) + "/tokens"
	if err := c.doJSON(ctx, http.MethodPost, path, c.Token, headers, input, &out); err != nil {
		return CreatedServiceAccountToken{}, err
	}
	return out.Data, nil
}

func (c APIClient) RevokeServiceAccountToken(ctx context.Context, tokenID string, headers map[string]string) error {
	path := "/api/v1/ai-gateway/service-account-tokens/" + url.PathEscape(tokenID) + "/revoke"
	return c.doJSON(ctx, http.MethodPost, path, c.Token, headers, nil, nil)
}

func (c APIClient) ListAuditLogs(ctx context.Context, query url.Values, headers map[string]string) ([]AuditLog, error) {
	var out itemsResponse[AuditLog]
	path := "/api/v1/ai-gateway/audit-logs"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) ListApprovalRequests(ctx context.Context, query url.Values, headers map[string]string) ([]ApprovalRequest, error) {
	var out itemsResponse[ApprovalRequest]
	path := "/api/v1/ai-gateway/approval-requests"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) GetApprovalTimeline(ctx context.Context, requestID string, headers map[string]string) (ApprovalTimeline, error) {
	var out itemResponse[ApprovalTimeline]
	path := "/api/v1/ai-gateway/approval-requests/" + url.PathEscape(strings.TrimSpace(requestID)) + "/timeline"
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return ApprovalTimeline{}, err
	}
	return out.Data, nil
}

func (c APIClient) DecideApprovalRequest(ctx context.Context, requestID, action, comment string, headers map[string]string) (ApprovalDecisionResult, error) {
	action = strings.TrimSpace(action)
	switch action {
	case "approve", "reject", "cancel":
	default:
		return ApprovalDecisionResult{}, fmt.Errorf("unsupported approval action %q", action)
	}
	var out itemResponse[ApprovalDecisionResult]
	path := "/api/v1/ai-gateway/approval-requests/" + url.PathEscape(strings.TrimSpace(requestID)) + "/" + action
	payload := map[string]any{}
	if strings.TrimSpace(comment) != "" {
		payload["comment"] = strings.TrimSpace(comment)
	}
	if err := c.doJSON(ctx, http.MethodPost, path, c.Token, headers, payload, &out); err != nil {
		return ApprovalDecisionResult{}, err
	}
	return out.Data, nil
}

func (c APIClient) GovernanceStatus(ctx context.Context, windowHours int, headers map[string]string) (GovernanceStatus, error) {
	var out itemResponse[GovernanceStatus]
	path := "/api/v1/ai-gateway/governance/status"
	if windowHours > 0 {
		query := url.Values{}
		query.Set("windowHours", fmt.Sprint(windowHours))
		path += "?" + query.Encode()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return GovernanceStatus{}, err
	}
	return out.Data, nil
}

func (c APIClient) ListMarketplacePlugins(ctx context.Context, query url.Values, headers map[string]string) ([]MarketplacePlugin, error) {
	var out itemsResponse[MarketplacePlugin]
	path := "/api/v1/plugins/marketplace"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) GetMarketplacePlugin(ctx context.Context, pluginID string, headers map[string]string) (MarketplacePlugin, error) {
	var out itemResponse[MarketplacePlugin]
	path := "/api/v1/plugins/marketplace/" + url.PathEscape(strings.TrimSpace(pluginID))
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return MarketplacePlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) ListInstalledPlugins(ctx context.Context, headers map[string]string) ([]InstalledPlugin, error) {
	var out itemsResponse[InstalledPlugin]
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/plugins/installed", c.Token, headers, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c APIClient) GetInstalledPlugin(ctx context.Context, pluginID string, headers map[string]string) (InstalledPlugin, error) {
	var out itemResponse[InstalledPlugin]
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID))
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return InstalledPlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) GetInstalledPluginManifest(ctx context.Context, pluginID string, headers map[string]string) (PluginManifest, error) {
	var out itemResponse[PluginManifest]
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID)) + "/manifest"
	if err := c.doJSON(ctx, http.MethodGet, path, c.Token, headers, nil, &out); err != nil {
		return PluginManifest{}, err
	}
	return out.Data, nil
}

func (c APIClient) InstallPlugin(ctx context.Context, input PluginInstallRequest, headers map[string]string) (InstalledPlugin, error) {
	var out itemResponse[InstalledPlugin]
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/plugins/install", c.Token, headers, input, &out); err != nil {
		return InstalledPlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) EnablePlugin(ctx context.Context, pluginID string, headers map[string]string) (InstalledPlugin, error) {
	return c.pluginAction(ctx, pluginID, "enable", headers)
}

func (c APIClient) DisablePlugin(ctx context.Context, pluginID string, headers map[string]string) (InstalledPlugin, error) {
	return c.pluginAction(ctx, pluginID, "disable", headers)
}

func (c APIClient) UpgradePlugin(ctx context.Context, pluginID string, input PluginInstallRequest, headers map[string]string) (InstalledPlugin, error) {
	var out itemResponse[InstalledPlugin]
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID)) + "/upgrade"
	if err := c.doJSON(ctx, http.MethodPost, path, c.Token, headers, input, &out); err != nil {
		return InstalledPlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) ConfigurePlugin(ctx context.Context, pluginID string, input PluginConfigRequest, headers map[string]string) (InstalledPlugin, error) {
	var out itemResponse[InstalledPlugin]
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID)) + "/config"
	if err := c.doJSON(ctx, http.MethodPut, path, c.Token, headers, input, &out); err != nil {
		return InstalledPlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) RemovePlugin(ctx context.Context, pluginID string, headers map[string]string) error {
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID))
	return c.doJSON(ctx, http.MethodDelete, path, c.Token, headers, nil, nil)
}

func (c APIClient) pluginAction(ctx context.Context, pluginID, action string, headers map[string]string) (InstalledPlugin, error) {
	var out itemResponse[InstalledPlugin]
	path := "/api/v1/plugins/" + url.PathEscape(strings.TrimSpace(pluginID)) + "/" + action
	if err := c.doJSON(ctx, http.MethodPost, path, c.Token, headers, nil, &out); err != nil {
		return InstalledPlugin{}, err
	}
	return out.Data, nil
}

func (c APIClient) doJSON(ctx context.Context, method, path, token string, headers map[string]string, body any, out any) error {
	base := strings.TrimRight(strings.TrimSpace(c.ServerURL), "/")
	if base == "" {
		return fmt.Errorf("server URL is required")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for key, value := range headers {
		value = strings.TrimSpace(value)
		if value != "" {
			req.Header.Set(key, value)
		}
	}
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s failed: %s: %s", method, path, resp.Status, responseErrorMessage(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func responseErrorMessage(raw []byte) string {
	var wrapped struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &wrapped) == nil && wrapped.Error.Message != "" {
		if wrapped.Error.Code != "" {
			return wrapped.Error.Code + ": " + wrapped.Error.Message
		}
		return wrapped.Error.Message
	}
	return strings.TrimSpace(string(raw))
}

func emptyInput(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	return input
}
