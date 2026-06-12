package sohacli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

func runDiagnose(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("diagnose", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	toolName := fs.String("tool", "", "tool name to inspect")
	resourceName := fs.String("resource", "", "resource URI to inspect")
	promptName := fs.String("prompt", "", "prompt name to inspect")
	clusterCapabilityKey := fs.String("cluster-capability", "", "cluster capability key to inspect")
	aiClientID := fs.String("ai-client-id", "", "override AI client id for this diagnostic request")
	aiClientName := fs.String("ai-client", "", "override AI client display name for this diagnostic request")
	skillID := fs.String("skill-id", "", "override skill id for this diagnostic request")
	source := fs.String("source", "", "override source label for this diagnostic request")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, name, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	manifest, err := gatewayClient(rt, profile).Capabilities(ctx, gatewayHeaders(profile, *aiClientID, *aiClientName, *skillID, *source))
	if err != nil {
		return err
	}
	fmt.Fprintf(rt.Out, "profile: %s\nserver: %s\nuser: %s\n", name, profile.ServerURL, profile.UserName)
	fmt.Fprintf(rt.Out, "tools: %d\nresources: %d\nprompts: %d\nskills: %d\n", len(manifest.Tools), len(manifest.Resources), len(manifest.Prompts), len(manifest.Skills))
	fmt.Fprintf(rt.Out, "permissionKeys: %d\n", len(manifest.PermissionKeys))
	fmt.Fprintf(rt.Out, "aiClientId: %s\naiClient: %s\nskillId: %s\nsource: %s\n", firstNonEmptyString(*aiClientID, profile.AIClientID), firstNonEmptyString(*aiClientName, profile.AIClientName), firstNonEmptyString(*skillID, profile.SkillID), firstNonEmptyString(*source, profile.Source, "soha"))
	if strings.TrimSpace(*toolName) != "" {
		diagnoseTool(rt.Out, manifest, strings.TrimSpace(*toolName))
	}
	if strings.TrimSpace(*resourceName) != "" {
		diagnoseResource(rt.Out, manifest, strings.TrimSpace(*resourceName))
	}
	if strings.TrimSpace(*promptName) != "" {
		diagnosePrompt(rt.Out, manifest, strings.TrimSpace(*promptName))
	}
	if strings.TrimSpace(*clusterCapabilityKey) != "" {
		items, err := gatewayClient(rt, profile).ClusterCapabilities(ctx)
		if err != nil {
			return err
		}
		diagnoseClusterCapability(rt.Out, items, strings.TrimSpace(*clusterCapabilityKey))
	}
	if len(manifest.Tools) == 0 {
		fmt.Fprintln(rt.Out, "hint: no tools visible; check ai.gateway.invoke, MCP tool grants, access policies, and skill bindings.")
	}
	return nil
}

func diagnoseTool(out io.Writer, manifest Manifest, toolName string) {
	for _, tool := range manifest.Tools {
		if tool.Name == toolName {
			fmt.Fprintf(out, "tool: %s\nriskLevel: %s\nrequiresApproval: %t\n", tool.Name, tool.RiskLevel, tool.RequiresApproval)
			fmt.Fprintf(out, "domain: %s\naction: %s\nmcpAdapterId: %s\nmcpToolName: %s\n", tool.Domain, tool.Action, tool.MCPAdapterID, tool.MCPToolName)
			fmt.Fprintf(out, "requiredPermissionKeys: %s\n", strings.Join(tool.PermissionKeys, ","))
			fmt.Fprintf(out, "requiredScopes: %s\n", strings.Join(tool.RequiredScopes, ","))
			required, fields := toolSchemaSummary(tool.InputSchema)
			fmt.Fprintf(out, "inputRequired: %s\n", strings.Join(required, ","))
			fmt.Fprintf(out, "inputFields: %s\n", strings.Join(fields, ","))
			outputRequired, outputFields := toolSchemaSummary(tool.OutputSchema)
			fmt.Fprintf(out, "outputRequired: %s\n", strings.Join(outputRequired, ","))
			fmt.Fprintf(out, "outputFields: %s\n", strings.Join(outputFields, ","))
			fmt.Fprintln(out, "hint: if invocation is denied, inspect MCP tool grants, AI access policies, skill bindings, resource scopes, and domain permission keys.")
			return
		}
	}
	fmt.Fprintf(out, "tool: %s not visible\n", toolName)
	fmt.Fprintln(out, "hint: check ai.gateway.invoke, domain permission keys, MCP tool grants, AI access policies, skill bindings, AI client context, and resource scopes.")
}

func toolSchemaSummary(schema map[string]any) ([]string, []string) {
	if len(schema) == 0 {
		return nil, nil
	}
	required := stringSliceFromAny(schema["required"])
	properties, _ := schema["properties"].(map[string]any)
	fields := make([]string, 0, len(properties))
	for key := range properties {
		key = strings.TrimSpace(key)
		if key != "" {
			fields = append(fields, key)
		}
	}
	sort.Strings(required)
	sort.Strings(fields)
	return required, fields
}

func stringSliceFromAny(value any) []string {
	switch typed := value.(type) {
	case []string:
		out := append([]string(nil), typed...)
		sort.Strings(out)
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(fmt.Sprint(item))
			if text != "" {
				out = append(out, text)
			}
		}
		sort.Strings(out)
		return out
	default:
		return nil
	}
}

func diagnoseResource(out io.Writer, manifest Manifest, resourceName string) {
	for _, resource := range manifest.Resources {
		if resource.Name == resourceName {
			fmt.Fprintf(out, "resource: %s\n", resource.Name)
			fmt.Fprintf(out, "description: %s\n", resource.Description)
			fmt.Fprintf(out, "requiredPermissionKeys: %s\n", strings.Join(resource.PermissionKeys, ","))
			fmt.Fprintf(out, "requiredScopes: %s\n", strings.Join(resource.RequiredScopes, ","))
			contextRequired, contextFields := toolSchemaSummary(resource.ContextSchema)
			fmt.Fprintf(out, "contextRequired: %s\n", strings.Join(contextRequired, ","))
			fmt.Fprintf(out, "contextFields: %s\n", strings.Join(contextFields, ","))
			fmt.Fprintln(out, "hint: resource reads proxy to Gateway resources/read; check ai.gateway.invoke, resource permission keys, skill bindings, AI client context, and context scope fields.")
			return
		}
	}
	fmt.Fprintf(out, "resource: %s not visible\n", resourceName)
	fmt.Fprintln(out, "hint: check ai.gateway.invoke, resource permission keys, skill bindings, AI client context, and the manifest resource URI.")
}

func diagnosePrompt(out io.Writer, manifest Manifest, promptName string) {
	for _, prompt := range manifest.Prompts {
		if prompt.Name == promptName {
			fmt.Fprintf(out, "prompt: %s\n", prompt.Name)
			fmt.Fprintf(out, "description: %s\n", prompt.Description)
			fmt.Fprintf(out, "requiredPermissionKeys: %s\n", strings.Join(prompt.PermissionKeys, ","))
			fmt.Fprintf(out, "requiredScopes: %s\n", strings.Join(prompt.RequiredScopes, ","))
			argumentRequired, argumentFields := toolSchemaSummary(prompt.ArgumentSchema)
			contextRequired, contextFields := toolSchemaSummary(prompt.ContextSchema)
			fmt.Fprintf(out, "argumentRequired: %s\n", strings.Join(argumentRequired, ","))
			fmt.Fprintf(out, "argumentFields: %s\n", strings.Join(argumentFields, ","))
			fmt.Fprintf(out, "contextRequired: %s\n", strings.Join(contextRequired, ","))
			fmt.Fprintf(out, "contextFields: %s\n", strings.Join(contextFields, ","))
			fmt.Fprintln(out, "hint: prompt reads proxy to Gateway prompts/get; check ai.gateway.invoke, prompt permission keys, skill bindings, AI client context, and prompt arguments/context.")
			return
		}
	}
	fmt.Fprintf(out, "prompt: %s not visible\n", promptName)
	fmt.Fprintln(out, "hint: check ai.gateway.invoke, prompt permission keys, skill bindings, AI client context, and the manifest prompt name.")
}

func diagnoseClusterCapability(out io.Writer, items []ClusterCapabilityMatrixEntry, key string) {
	sortClusterCapabilities(items)
	for _, item := range items {
		if item.Key != key {
			continue
		}
		fmt.Fprintf(out, "clusterCapability: %s\n", item.Key)
		fmt.Fprintf(out, "label: %s\ncategory: %s\nriskLevel: %s\nrequiresApproval: %t\n", item.Label, item.Category, item.RiskLevel, item.RequiresApproval)
		fmt.Fprintf(out, "requiredScopes: %s\ndocsUrl: %s\n", strings.Join(item.RequiredScopes, ","), item.DocsURL)
		fmt.Fprintf(out, "directStatus: %s\ndirectReason: %s\n", item.Direct.Status, redactSensitiveText(item.Direct.Reason))
		fmt.Fprintf(out, "agentStatus: %s\nagentReason: %s\n", item.Agent.Status, redactSensitiveText(item.Agent.Reason))
		if item.Agent.Status == "unsupported" || item.Agent.Status == "partial" {
			fmt.Fprintln(out, "hint: agent mode is not fully available; use the reason above to decide whether to switch to direct mode, request scope, or wait for Agent parity.")
		}
		return
	}
	fmt.Fprintf(out, "clusterCapability: %s not found\n", key)
	fmt.Fprintln(out, "hint: run soha capabilities --domain platform --output names to list known cluster capability keys.")
}

func sortClusterCapabilities(items []ClusterCapabilityMatrixEntry) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
}

func clusterCapabilitySupportText(support ClusterCapabilityModeSupport) string {
	status := strings.TrimSpace(string(support.Status))
	if status == "" {
		status = "unknown"
	}
	if reason := strings.TrimSpace(support.Reason); reason != "" {
		return status + ":" + redactSensitiveText(reason)
	}
	return status
}
