package sohacli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func runPlugin(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("plugin requires a subcommand: search, show, install, list, enable, disable, upgrade, config, remove")
	}
	switch args[0] {
	case "search":
		return runPluginSearch(ctx, args[1:], rt)
	case "show":
		return runPluginShow(ctx, args[1:], rt)
	case "install":
		return runPluginInstall(ctx, args[1:], rt)
	case "list":
		return runPluginList(ctx, args[1:], rt)
	case "enable":
		return runPluginEnable(ctx, args[1:], rt)
	case "disable":
		return runPluginDisable(ctx, args[1:], rt)
	case "upgrade":
		return runPluginUpgrade(ctx, args[1:], rt)
	case "config":
		return runPluginConfig(ctx, args[1:], rt)
	case "remove", "rm":
		return runPluginRemove(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown plugin command %q", args[0])
	}
}

func runPluginSearch(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin search", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	queryFlag := fs.String("query", "", "search query")
	typeFlag := fs.String("type", "", "plugin type")
	publisherFlag := fs.String("publisher", "", "publisher id")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	query := url.Values{}
	setQuery(query, "q", firstNonEmptyString(*queryFlag, strings.Join(fs.Args(), " ")))
	setQuery(query, "type", *typeFlag)
	setQuery(query, "publisher", *publisherFlag)
	items, err := gatewayClient(rt, profile).ListMarketplacePlugins(ctx, query, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, items)
	}
	for _, item := range items {
		fmt.Fprintf(rt.Out, "%s\t%s\t%s\t%s\tinstalled=%t\t%s\n", item.ID, item.Type, item.Publisher, item.Version, item.Installed, item.Name)
	}
	return nil
}

func runPluginShow(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin show", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	installed := fs.Bool("installed", false, "show installed plugin record instead of marketplace detail")
	manifestOnly := fs.Bool("manifest", false, "show installed manifest snapshot")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pluginID := strings.TrimSpace(fs.Arg(0))
	if pluginID == "" {
		return fmt.Errorf("plugin show requires a plugin id")
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	client := gatewayClient(rt, profile)
	headers := gatewayHeaders(profile, "", "", "", "soha-cli")
	if *manifestOnly {
		item, err := client.GetInstalledPluginManifest(ctx, pluginID, headers)
		if err != nil {
			return err
		}
		return writePrettyJSON(rt.Out, item)
	}
	if *installed {
		item, err := client.GetInstalledPlugin(ctx, pluginID, headers)
		if err != nil {
			return err
		}
		if *jsonOutput {
			return writePrettyJSON(rt.Out, item)
		}
		printInstalledPlugin(rt.Out, item)
		return nil
	}
	item, err := client.GetMarketplacePlugin(ctx, pluginID, headers)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, item)
	}
	fmt.Fprintf(rt.Out, "id: %s\nname: %s\npublisher: %s\nversion: %s\ntype: %s\nsource: %s\ninstalled: %t\nsummary: %s\n", item.ID, item.Name, item.Publisher, item.Version, item.Type, item.Source, item.Installed, item.Summary)
	return nil
}

func runPluginInstall(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin install", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	manifestPath := fs.String("manifest", "", "plugin manifest JSON file")
	source := fs.String("source", "", "plugin source URL, path, or marketplace id")
	expectedChecksum := fs.String("expected-checksum", "", "expected sha256:<hex> manifest checksum")
	enable := fs.Bool("enable", false, "enable after install")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	req, err := pluginInstallRequest(firstNonEmptyString(fs.Arg(0), *source), *source, *manifestPath, *expectedChecksum, *enable)
	if err != nil {
		return err
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	item, err := gatewayClient(rt, profile).InstallPlugin(ctx, req, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, item)
	}
	fmt.Fprintf(rt.Out, "Installed plugin %s %s (%s)\n", item.ID, item.Version, item.Status)
	return nil
}

func runPluginList(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin list", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	items, err := gatewayClient(rt, profile).ListInstalledPlugins(ctx, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, items)
	}
	for _, item := range items {
		fmt.Fprintf(rt.Out, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Status, item.Type, item.Version, item.Name)
	}
	return nil
}

func runPluginEnable(ctx context.Context, args []string, rt Runtime) error {
	return runPluginAction(ctx, args, rt, "enable")
}

func runPluginDisable(ctx context.Context, args []string, rt Runtime) error {
	return runPluginAction(ctx, args, rt, "disable")
}

func runPluginAction(ctx context.Context, args []string, rt Runtime, action string) error {
	fs := newRuntimeFlagSet("plugin "+action, args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pluginID := strings.TrimSpace(fs.Arg(0))
	if pluginID == "" {
		return fmt.Errorf("plugin %s requires a plugin id", action)
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	client := gatewayClient(rt, profile)
	headers := gatewayHeaders(profile, "", "", "", "soha-cli")
	var item InstalledPlugin
	if action == "enable" {
		item, err = client.EnablePlugin(ctx, pluginID, headers)
	} else {
		item, err = client.DisablePlugin(ctx, pluginID, headers)
	}
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, item)
	}
	fmt.Fprintf(rt.Out, "%sd plugin %s (%s)\n", strings.ToUpper(action[:1])+action[1:], item.ID, item.Status)
	return nil
}

func runPluginUpgrade(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin upgrade", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	manifestPath := fs.String("manifest", "", "plugin manifest JSON file")
	source := fs.String("source", "", "plugin source URL, path, or marketplace id")
	expectedChecksum := fs.String("expected-checksum", "", "expected sha256:<hex> manifest checksum")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pluginID := strings.TrimSpace(fs.Arg(0))
	if pluginID == "" {
		return fmt.Errorf("plugin upgrade requires a plugin id")
	}
	req, err := pluginInstallRequest(pluginID, *source, *manifestPath, *expectedChecksum, false)
	if err != nil {
		return err
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	item, err := gatewayClient(rt, profile).UpgradePlugin(ctx, pluginID, req, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, item)
	}
	fmt.Fprintf(rt.Out, "Upgraded plugin %s to %s\n", item.ID, item.Version)
	return nil
}

func runPluginConfig(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin config", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	enabled := fs.Bool("enable", false, "enable plugin in the same request")
	metadataJSON := fs.String("metadata-json", "", "metadata JSON object")
	secretRefFlags := repeatableFlag{}
	fs.Var(&secretRefFlags, "secret-ref", "secret ref as name=secret://ref, repeatable")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pluginID := strings.TrimSpace(fs.Arg(0))
	if pluginID == "" {
		return fmt.Errorf("plugin config requires a plugin id")
	}
	var enabledValue *bool
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == "enable" {
			enabledValue = enabled
		}
	})
	metadata := map[string]any{}
	if strings.TrimSpace(*metadataJSON) != "" {
		parsed, err := parseJSONObject([]byte(*metadataJSON))
		if err != nil {
			return err
		}
		metadata = parsed
	}
	req := PluginConfigRequest{
		Enabled:    enabledValue,
		SecretRefs: parseKeyValueFlags(secretRefFlags),
		Metadata:   metadata,
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	item, err := gatewayClient(rt, profile).ConfigurePlugin(ctx, pluginID, req, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writePrettyJSON(rt.Out, item)
	}
	fmt.Fprintf(rt.Out, "Configured plugin %s\n", item.ID)
	return nil
}

func runPluginRemove(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin remove", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pluginID := strings.TrimSpace(fs.Arg(0))
	if pluginID == "" {
		return fmt.Errorf("plugin remove requires a plugin id")
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	if err := gatewayClient(rt, profile).RemovePlugin(ctx, pluginID, gatewayHeaders(profile, "", "", "", "soha-cli")); err != nil {
		return err
	}
	fmt.Fprintf(rt.Out, "Removed plugin %s\n", pluginID)
	return nil
}

func pluginInstallRequest(pluginID, source, manifestPath, expectedChecksum string, enable bool) (PluginInstallRequest, error) {
	req := PluginInstallRequest{
		PluginID:         strings.TrimSpace(pluginID),
		Source:           strings.TrimSpace(source),
		ExpectedChecksum: strings.TrimSpace(expectedChecksum),
		Enable:           enable,
	}
	if strings.TrimSpace(manifestPath) != "" {
		manifest, err := readPluginManifest(manifestPath)
		if err != nil {
			return PluginInstallRequest{}, err
		}
		req.Manifest = &manifest
		if req.PluginID == "" {
			req.PluginID = manifest.ID
		}
	}
	if req.PluginID == "" && req.Manifest == nil {
		return PluginInstallRequest{}, fmt.Errorf("plugin id or --manifest is required")
	}
	return req, nil
}

func readPluginManifest(path string) (PluginManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return PluginManifest{}, err
	}
	var manifest PluginManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return PluginManifest{}, err
	}
	return manifest, nil
}

func printInstalledPlugin(out interface{ Write([]byte) (int, error) }, item InstalledPlugin) {
	fmt.Fprintf(out, "id: %s\nname: %s\npublisher: %s\nversion: %s\ntype: %s\nstatus: %s\nsource: %s\nchecksum: %s\nsignature: %s\n", item.ID, item.Name, item.Publisher, item.Version, item.Type, item.Status, item.Source, item.ChecksumStatus, item.SignatureStatus)
	if item.RequestedPermissions != nil {
		fmt.Fprintf(out, "requestedPermissions: required=%s domain=%s\n", strings.Join(item.RequestedPermissions.Required, ","), strings.Join(item.RequestedPermissions.Domain, ","))
	}
}

type repeatableFlag []string

func (r *repeatableFlag) String() string {
	return strings.Join(*r, ",")
}

func (r *repeatableFlag) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func parseKeyValueFlags(values []string) map[string]string {
	out := map[string]string{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key != "" && val != "" {
			out[key] = val
		}
	}
	return out
}
