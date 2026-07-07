package sohacli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

func runPluginSearchWithOutput(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin search", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	queryFlag := fs.String("query", "", "search query")
	typeFlag := fs.String("type", "", "plugin type")
	publisherFlag := fs.String("publisher", "", "publisher id")
	marketplace := fs.String("marketplace", "", "marketplace catalog URL")
	sourceID := fs.String("source-id", "", "marketplace source id")
	version := fs.String("version", "", "plugin version")
	format := fs.String("output", "table", "output format: table, json, or yaml")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *jsonOutput {
		*format = "json"
	}
	formatValue, err := normalizeOutputFormat(*format, "table", "json", "yaml")
	if err != nil {
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
	applyMarketplaceQuery(query, profile, *marketplace, *sourceID, *version)
	items, err := gatewayClient(rt, profile).ListMarketplacePlugins(ctx, query, gatewayHeaders(profile, "", "", "", "soha-cli"))
	if err != nil {
		return err
	}
	if formatValue != "table" {
		return writeStructuredOutput(rt.Out, formatValue, sanitizeCLIValue(items))
	}
	printMarketplacePluginSearchTable(rt.Out, items)
	return nil
}

func runPluginShowWithOutput(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin show", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	installed := fs.Bool("installed", false, "show installed plugin record instead of marketplace detail")
	manifestOnly := fs.Bool("manifest", false, "show installed manifest snapshot")
	marketplace := fs.String("marketplace", "", "marketplace catalog URL")
	sourceID := fs.String("source-id", "", "marketplace source id")
	version := fs.String("version", "", "plugin version")
	format := fs.String("output", "table", "output format: table, json, or yaml")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *jsonOutput {
		*format = "json"
	} else if *manifestOnly && !pluginOutputFlagSet(fs) {
		*format = "json"
	}
	formatValue, err := normalizeOutputFormat(*format, "table", "json", "yaml")
	if err != nil {
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
		if formatValue != "table" {
			return writeStructuredOutput(rt.Out, formatValue, sanitizeCLIValue(item))
		}
		printPluginManifestTable(rt.Out, item)
		return nil
	}
	if *installed {
		item, err := client.GetInstalledPlugin(ctx, pluginID, headers)
		if err != nil {
			return err
		}
		if formatValue != "table" {
			return writeStructuredOutput(rt.Out, formatValue, sanitizeCLIValue(item))
		}
		printInstalledPluginTable(rt.Out, item)
		return nil
	}
	query := url.Values{}
	applyMarketplaceQuery(query, profile, *marketplace, *sourceID, *version)
	item, err := client.GetMarketplacePlugin(ctx, pluginID, query, headers)
	if err != nil {
		return err
	}
	if formatValue != "table" {
		return writeStructuredOutput(rt.Out, formatValue, sanitizeCLIValue(item))
	}
	printMarketplacePluginTable(rt.Out, item)
	return nil
}

func runPluginListWithOutput(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("plugin list", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	format := fs.String("output", "table", "output format: table, json, or yaml")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *jsonOutput {
		*format = "json"
	}
	formatValue, err := normalizeOutputFormat(*format, "table", "json", "yaml")
	if err != nil {
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
	if formatValue != "table" {
		return writeStructuredOutput(rt.Out, formatValue, sanitizeCLIValue(items))
	}
	printInstalledPluginListTable(rt.Out, items)
	return nil
}

func pluginOutputFlagSet(fs *flag.FlagSet) bool {
	set := false
	fs.Visit(func(item *flag.Flag) {
		if item.Name == "output" {
			set = true
		}
	})
	return set
}

func applyMarketplaceQuery(query url.Values, profile ProfileConfig, marketplaceURL, sourceID, version string) {
	setQuery(query, "marketplaceUrl", normalizeServerURL(firstNonEmptyString(marketplaceURL, profile.MarketplaceURL)))
	setQuery(query, "sourceId", firstNonEmptyString(sourceID, profile.MarketplaceSourceID))
	setQuery(query, "version", version)
}

func printMarketplacePluginSearchTable(out io.Writer, items []MarketplacePlugin) {
	for _, item := range items {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\tinstalled=%t\t%s\n",
			redactSensitiveText(item.ID),
			redactSensitiveText(item.Type),
			redactSensitiveText(item.Publisher),
			redactSensitiveText(item.Version),
			item.Installed,
			redactSensitiveText(item.Name),
		)
	}
}

func printMarketplacePluginTable(out io.Writer, item MarketplacePlugin) {
	fmt.Fprintf(out, "id: %s\nname: %s\npublisher: %s\nversion: %s\ntype: %s\nsource: %s\ninstalled: %t\nsummary: %s\n",
		redactSensitiveText(item.ID),
		redactSensitiveText(item.Name),
		redactSensitiveText(item.Publisher),
		redactSensitiveText(item.Version),
		redactSensitiveText(item.Type),
		redactSensitiveText(item.Source),
		item.Installed,
		redactSensitiveText(item.Summary),
	)
}

func printInstalledPluginTable(out io.Writer, item InstalledPlugin) {
	fmt.Fprintf(out, "id: %s\nname: %s\npublisher: %s\nversion: %s\ntype: %s\nstatus: %s\nsource: %s\nchecksum: %s\nsignature: %s\n",
		redactSensitiveText(item.ID),
		redactSensitiveText(item.Name),
		redactSensitiveText(item.Publisher),
		redactSensitiveText(item.Version),
		redactSensitiveText(item.Type),
		redactSensitiveText(item.Status),
		redactSensitiveText(item.Source),
		redactSensitiveText(item.ChecksumStatus),
		redactSensitiveText(item.SignatureStatus),
	)
	if item.RequestedPermissions != nil {
		fmt.Fprintf(out, "requestedPermissions: required=%s domain=%s\n",
			strings.Join(redactSensitiveStrings(item.RequestedPermissions.Required), ","),
			strings.Join(redactSensitiveStrings(item.RequestedPermissions.Domain), ","),
		)
	}
}

func printInstalledPluginListTable(out io.Writer, items []InstalledPlugin) {
	for _, item := range items {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\n",
			redactSensitiveText(item.ID),
			redactSensitiveText(item.Status),
			redactSensitiveText(item.Type),
			redactSensitiveText(item.Version),
			redactSensitiveText(item.Name),
		)
	}
}

func printPluginManifestTable(out io.Writer, item PluginManifest) {
	fmt.Fprintf(out, "id: %s\nname: %s\npublisher: %s\nversion: %s\ntype: %s\ndescription: %s\nhomepage: %s\n",
		redactSensitiveText(item.ID),
		redactSensitiveText(item.Name),
		redactSensitiveText(item.Publisher),
		redactSensitiveText(item.Version),
		redactSensitiveText(item.Type),
		redactSensitiveText(item.Description),
		redactSensitiveText(item.Homepage),
	)
	if item.Permissions != nil {
		fmt.Fprintf(out, "permissions: required=%s domain=%s\n",
			strings.Join(redactSensitiveStrings(item.Permissions.Required), ","),
			strings.Join(redactSensitiveStrings(item.Permissions.Domain), ","),
		)
	}
	if item.Secrets != nil {
		names := make([]string, 0, len(item.Secrets.Required))
		for _, requirement := range item.Secrets.Required {
			names = append(names, redactSensitiveText(requirement.Name))
		}
		sort.Strings(names)
		fmt.Fprintf(out, "secrets: required=%s\n", strings.Join(names, ","))
	}
}

func redactSensitiveStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, redactSensitiveText(value))
	}
	return out
}
