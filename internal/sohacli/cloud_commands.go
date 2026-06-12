package sohacli

import (
	"context"
	"fmt"
	"strings"
)

func runCloud(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("cloud requires a subcommand: fleet")
	}
	switch args[0] {
	case "fleet":
		return runCloudFleet(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown cloud command %q", args[0])
	}
}

func runCloudFleet(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("cloud fleet requires a subcommand: diagnostics")
	}
	switch args[0] {
	case "diagnostics":
		return runCloudFleetDiagnostics(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown cloud fleet command %q", args[0])
	}
}

func runCloudFleetDiagnostics(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("cloud fleet diagnostics", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	tenantID := fs.String("tenant-id", "", "Cloud tenant id")
	fleetID := fs.String("fleet-id", "", "managed agent fleet id")
	format := fs.String("output", "summary", "output format: summary, json, or yaml")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *jsonOutput {
		*format = "json"
	}
	formatValue, err := normalizeOutputFormat(*format, "summary", "json", "yaml")
	if err != nil {
		return err
	}
	if strings.TrimSpace(*tenantID) == "" {
		return fmt.Errorf("--tenant-id is required")
	}
	if strings.TrimSpace(*fleetID) == "" {
		return fmt.Errorf("--fleet-id is required")
	}

	_, profileName, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	diagnostics, err := gatewayClient(rt, profile).CloudFleetDiagnostics(ctx, *tenantID, *fleetID)
	if err != nil {
		return err
	}

	switch formatValue {
	case "json":
		return writePrettyJSON(rt.Out, diagnostics)
	case "yaml":
		return writeYAML(rt.Out, diagnostics)
	case "summary":
		writeCloudFleetDiagnosticsSummary(rt, profileName, diagnostics)
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", formatValue)
	}
}

func writeCloudFleetDiagnosticsSummary(rt Runtime, profileName string, diagnostics CloudFleetCapabilityDiagnostics) {
	fmt.Fprintf(rt.Out, "profile: %s\n", profileName)
	fmt.Fprintf(rt.Out, "tenant: %s\nfleet: %s\nmode: %s\nstatus: %s\n", diagnostics.TenantID, diagnostics.FleetID, diagnostics.Mode, diagnostics.Status)
	fmt.Fprintf(rt.Out, "clusters: total=%d available=%d degraded=%d unknown=%d\n",
		diagnostics.ClusterStatusCounts.Total,
		diagnostics.ClusterStatusCounts.Available,
		diagnostics.ClusterStatusCounts.Degraded,
		diagnostics.ClusterStatusCounts.Unknown,
	)
	fmt.Fprintf(rt.Out, "capabilities: available=%d partial=%d unsupported=%d\n",
		diagnostics.CapabilityStatusCounts.Available,
		diagnostics.CapabilityStatusCounts.Partial,
		diagnostics.CapabilityStatusCounts.Unsupported,
	)
	if strings.TrimSpace(diagnostics.Message) != "" {
		fmt.Fprintf(rt.Out, "message: %s\n", redactSensitiveText(diagnostics.Message))
	}
	if len(diagnostics.CapabilityGaps) > 0 {
		fmt.Fprintln(rt.Out, "gaps:")
		for _, gap := range diagnostics.CapabilityGaps {
			fmt.Fprintf(rt.Out, "gap\t%s\tmissing=%s\tpartial=%s\tunsupported=%s\n",
				gap.Key,
				strings.Join(gap.MissingClusterIDs, ","),
				strings.Join(gap.PartialClusterIDs, ","),
				strings.Join(gap.UnsupportedClusterIDs, ","),
			)
		}
	}
	if len(diagnostics.Clusters) > 0 {
		fmt.Fprintln(rt.Out, "clusterDiagnostics:")
		for _, cluster := range diagnostics.Clusters {
			degradedKeys := make([]string, 0, len(cluster.DegradedCapabilities))
			for _, capability := range cluster.DegradedCapabilities {
				degradedKeys = append(degradedKeys, capability.Key+":"+capability.Status)
			}
			fmt.Fprintf(rt.Out, "cluster\t%s\t%s\tavailable=%d\tpartial=%d\tunsupported=%d\tmissing=%s\tdegraded=%s\n",
				cluster.ClusterID,
				cluster.Status,
				cluster.Counts.Available,
				cluster.Counts.Partial,
				cluster.Counts.Unsupported,
				strings.Join(cluster.MissingCapabilities, ","),
				strings.Join(degradedKeys, ","),
			)
		}
	}
}
