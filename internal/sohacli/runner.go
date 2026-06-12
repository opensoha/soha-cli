package sohacli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Runtime struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	ConfigPath  string
	HTTPClient  *http.Client
	HTTPTimeout time.Duration
}

func Run(ctx context.Context, args []string, rt Runtime) int {
	if rt.In == nil {
		rt.In = os.Stdin
	}
	if rt.Out == nil {
		rt.Out = os.Stdout
	}
	if rt.Err == nil {
		rt.Err = os.Stderr
	}
	if rt.ConfigPath == "" {
		rt.ConfigPath = defaultConfigPath()
	}
	var err error
	args, rt.HTTPTimeout, err = resolveRuntimeTimeout(args, rt.HTTPTimeout)
	if err != nil {
		fmt.Fprintln(rt.Err, "error:", err)
		return 1
	}
	if len(args) == 0 {
		printUsage(rt.Err)
		return 2
	}
	if len(args) > 1 && isHelpArg(args[1]) && printCommandHelp(args[0], rt.Out) {
		return 0
	}
	cmd := args[0]
	switch cmd {
	case "version":
		err = runVersion(args[1:], rt)
	case "login":
		err = runLogin(ctx, args[1:], rt)
	case "capabilities":
		err = runCapabilities(ctx, args[1:], rt)
	case "tool":
		err = runTool(ctx, args[1:], rt)
	case "resource":
		err = runResource(ctx, args[1:], rt)
	case "prompt":
		err = runPrompt(ctx, args[1:], rt)
	case "token":
		err = runToken(ctx, args[1:], rt)
	case "service-account":
		err = runServiceAccount(ctx, args[1:], rt)
	case "audit":
		err = runAudit(ctx, args[1:], rt)
	case "approval":
		err = runApproval(ctx, args[1:], rt)
	case "governance":
		err = runGovernance(ctx, args[1:], rt)
	case "cloud":
		err = runCloud(ctx, args[1:], rt)
	case "profile":
		err = runProfile(args[1:], rt)
	case "context":
		err = runContext(ctx, args[1:], rt)
	case "mcp":
		err = runMCP(ctx, args[1:], rt)
	case "skill":
		err = runSkill(args[1:], rt)
	case "add":
		err = runAdd(args[1:], rt)
	case "plugin":
		err = runPluginWithOutput(ctx, args[1:], rt)
	case "diagnose":
		err = runDiagnose(ctx, args[1:], rt)
	case "completion":
		err = runCompletion(args[1:], rt)
	case "docs":
		err = runDocs(args[1:], rt)
	case "help", "-h", "--help":
		printUsage(rt.Out)
		return 0
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(rt.Err, "error:", err)
		return 1
	}
	return 0
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: soha <command> [options]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Global options:")
	fmt.Fprintln(out, "  --timeout <duration>  HTTP request timeout, e.g. 10s or 1m (default 30s)")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  version        Print build version information")
	fmt.Fprintln(out, "  login          Authenticate and store a local profile")
	fmt.Fprintln(out, "  capabilities   Print AI Gateway or platform capability metadata")
	fmt.Fprintln(out, "  tool call      Invoke an AI Gateway tool with JSON input")
	fmt.Fprintln(out, "  resource read  Read an AI Gateway MCP resource")
	fmt.Fprintln(out, "  prompt get     Get an AI Gateway MCP prompt")
	fmt.Fprintln(out, "  token          Manage personal access tokens")
	fmt.Fprintln(out, "  service-account Manage AI Gateway service accounts and tokens")
	fmt.Fprintln(out, "  audit list     Query AI Gateway audit logs")
	fmt.Fprintln(out, "  approval       List, trace, or decide AI Gateway approval requests")
	fmt.Fprintln(out, "  governance status Show AI Gateway governance health and metrics")
	fmt.Fprintln(out, "  cloud fleet diagnostics Show Cloud managed agent fleet capability diagnostics")
	fmt.Fprintln(out, "  profile        List, show, or switch profiles")
	fmt.Fprintln(out, "  context        Show or update AI client context headers")
	fmt.Fprintln(out, "  mcp start      Run the soha MCP stdio server")
	fmt.Fprintln(out, "  mcp install    Print MCP client configuration")
	fmt.Fprintln(out, "  skill list     List local soha AI Gateway skill files")
	fmt.Fprintln(out, "  skill install  Install local soha AI Gateway skill files")
	fmt.Fprintln(out, "  add            Add Soha MCP and skills to an AI agent or IDE")
	fmt.Fprintln(out, "  plugin         Search, install, and manage Soha plugins")
	fmt.Fprintln(out, "  diagnose       Check profile and Gateway connectivity")
	fmt.Fprintln(out, "  completion     Print shell completion script")
	fmt.Fprintln(out, "  docs           Generate CLI command reference documentation")
}

func printCommandHelp(command string, out io.Writer) bool {
	switch command {
	case "version":
		fmt.Fprintln(out, "Usage: soha version [--json]")
	case "capabilities":
		fmt.Fprintln(out, "Usage: soha capabilities [--domain gateway|platform] [--output json|yaml|names|inputs]")
	case "tool":
		fmt.Fprintln(out, "Usage: soha tool call <name> [options]")
	case "resource":
		fmt.Fprintln(out, "Usage: soha resource read <uri> [options]")
	case "prompt":
		fmt.Fprintln(out, "Usage: soha prompt get <name> [options]")
	case "token":
		fmt.Fprintln(out, "Usage: soha token <list|create|revoke> [options]")
	case "service-account":
		fmt.Fprintln(out, "Usage: soha service-account <list|create|token-list|token-create|token-revoke> [options]")
	case "audit":
		fmt.Fprintln(out, "Usage: soha audit list [options]")
	case "approval":
		fmt.Fprintln(out, "Usage: soha approval <list|timeline|approve|reject|cancel> [options]")
	case "governance":
		fmt.Fprintln(out, "Usage: soha governance status [options]")
	case "cloud":
		fmt.Fprintln(out, "Usage: soha cloud fleet diagnostics [options]")
	case "profile":
		fmt.Fprintln(out, "Usage: soha profile <list|show|use> [options]")
	case "context":
		fmt.Fprintln(out, "Usage: soha context <show|set> [options]")
	case "mcp":
		fmt.Fprintln(out, "Usage: soha mcp <start|install> [options]")
	case "skill":
		fmt.Fprintln(out, "Usage: soha skill <list|install> [options]")
	case "add":
		fmt.Fprintln(out, "Usage: soha add [codex|claude|cursor|kiro|gemini|antigravity|antigravity-ide|trae|all] [options]")
	case "plugin":
		fmt.Fprintln(out, "Usage: soha plugin <search|show|install|list|enable|disable|upgrade|config|remove> [options]")
	case "diagnose":
		fmt.Fprintln(out, "Usage: soha diagnose [options]")
	case "completion":
		fmt.Fprintln(out, "Usage: soha completion [bash|zsh]")
	case "docs":
		fmt.Fprintln(out, "Usage: soha docs [--format markdown]")
	default:
		return false
	}
	return true
}

func newRuntimeFlagSet(name string, args []string, rt Runtime) *flag.FlagSet {
	return newFlagSet(name, flagOutput(args, rt))
}

func flagOutput(args []string, rt Runtime) io.Writer {
	if hasHelpArg(args) {
		return rt.Out
	}
	return rt.Err
}

func hasHelpArg(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func isHelpArg(arg string) bool {
	arg = strings.TrimSpace(arg)
	return arg == "-h" || arg == "--help"
}

func runVersion(args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("version", args, rt)
	jsonOutput := fs.Bool("json", false, "print JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return writeVersion(rt.Out, *jsonOutput)
}

func runLogin(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("login", args, rt)
	server := fs.String("server", defaultServerFromEnv(), "soha server URL")
	login := fs.String("login", env("SOHA_LOGIN"), "login name")
	password := fs.String("password", env("SOHA_PASSWORD"), "login password")
	profile := fs.String("profile", defaultProfile, "profile name")
	aiClientID := fs.String("ai-client-id", env("SOHA_AI_CLIENT_ID"), "AI client id")
	aiClientName := fs.String("ai-client", env("SOHA_AI_CLIENT"), "AI client display name")
	source := fs.String("source", "soha", "request source label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	serverURL := normalizeServerURL(*server)
	if serverURL == "" {
		return fmt.Errorf("--server is required")
	}
	if strings.TrimSpace(*login) == "" {
		return fmt.Errorf("--login is required")
	}
	if strings.TrimSpace(*password) == "" {
		fmt.Fprint(rt.Err, "Password: ")
		value, err := readPassword(rt)
		if err != nil {
			return err
		}
		*password = value
	}
	if strings.TrimSpace(*password) == "" {
		return fmt.Errorf("password is required")
	}
	client := APIClient{ServerURL: serverURL, Client: rt.HTTPClient, Timeout: rt.HTTPTimeout}
	result, err := client.Login(ctx, *login, *password)
	if err != nil {
		return err
	}
	if strings.TrimSpace(result.Data.Tokens.AccessToken) == "" {
		return fmt.Errorf("login response did not include an access token")
	}
	cfg, err := loadConfig(rt.ConfigPath)
	if err != nil {
		return err
	}
	name := profileName(*profile)
	cfg.CurrentProfile = name
	profileConfig, err := profileFromAuthResult(ProfileConfig{
		ServerURL:    serverURL,
		AIClientID:   strings.TrimSpace(*aiClientID),
		AIClientName: strings.TrimSpace(*aiClientName),
		Source:       strings.TrimSpace(*source),
	}, result, time.Now())
	if err != nil {
		return err
	}
	cfg.Profiles[name] = profileConfig
	if err := saveConfig(rt.ConfigPath, cfg); err != nil {
		return err
	}
	fmt.Fprintf(rt.Out, "Logged in to %s as %s (profile %s)\n", cfg.Profiles[name].ServerURL, result.Data.User.UserName, name)
	return nil
}

func runCapabilities(ctx context.Context, args []string, rt Runtime) error {
	fs := newRuntimeFlagSet("capabilities", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	domain := fs.String("domain", "gateway", "capability domain: gateway or platform")
	format := fs.String("output", "json", "output format: json, yaml, names, or inputs")
	jsonOutput := fs.Bool("json", false, "print JSON output")
	aiClientID := fs.String("ai-client-id", "", "override AI client id")
	aiClientName := fs.String("ai-client", "", "override AI client display name")
	skillID := fs.String("skill-id", "", "override skill id")
	source := fs.String("source", "", "override source label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	domainValue, err := normalizeCapabilityDomain(*domain)
	if err != nil {
		return err
	}
	if *jsonOutput {
		*format = "json"
	}
	formatValue, err := normalizeOutputFormat(*format, "json", "yaml", "names", "inputs")
	if err != nil {
		return err
	}
	_, name, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	if domainValue == "platform" {
		return runPlatformCapabilities(ctx, rt, name, profile, formatValue)
	}
	manifest, err := gatewayClient(rt, profile).Capabilities(ctx, gatewayHeaders(profile, *aiClientID, *aiClientName, *skillID, *source))
	if err != nil {
		return err
	}
	switch formatValue {
	case "", "json":
		return writePrettyJSON(rt.Out, manifest)
	case "yaml":
		return writeYAML(rt.Out, manifest)
	case "names":
		fmt.Fprintf(rt.Out, "profile: %s\n", name)
		for _, tool := range manifest.Tools {
			fmt.Fprintf(rt.Out, "tool\t%s\t%s\t%s\n", tool.Name, tool.RiskLevel, approvalText(tool.RequiresApproval))
		}
		for _, item := range manifest.Resources {
			fmt.Fprintf(rt.Out, "resource\t%s\n", item.Name)
		}
		for _, item := range manifest.Prompts {
			fmt.Fprintf(rt.Out, "prompt\t%s\n", item.Name)
		}
		for _, item := range manifest.Skills {
			fmt.Fprintf(rt.Out, "skill\t%s\t%s\n", item.ID, item.Name)
		}
		return nil
	case "inputs":
		fmt.Fprintf(rt.Out, "profile: %s\n", name)
		for _, tool := range manifest.Tools {
			required, fields := toolSchemaSummary(tool.InputSchema)
			outputRequired, outputFields := toolSchemaSummary(tool.OutputSchema)
			fmt.Fprintf(rt.Out, "tool\t%s\trequired=%s\tfields=%s\toutputRequired=%s\toutputFields=%s\n", tool.Name, strings.Join(required, ","), strings.Join(fields, ","), strings.Join(outputRequired, ","), strings.Join(outputFields, ","))
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", *format)
	}
}

func runPlatformCapabilities(ctx context.Context, rt Runtime, profileName string, profile ProfileConfig, formatValue string) error {
	items, err := gatewayClient(rt, profile).ClusterCapabilities(ctx)
	if err != nil {
		return err
	}
	sortClusterCapabilities(items)
	switch formatValue {
	case "", "json":
		return writePrettyJSON(rt.Out, items)
	case "yaml":
		return writeYAML(rt.Out, items)
	case "names":
		fmt.Fprintf(rt.Out, "profile: %s\n", profileName)
		for _, item := range items {
			fmt.Fprintf(rt.Out, "capability\t%s\t%s\t%s\tscopes=%s\tdirect=%s\tagent=%s",
				item.Key,
				item.RiskLevel,
				approvalText(item.RequiresApproval),
				strings.Join(item.RequiredScopes, ","),
				clusterCapabilitySupportText(item.Direct),
				clusterCapabilitySupportText(item.Agent),
			)
			if reason := strings.TrimSpace(item.Agent.Reason); reason != "" {
				fmt.Fprintf(rt.Out, "\treason=%s", redactSensitiveText(reason))
			}
			fmt.Fprintln(rt.Out)
		}
		return nil
	case "inputs":
		return fmt.Errorf("platform capabilities do not expose input schemas; use --output names, json, or yaml")
	default:
		return fmt.Errorf("unsupported output format %q", formatValue)
	}
}

func runTool(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("tool requires a subcommand: call")
	}
	switch args[0] {
	case "call":
		return runToolCall(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown tool command %q", args[0])
	}
}

func runToolCall(ctx context.Context, args []string, rt Runtime) error {
	toolName, args := extractLeadingToolName(args)
	fs := newRuntimeFlagSet("tool call", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	inputPath := fs.String("input", "", "JSON input file path, or - for stdin")
	inputJSON := fs.String("input-json", "", "inline JSON tool input")
	aiClientID := fs.String("ai-client-id", "", "override AI client id")
	aiClientName := fs.String("ai-client", "", "override AI client display name")
	skillID := fs.String("skill-id", "", "override skill id")
	source := fs.String("source", "", "override source label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	toolName = firstNonEmptyString(toolName, fs.Arg(0))
	if toolName == "" {
		return fmt.Errorf("tool call requires a tool name")
	}
	if *inputPath != "" && *inputJSON != "" {
		return fmt.Errorf("use either --input or --input-json, not both")
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	input, err := readToolInput(rt, *inputPath, *inputJSON)
	if err != nil {
		return err
	}
	result, err := gatewayClient(rt, profile).InvokeTool(ctx, toolName, input, gatewayHeaders(profile, *aiClientID, *aiClientName, *skillID, *source))
	if err != nil {
		return err
	}
	return writePrettyJSON(rt.Out, sanitizeCLIValue(result))
}

func runResource(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("resource requires a subcommand: read")
	}
	switch args[0] {
	case "read":
		return runResourceRead(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown resource command %q", args[0])
	}
}

func runResourceRead(ctx context.Context, args []string, rt Runtime) error {
	uri, args := extractLeadingValue(args)
	fs := newRuntimeFlagSet("resource read", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	contextPath := fs.String("context", "", "JSON context file path, or - for stdin")
	contextJSON := fs.String("context-json", "", "inline JSON context")
	aiClientID := fs.String("ai-client-id", "", "override AI client id")
	aiClientName := fs.String("ai-client", "", "override AI client display name")
	skillID := fs.String("skill-id", "", "override skill id")
	source := fs.String("source", "", "override source label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	uri = firstNonEmptyString(uri, fs.Arg(0))
	if uri == "" {
		return fmt.Errorf("resource read requires a resource URI")
	}
	if *contextPath != "" && *contextJSON != "" {
		return fmt.Errorf("use either --context or --context-json, not both")
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	contextValues, err := readJSONInput(rt, *contextPath, *contextJSON)
	if err != nil {
		return err
	}
	result, err := gatewayClient(rt, profile).ReadResource(ctx, uri, contextValues, gatewayHeaders(profile, *aiClientID, *aiClientName, *skillID, *source))
	if err != nil {
		return err
	}
	return writePrettyJSON(rt.Out, sanitizeCLIValue(result))
}

func runPrompt(ctx context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt requires a subcommand: get")
	}
	switch args[0] {
	case "get":
		return runPromptGet(ctx, args[1:], rt)
	default:
		return fmt.Errorf("unknown prompt command %q", args[0])
	}
}

func runPromptGet(ctx context.Context, args []string, rt Runtime) error {
	name, args := extractLeadingValue(args)
	fs := newRuntimeFlagSet("prompt get", args, rt)
	profileFlag := fs.String("profile", "", "profile name")
	argumentsPath := fs.String("arguments", "", "JSON arguments file path, or - for stdin")
	argumentsJSON := fs.String("arguments-json", "", "inline JSON arguments")
	contextPath := fs.String("context", "", "JSON context file path")
	contextJSON := fs.String("context-json", "", "inline JSON context")
	aiClientID := fs.String("ai-client-id", "", "override AI client id")
	aiClientName := fs.String("ai-client", "", "override AI client display name")
	skillID := fs.String("skill-id", "", "override skill id")
	source := fs.String("source", "", "override source label")
	if err := fs.Parse(args); err != nil {
		return err
	}
	name = firstNonEmptyString(name, fs.Arg(0))
	if name == "" {
		return fmt.Errorf("prompt get requires a prompt name")
	}
	if *argumentsPath != "" && *argumentsJSON != "" {
		return fmt.Errorf("use either --arguments or --arguments-json, not both")
	}
	if *contextPath != "" && *contextJSON != "" {
		return fmt.Errorf("use either --context or --context-json, not both")
	}
	if strings.TrimSpace(*argumentsPath) == "-" && strings.TrimSpace(*contextPath) == "-" {
		return fmt.Errorf("only one of --arguments or --context can read from stdin")
	}
	_, _, profile, err := loadRuntimeProfile(ctx, rt, *profileFlag)
	if err != nil {
		return err
	}
	arguments, err := readJSONInput(rt, *argumentsPath, *argumentsJSON)
	if err != nil {
		return err
	}
	contextValues, err := readJSONInput(rt, *contextPath, *contextJSON)
	if err != nil {
		return err
	}
	result, err := gatewayClient(rt, profile).GetPrompt(ctx, name, arguments, contextValues, gatewayHeaders(profile, *aiClientID, *aiClientName, *skillID, *source))
	if err != nil {
		return err
	}
	return writePrettyJSON(rt.Out, sanitizeCLIValue(result))
}

func resolveRuntimeTimeout(args []string, runtimeTimeout time.Duration) ([]string, time.Duration, error) {
	stripped, flagValue, err := extractTimeoutFlag(args)
	if err != nil {
		return nil, 0, err
	}
	if flagValue != "" {
		timeout, err := parseHTTPTimeout("--timeout", flagValue)
		if err != nil {
			return nil, 0, err
		}
		return stripped, timeout, nil
	}
	if runtimeTimeout > 0 {
		return stripped, runtimeTimeout, nil
	}
	if value := env("SOHA_HTTP_TIMEOUT"); value != "" {
		timeout, err := parseHTTPTimeout("SOHA_HTTP_TIMEOUT", value)
		if err != nil {
			return nil, 0, err
		}
		return stripped, timeout, nil
	}
	return stripped, 0, nil
}

func extractTimeoutFlag(args []string) ([]string, string, error) {
	out := make([]string, 0, len(args))
	value := ""
	for index := 0; index < len(args); index++ {
		arg := strings.TrimSpace(args[index])
		switch {
		case arg == "--":
			out = append(out, args[index:]...)
			return out, value, nil
		case arg == "--timeout":
			if index+1 >= len(args) {
				return nil, "", fmt.Errorf("--timeout requires a duration")
			}
			index++
			value = strings.TrimSpace(args[index])
		case strings.HasPrefix(arg, "--timeout="):
			value = strings.TrimSpace(strings.TrimPrefix(arg, "--timeout="))
		default:
			out = append(out, args[index])
		}
	}
	return out, value, nil
}

func parseHTTPTimeout(source, value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("%s requires a duration", source)
	}
	timeout, err := time.ParseDuration(value)
	if err != nil {
		seconds, secondsErr := strconv.ParseFloat(value, 64)
		if secondsErr != nil {
			return 0, fmt.Errorf("invalid %s %q; use a Go duration like 30s or a number of seconds", source, value)
		}
		timeout = time.Duration(seconds * float64(time.Second))
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", source)
	}
	return timeout, nil
}

func loadRuntimeProfile(ctx context.Context, rt Runtime, requested string) (Config, string, ProfileConfig, error) {
	cfg, err := loadConfig(rt.ConfigPath)
	if err != nil {
		return Config{}, "", ProfileConfig{}, err
	}
	name, profile, err := resolveProfile(cfg, requested)
	if err != nil {
		return Config{}, "", ProfileConfig{}, err
	}
	token := strings.TrimSpace(env("SOHA_TOKEN"))
	server := normalizeServerURL(env("SOHA_SERVER"))
	if server != "" {
		profile.ServerURL = server
	}
	if token != "" {
		profile.AccessToken = token
	} else if accessTokenExpired(profile, time.Now()) {
		persist := server == ""
		profile, err = refreshRuntimeProfile(ctx, rt, cfg, name, profile, persist)
		if err != nil {
			return Config{}, "", ProfileConfig{}, err
		}
	}
	if strings.TrimSpace(profile.Source) == "" {
		profile.Source = "soha"
	}
	return cfg, name, profile, nil
}

func accessTokenExpired(profile ProfileConfig, now time.Time) bool {
	return !profile.ExpiresAt.IsZero() && !profile.ExpiresAt.After(now)
}

func refreshRuntimeProfile(ctx context.Context, rt Runtime, cfg Config, name string, profile ProfileConfig, persist bool) (ProfileConfig, error) {
	if strings.TrimSpace(profile.RefreshToken) == "" {
		return ProfileConfig{}, fmt.Errorf("profile %q access token expired at %s and has no refresh token; run soha login again", name, profile.ExpiresAt.Format(time.RFC3339))
	}
	result, err := (APIClient{ServerURL: profile.ServerURL, Client: rt.HTTPClient, Timeout: rt.HTTPTimeout}).Refresh(ctx, profile.RefreshToken)
	if err != nil {
		return ProfileConfig{}, fmt.Errorf("refresh profile %q: %w", name, err)
	}
	updated, err := profileFromAuthResult(profile, result, time.Now())
	if err != nil {
		return ProfileConfig{}, fmt.Errorf("refresh profile %q: %w", name, err)
	}
	if persist {
		cfg.Profiles[name] = updated
		if err := saveConfig(rt.ConfigPath, cfg); err != nil {
			return ProfileConfig{}, err
		}
	}
	return updated, nil
}

func profileFromAuthResult(profile ProfileConfig, result loginResponse, now time.Time) (ProfileConfig, error) {
	tokens := result.Data.Tokens
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return ProfileConfig{}, fmt.Errorf("auth response did not include an access token")
	}
	profile.AccessToken = tokens.AccessToken
	profile.RefreshToken = firstNonEmptyString(tokens.RefreshToken, profile.RefreshToken)
	profile.ExpiresAt = tokenExpiresAt(tokens.ExpiresAt, tokens.ExpiresIn, now)
	profile.UserID = firstNonEmptyString(result.Data.User.UserID, profile.UserID)
	profile.UserName = firstNonEmptyString(result.Data.User.UserName, profile.UserName)
	return profile, nil
}

func tokenExpiresAt(expiresAt time.Time, expiresIn int64, now time.Time) time.Time {
	if !expiresAt.IsZero() {
		return expiresAt
	}
	if expiresIn > 0 {
		return now.Add(time.Duration(expiresIn) * time.Second)
	}
	return time.Time{}
}

func gatewayClient(rt Runtime, profile ProfileConfig) APIClient {
	return APIClient{ServerURL: profile.ServerURL, Token: profile.AccessToken, Client: rt.HTTPClient, Timeout: rt.HTTPTimeout}
}

func gatewayHeaders(profile ProfileConfig, aiClientID, aiClientName, skillID, source string) map[string]string {
	return map[string]string{
		"X-Soha-AI-Client-ID": firstNonEmptyString(aiClientID, profile.AIClientID),
		"X-Soha-AI-Client":    firstNonEmptyString(aiClientName, profile.AIClientName),
		"X-Soha-Skill-ID":     firstNonEmptyString(skillID, profile.SkillID),
		"X-Soha-Source":       firstNonEmptyString(source, profile.Source, "soha"),
	}
}

func writePrettyJSON(out io.Writer, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(raw))
	return err
}

func normalizeOutputFormat(format string, allowed ...string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	if normalized == "" && len(allowed) > 0 {
		normalized = allowed[0]
	}
	for _, item := range allowed {
		if normalized == item {
			return normalized, nil
		}
	}
	return "", fmt.Errorf("unsupported output format %q", format)
}

func normalizeCapabilityDomain(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "gateway", "ai-gateway", "aigateway":
		return "gateway", nil
	case "platform", "cluster", "clusters":
		return "platform", nil
	default:
		return "", fmt.Errorf("unsupported capability domain %q", value)
	}
}

func writeStructuredOutput(out io.Writer, format string, value any) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return writePrettyJSON(out, value)
	case "yaml":
		return writeYAML(out, value)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func writeYAML(out io.Writer, value any) error {
	normalized, err := normalizeYAMLValue(value)
	if err != nil {
		return err
	}
	if err := writeYAMLValue(out, normalized, 0); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out)
	return err
}

func normalizeYAMLValue(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func writeYAMLValue(out io.Writer, value any, indent int) error {
	prefix := strings.Repeat(" ", indent)
	switch typed := value.(type) {
	case map[string]any:
		if len(typed) == 0 {
			_, err := fmt.Fprintf(out, "%s{}\n", prefix)
			return err
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item := typed[key]
			if yamlInlineValue(item) {
				if _, err := fmt.Fprintf(out, "%s%s: %s\n", prefix, yamlKey(key), yamlScalar(item)); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(out, "%s%s:\n", prefix, yamlKey(key)); err != nil {
				return err
			}
			if err := writeYAMLValue(out, item, indent+2); err != nil {
				return err
			}
		}
		return nil
	case []any:
		if len(typed) == 0 {
			_, err := fmt.Fprintf(out, "%s[]\n", prefix)
			return err
		}
		for _, item := range typed {
			if yamlInlineValue(item) {
				if _, err := fmt.Fprintf(out, "%s- %s\n", prefix, yamlScalar(item)); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(out, "%s-\n", prefix); err != nil {
				return err
			}
			if err := writeYAMLValue(out, item, indent+2); err != nil {
				return err
			}
		}
		return nil
	default:
		_, err := fmt.Fprintf(out, "%s%s\n", prefix, yamlScalar(typed))
		return err
	}
}

func yamlInlineValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	default:
		return true
	}
}

func yamlKey(value string) string {
	if yamlBareString(value) {
		return value
	}
	return strconv.Quote(value)
}

func yamlScalar(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		if yamlBareString(typed) {
			return typed
		}
		return strconv.Quote(typed)
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case map[string]any:
		if len(typed) == 0 {
			return "{}"
		}
	case []any:
		if len(typed) == 0 {
			return "[]"
		}
	}
	return strconv.Quote(fmt.Sprint(value))
}

func yamlBareString(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "null", "true", "false", "yes", "no", "on", "off", "~":
		return false
	}
	if strings.HasPrefix(value, "-") {
		return false
	}
	for _, char := range value {
		switch char {
		case ' ', '\t', '\n', '\r', ':', '#', '{', '}', '[', ']', '&', '*', '?', '|', '<', '>', '=', '!', '%', '@', '`', '"', '\'':
			return false
		}
	}
	return true
}

func readToolInput(rt Runtime, path, inline string) (map[string]any, error) {
	return readJSONInput(rt, path, inline)
}

func readJSONInput(rt Runtime, path, inline string) (map[string]any, error) {
	switch {
	case strings.TrimSpace(inline) != "":
		return parseJSONObject([]byte(inline))
	case strings.TrimSpace(path) == "":
		return map[string]any{}, nil
	case strings.TrimSpace(path) == "-":
		raw, err := io.ReadAll(io.LimitReader(rt.In, 10<<20))
		if err != nil {
			return nil, err
		}
		return parseJSONObject(raw)
	default:
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return parseJSONObject(raw)
	}
}

func parseJSONObject(raw []byte) (map[string]any, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]any{}, nil
	}
	return out, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func setQuery(values url.Values, key, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		values.Set(key, value)
	}
}

func sanitizeCreatedToken(value any) any {
	return sanitizeCLIValue(value)
}

func sanitizeCLIValue(value any) any {
	switch typed := value.(type) {
	case CreatedPersonalAccessToken:
		return map[string]any{"token": sanitizeCLIValue(typed.Token), "value": typed.Value}
	case CreatedServiceAccountToken:
		return map[string]any{"token": sanitizeCLIValue(typed.Token), "value": typed.Value}
	case ToolInvocationResult:
		return ToolInvocationResult{
			ToolName:         typed.ToolName,
			RiskLevel:        typed.RiskLevel,
			RequiresApproval: typed.RequiresApproval,
			Result:           typed.Result,
			Output:           sanitizeCLIValue(typed.Output),
			RelatedIDs:       sanitizeCLIMap(typed.RelatedIDs),
			Audit:            sanitizeCLIMap(typed.Audit),
		}
	case ResourceReadResult:
		typed.Text = redactSensitiveText(typed.Text)
		typed.Data = sanitizeCLIValue(typed.Data)
		typed.RelatedIDs = sanitizeCLIMap(typed.RelatedIDs)
		typed.Audit = sanitizeCLIMap(typed.Audit)
		return typed
	case PromptGetResult:
		typed.Description = redactSensitiveText(typed.Description)
		for index := range typed.Messages {
			typed.Messages[index].Content = redactSensitiveText(typed.Messages[index].Content)
		}
		typed.RelatedIDs = sanitizeCLIMap(typed.RelatedIDs)
		typed.Audit = sanitizeCLIMap(typed.Audit)
		return typed
	case PersonalAccessToken:
		return typed
	case ServiceAccountToken:
		return typed
	case AuditLog:
		typed.Metadata = sanitizeCLIMap(typed.Metadata)
		typed.ResourceScope = sanitizeCLIMap(typed.ResourceScope)
		typed.Summary = redactSensitiveText(typed.Summary)
		return typed
	case ApprovalRequest:
		typed.ResourceScope = sanitizeCLIMap(typed.ResourceScope)
		typed.ToolInput = sanitizeCLIMap(typed.ToolInput)
		typed.RelatedIDs = sanitizeCLIMap(typed.RelatedIDs)
		if typed.ApprovalTrace != nil {
			trace := sanitizeCLIValue(*typed.ApprovalTrace).(ApprovalTrace)
			typed.ApprovalTrace = &trace
		}
		typed.Output = sanitizeCLIValue(typed.Output)
		typed.Summary = redactSensitiveText(typed.Summary)
		typed.DecisionComment = redactSensitiveText(typed.DecisionComment)
		return typed
	case ApprovalDecisionResult:
		typed.Request = sanitizeCLIValue(typed.Request).(ApprovalRequest)
		if typed.Invocation != nil {
			invocation := sanitizeCLIValue(*typed.Invocation).(ToolInvocationResult)
			typed.Invocation = &invocation
		}
		return typed
	case ApprovalTrace:
		for index := range typed.Decisions {
			typed.Decisions[index].Comment = redactSensitiveText(typed.Decisions[index].Comment)
		}
		return typed
	case ApprovalTimelineEvent:
		typed.Summary = redactSensitiveText(typed.Summary)
		typed.Metadata = sanitizeCLIMap(typed.Metadata)
		return typed
	case ApprovalTimeline:
		typed.Request = sanitizeCLIValue(typed.Request).(ApprovalRequest)
		if typed.Trace != nil {
			trace := sanitizeCLIValue(*typed.Trace).(ApprovalTrace)
			typed.Trace = &trace
		}
		for index := range typed.Events {
			typed.Events[index] = sanitizeCLIValue(typed.Events[index]).(ApprovalTimelineEvent)
		}
		return typed
	case GovernanceStatus:
		return sanitizeGovernanceStatus(typed)
	case PluginManifest:
		return sanitizeCLIJSONValue(typed)
	case MarketplacePlugin:
		return sanitizeCLIJSONValue(typed)
	case InstalledPlugin:
		return sanitizeCLIJSONValue(typed)
	case []PersonalAccessToken:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case []ServiceAccountToken:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case []ServiceAccount:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case []AuditLog:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case []ApprovalRequest:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case []MarketplacePlugin:
		return sanitizeCLIJSONValue(typed)
	case []InstalledPlugin:
		return sanitizeCLIJSONValue(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeCLIValue(item)
		}
		return out
	case map[string]any:
		return sanitizeCLIMap(typed)
	case string:
		return redactSensitiveText(typed)
	default:
		return typed
	}
}

func sanitizeCLIJSONValue(value any) any {
	raw, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var normalized any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return value
	}
	return sanitizeCLIValue(normalized)
}

func sanitizeGovernanceStatus(status GovernanceStatus) GovernanceStatus {
	status.Health.Message = redactSensitiveText(status.Health.Message)
	for index := range status.Health.Checks {
		status.Health.Checks[index].Message = redactSensitiveText(status.Health.Checks[index].Message)
	}
	for index := range status.Tokens.ExpiringSoon {
		status.Tokens.ExpiringSoon[index] = sanitizeGovernanceTokenFinding(status.Tokens.ExpiringSoon[index])
	}
	for index := range status.Tokens.ExpiredActive {
		status.Tokens.ExpiredActive[index] = sanitizeGovernanceTokenFinding(status.Tokens.ExpiredActive[index])
	}
	for index := range status.Tokens.Stale {
		status.Tokens.Stale[index] = sanitizeGovernanceTokenFinding(status.Tokens.Stale[index])
	}
	for index := range status.Tokens.NeverUsed {
		status.Tokens.NeverUsed[index] = sanitizeGovernanceTokenFinding(status.Tokens.NeverUsed[index])
	}
	for index := range status.Anomalies {
		status.Anomalies[index].Summary = redactSensitiveText(status.Anomalies[index].Summary)
		status.Anomalies[index].ActorID = redactSensitiveText(status.Anomalies[index].ActorID)
		status.Anomalies[index].SubjectID = redactSensitiveText(status.Anomalies[index].SubjectID)
		status.Anomalies[index].AIClientID = redactSensitiveText(status.Anomalies[index].AIClientID)
		status.Anomalies[index].PolicyID = redactSensitiveText(status.Anomalies[index].PolicyID)
		status.Anomalies[index].ApprovalRequestID = redactSensitiveText(status.Anomalies[index].ApprovalRequestID)
		status.Anomalies[index].GrantID = redactSensitiveText(status.Anomalies[index].GrantID)
		status.Anomalies[index].ToolName = redactSensitiveText(status.Anomalies[index].ToolName)
	}
	status.Approvals.OldestPendingRequestID = redactSensitiveText(status.Approvals.OldestPendingRequestID)
	status.Approvals.NextDueRequestID = redactSensitiveText(status.Approvals.NextDueRequestID)
	for index := range status.Approvals.DueSoonRequestIDs {
		status.Approvals.DueSoonRequestIDs[index] = redactSensitiveText(status.Approvals.DueSoonRequestIDs[index])
	}
	for index := range status.Approvals.StalePendingRequestIDs {
		status.Approvals.StalePendingRequestIDs[index] = redactSensitiveText(status.Approvals.StalePendingRequestIDs[index])
	}
	for index := range status.Approvals.OverdueRequestIDs {
		status.Approvals.OverdueRequestIDs[index] = redactSensitiveText(status.Approvals.OverdueRequestIDs[index])
	}
	for index := range status.Recommendations {
		status.Recommendations[index] = redactSensitiveText(status.Recommendations[index])
	}
	for index := range status.RecommendationActions {
		status.RecommendationActions[index] = sanitizeGovernanceRecommendationAction(status.RecommendationActions[index])
	}
	status.Metadata = sanitizeCLIMap(status.Metadata)
	return status
}

func sanitizeGovernanceRecommendationAction(action GovernanceRecommendationAction) GovernanceRecommendationAction {
	action.Summary = redactSensitiveText(action.Summary)
	action.TargetID = redactSensitiveText(action.TargetID)
	for index := range action.Refs {
		action.Refs[index] = redactSensitiveText(action.Refs[index])
	}
	action.Metadata = sanitizeCLIMap(action.Metadata)
	return action
}

func sanitizeGovernanceTokenFinding(finding GovernanceTokenFinding) GovernanceTokenFinding {
	finding.Name = redactSensitiveText(finding.Name)
	finding.OwnerID = redactSensitiveText(finding.OwnerID)
	finding.Message = redactSensitiveText(finding.Message)
	return finding
}

func sanitizeCLIMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		if sensitiveCLIKey(key) {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = sanitizeCLIValue(value)
	}
	return out
}

func sensitiveCLIKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, needle := range []string{"token", "password", "passwd", "secret", "credential", "apikey", "api_key", "authorization", "kubeconfig", "envvar", "environmentvariable"} {
		if strings.Contains(normalized, needle) {
			return true
		}
	}
	return false
}

func redactSensitiveText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	replacements := []string{"token=", "password=", "passwd=", "secret=", "authorization=", "api_key=", "apikey="}
	lower := strings.ToLower(value)
	for _, marker := range replacements {
		if index := strings.Index(lower, marker); index >= 0 {
			end := index + len(marker)
			tail := value[end:]
			if stop := strings.IndexAny(tail, " \t\n,;"); stop >= 0 {
				return value[:end] + "[REDACTED]" + tail[stop:]
			}
			return value[:end] + "[REDACTED]"
		}
	}
	return value
}

func extractLeadingToolName(args []string) (string, []string) {
	return extractLeadingValue(args)
}

func extractLeadingValue(args []string) (string, []string) {
	if len(args) == 0 || strings.HasPrefix(strings.TrimSpace(args[0]), "-") {
		return "", args
	}
	out := append([]string(nil), args[1:]...)
	return strings.TrimSpace(args[0]), out
}

func firstArg(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func approvalText(value bool) string {
	if value {
		return "approval"
	}
	return "direct"
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
