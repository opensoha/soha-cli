package sohacli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type addTargetSpec struct {
	Name           string
	DisplayName    string
	ConfigKind     string
	DefaultConfig  func() string
	DefaultSkill   func() string
	JSONC          bool
	JSONDisabled   bool
	RestartMessage string
}

func runAdd(args []string, rt Runtime) error {
	if len(args) == 0 || strings.HasPrefix(strings.TrimSpace(args[0]), "-") {
		targets, err := promptAddTargets(rt)
		if err != nil {
			return err
		}
		for _, spec := range targets {
			if err := runAddTarget(args, rt, spec); err != nil {
				return err
			}
		}
		return nil
	}
	targetName := strings.ToLower(strings.TrimSpace(args[0]))
	if targetName == "all" {
		for _, spec := range addAllTargets() {
			if err := runAddTarget(args[1:], rt, spec); err != nil {
				return err
			}
		}
		return nil
	}
	spec, ok := addTarget(targetName)
	if !ok {
		return fmt.Errorf("unknown add target %q", args[0])
	}
	return runAddTarget(args[1:], rt, spec)
}

func promptAddTargets(rt Runtime) ([]addTargetSpec, error) {
	targets := addSelectableTargets()
	fmt.Fprintln(rt.Out, "Select AI agents or IDEs to add Soha MCP and skills:")
	for i, spec := range targets {
		fmt.Fprintf(rt.Out, "  %d) %s (%s)\n", i+1, spec.DisplayName, spec.Name)
	}
	fmt.Fprint(rt.Out, "Enter numbers or names separated by commas, or all: ")
	line, err := bufio.NewReader(rt.In).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return nil, err
	}
	return parseAddTargetSelection(line, targets)
}

func parseAddTargetSelection(input string, targets []addTargetSpec) ([]addTargetSpec, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("no add targets selected")
	}
	if strings.EqualFold(input, "all") {
		return targets, nil
	}
	byName := map[string]addTargetSpec{}
	for _, spec := range targets {
		byName[spec.Name] = spec
		byName[strings.ToLower(spec.DisplayName)] = spec
	}
	var selected []addTargetSpec
	seen := map[string]bool{}
	for _, token := range strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	}) {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			continue
		}
		if token == "all" {
			return targets, nil
		}
		var spec addTargetSpec
		if index, ok := parsePositiveIndex(token); ok {
			if index < 1 || index > len(targets) {
				return nil, fmt.Errorf("target number %d is out of range", index)
			}
			spec = targets[index-1]
		} else {
			var ok bool
			spec, ok = addTarget(token)
			if !ok {
				spec, ok = byName[token]
			}
			if !ok {
				return nil, fmt.Errorf("unknown add target %q", token)
			}
		}
		if !seen[spec.Name] {
			selected = append(selected, spec)
			seen[spec.Name] = true
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no add targets selected")
	}
	return selected, nil
}

func parsePositiveIndex(value string) (int, bool) {
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, value != ""
}

func runAddCodex(args []string, rt Runtime) error {
	return runAddTarget(args, rt, mustAddTarget("codex"))
}

func runAddTarget(args []string, rt Runtime, target addTargetSpec) error {
	fs := newRuntimeFlagSet("add "+target.Name, args, rt)
	profileFlag := fs.String("profile", "", "profile name for the generated MCP server")
	serverFlag := fs.String("server", "", "Soha API base URL for the profile")
	baseURLFlag := fs.String("base-url", "", "alias for --server")
	commandFlag := fs.String("command", "", "soha executable path for the generated MCP server")
	aiClientIDFlag := fs.String("ai-client-id", "", "AI client id for Gateway audit context")
	aiClientNameFlag := fs.String("ai-client", "", "AI client display name for Gateway audit context")
	skillIDFlag := fs.String("skill-id", "", "skill id for the generated MCP server")
	sourceFlag := fs.String("skills-source", defaultSkillSourcePath(), "source Soha skill directory")
	sourceAliasFlag := fs.String("source", "", "alias for --skills-source")
	skillsFlag := fs.String("skills", "all", "comma-separated skill ids to install, or all")
	codexHomeFlag := fs.String("codex-home", defaultCodexHome(), "Codex home containing config.toml")
	codexConfigFlag := fs.String("codex-config", "", "Codex config.toml path")
	configFlag := fs.String("config", "", "target MCP config path")
	agentsHomeFlag := fs.String("agents-home", defaultAgentsHome(), "Codex agents home containing skills")
	skillDestFlag := fs.String("skill-dest", "", "Codex skill destination directory")
	destAliasFlag := fs.String("dest", "", "alias for --skill-dest")
	runtimeSkillDestFlag := fs.String("runtime-skill-dest", defaultSkillInstallPath(), "Soha runtime skill destination directory")
	mcpNameFlag := fs.String("mcp-name", "soha", "Codex MCP server name")
	noRuntimeSkillsFlag := fs.Bool("no-runtime-skills", false, "skip installing raw Soha runtime skills")
	overwriteFlag := fs.Bool("overwrite", true, "overwrite generated skill files")
	dryRunFlag := fs.Bool("dry-run", false, "print planned changes without writing files")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(rt.ConfigPath)
	if err != nil {
		return err
	}
	profileNameValue := profileName(firstNonEmptyString(*profileFlag, cfg.CurrentProfile))
	profileConfig, profileExists := cfg.Profiles[profileNameValue]
	command := strings.TrimSpace(*commandFlag)
	if command == "" {
		command = defaultSohaCommand()
	}
	aiClientID := firstNonEmptyString(*aiClientIDFlag, profileConfig.AIClientID, "codex-local")
	aiClientName := firstNonEmptyString(*aiClientNameFlag, profileConfig.AIClientName, "Codex")
	skillID := firstNonEmptyString(*skillIDFlag, profileConfig.SkillID)
	serverURL := normalizeServerURL(firstNonEmptyString(*serverFlag, *baseURLFlag, profileConfig.ServerURL, defaultServerURL))
	if serverURL == "" {
		return fmt.Errorf("Soha API base URL is required")
	}
	updatedProfile := profileConfig
	updatedProfile.ServerURL = serverURL
	if strings.TrimSpace(updatedProfile.AIClientID) == "" || strings.TrimSpace(*aiClientIDFlag) != "" {
		updatedProfile.AIClientID = aiClientID
	}
	if strings.TrimSpace(updatedProfile.AIClientName) == "" || strings.TrimSpace(*aiClientNameFlag) != "" {
		updatedProfile.AIClientName = aiClientName
	}
	if strings.TrimSpace(updatedProfile.SkillID) == "" || strings.TrimSpace(*skillIDFlag) != "" {
		updatedProfile.SkillID = skillID
	}
	if strings.TrimSpace(updatedProfile.Source) == "" {
		updatedProfile.Source = "codex"
	}
	source := firstNonEmptyString(*sourceAliasFlag, *sourceFlag)
	if source == "" {
		return fmt.Errorf("skill source directory is required")
	}
	source = normalizeSkillSourcePath(source)
	mcpName := strings.TrimSpace(*mcpNameFlag)
	if !isSafeSkillName(mcpName) {
		return fmt.Errorf("invalid MCP server name %q", mcpName)
	}

	skillNames, err := addSkillNames(source, *skillsFlag)
	if err != nil {
		return err
	}
	codexHome := expandHome(*codexHomeFlag)
	agentsHome := expandHome(*agentsHomeFlag)
	defaultSkillDest := target.DefaultSkill()
	if target.Name == "codex" {
		defaultSkillDest = filepath.Join(agentsHome, "skills")
	}
	skillDest := firstNonEmptyString(*destAliasFlag, *skillDestFlag, defaultSkillDest)
	skillDest = expandHome(skillDest)
	runtimeSkillDest := expandHome(*runtimeSkillDestFlag)
	targetConfigPath := target.DefaultConfig()
	if target.Name == "codex" {
		targetConfigPath = filepath.Join(codexHome, "config.toml")
	}
	targetConfigPath = expandHome(firstNonEmptyString(*configFlag, *codexConfigFlag, targetConfigPath))
	mcpArgs := mcpInstallArgs(profileNameValue, aiClientID, aiClientName, skillID)
	mcpBlock := codexMCPBlock(mcpName, command, mcpArgs)

	if *dryRunFlag {
		if !profileExists || updatedProfile != profileConfig {
			fmt.Fprintf(rt.Out, "Would configure Soha profile %s with API base URL: %s\n", profileNameValue, serverURL)
		} else {
			fmt.Fprintf(rt.Out, "Soha profile %s already uses API base URL: %s\n", profileNameValue, serverURL)
		}
		if target.ConfigKind == "toml" {
			fmt.Fprintf(rt.Out, "Would update %s MCP config: %s\n\n%s\n", target.DisplayName, targetConfigPath, mcpBlock)
		} else {
			fmt.Fprintf(rt.Out, "Would update %s MCP config: %s\n", target.DisplayName, targetConfigPath)
			fmt.Fprintf(rt.Out, "Would set mcpServers.%s command=%s args=%s\n", mcpName, command, strings.Join(mcpArgs, " "))
		}
		fmt.Fprintf(rt.Out, "Would install %s Soha skill package: %s\n", target.DisplayName, filepath.Join(skillDest, "soha", "SKILL.md"))
		fmt.Fprintf(rt.Out, "Would copy skill references from %s: %s\n", source, strings.Join(skillNames, ", "))
		if !*noRuntimeSkillsFlag {
			fmt.Fprintf(rt.Out, "Would install Soha runtime skills to %s: %s\n", runtimeSkillDest, strings.Join(skillNames, ", "))
		}
		return nil
	}

	if !profileExists || updatedProfile != profileConfig {
		cfg.Profiles[profileNameValue] = updatedProfile
		if err := saveConfig(rt.ConfigPath, cfg); err != nil {
			return err
		}
		fmt.Fprintf(rt.Out, "Configured Soha profile %s with API base URL %s\n", profileNameValue, serverURL)
	} else {
		fmt.Fprintf(rt.Out, "Soha profile %s already uses API base URL %s\n", profileNameValue, serverURL)
	}

	var backupPath string
	var changed bool
	if target.ConfigKind == "toml" {
		backupPath, changed, err = upsertCodexMCPConfig(targetConfigPath, "mcp_servers."+mcpName, mcpBlock)
		if err != nil {
			return err
		}
	} else {
		backupPath, changed, err = upsertJSONMCPConfig(targetConfigPath, mcpName, command, mcpArgs, target.JSONC, target.JSONDisabled)
		if err != nil {
			return err
		}
	}
	if changed {
		fmt.Fprintf(rt.Out, "Configured %s MCP server %s in %s\n", target.DisplayName, mcpName, targetConfigPath)
		if backupPath != "" {
			fmt.Fprintf(rt.Out, "Backed up previous %s config to %s\n", target.DisplayName, backupPath)
		}
	} else {
		fmt.Fprintf(rt.Out, "%s MCP server %s already up to date in %s\n", target.DisplayName, mcpName, targetConfigPath)
	}

	installed, err := installSohaSkillPackage(source, skillDest, skillNames, target.Name, profileNameValue, command, *overwriteFlag)
	if err != nil {
		return err
	}
	for _, path := range installed {
		fmt.Fprintf(rt.Out, "Installed %s Soha skill file %s\n", target.DisplayName, path)
	}

	if !*noRuntimeSkillsFlag {
		for _, name := range skillNames {
			installedPath, err := installLocalSkill(source, runtimeSkillDest, name, *overwriteFlag)
			if err != nil {
				return err
			}
			fmt.Fprintf(rt.Out, "Installed Soha runtime skill %s to %s\n", name, installedPath)
		}
	}

	if strings.TrimSpace(target.RestartMessage) != "" {
		fmt.Fprintln(rt.Out, target.RestartMessage)
	} else {
		fmt.Fprintf(rt.Out, "Restart %s or open a new session so MCP servers and skills are reloaded.\n", target.DisplayName)
	}
	return nil
}

func addTarget(name string) (addTargetSpec, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "codex", "openai":
		return addTargetSpec{
			Name:           "codex",
			DisplayName:    "Codex",
			ConfigKind:     "toml",
			DefaultConfig:  func() string { return filepath.Join(defaultCodexHome(), "config.toml") },
			DefaultSkill:   func() string { return filepath.Join(defaultAgentsHome(), "skills") },
			RestartMessage: "Restart Codex or open a new Codex session so MCP servers and skills are reloaded.",
		}, true
	case "claude", "claude-desktop":
		return addTargetSpec{
			Name:        "claude",
			DisplayName: "Claude",
			ConfigKind:  "json",
			DefaultConfig: func() string {
				return filepath.Join(userHomeDirFallback(), "Library", "Application Support", "Claude", "claude_desktop_config.json")
			},
			DefaultSkill: func() string { return filepath.Join(userHomeDirFallback(), ".claude", "skills") },
		}, true
	case "cursor":
		return addTargetSpec{
			Name:          "cursor",
			DisplayName:   "Cursor",
			ConfigKind:    "json",
			DefaultConfig: func() string { return filepath.Join(userHomeDirFallback(), ".cursor", "mcp.json") },
			DefaultSkill:  func() string { return filepath.Join(userHomeDirFallback(), ".cursor", "skills-cursor") },
		}, true
	case "kiro":
		return addTargetSpec{
			Name:          "kiro",
			DisplayName:   "Kiro",
			ConfigKind:    "json",
			DefaultConfig: func() string { return filepath.Join(userHomeDirFallback(), ".kiro", "settings", "mcp.json") },
			DefaultSkill:  func() string { return filepath.Join(userHomeDirFallback(), ".kiro", "steering") },
			JSONC:         true,
			JSONDisabled:  true,
		}, true
	case "gemini", "gemini-cli":
		return addTargetSpec{
			Name:          "gemini",
			DisplayName:   "Gemini",
			ConfigKind:    "json",
			DefaultConfig: func() string { return filepath.Join(userHomeDirFallback(), ".gemini", "settings.json") },
			DefaultSkill:  func() string { return filepath.Join(userHomeDirFallback(), ".gemini", "skills") },
		}, true
	case "antigravity", "anti-gravity", "反重力":
		return addTargetSpec{
			Name:        "antigravity",
			DisplayName: "Antigravity",
			ConfigKind:  "json",
			DefaultConfig: func() string {
				return filepath.Join(userHomeDirFallback(), ".gemini", "antigravity", "mcp_config.json")
			},
			DefaultSkill: func() string { return filepath.Join(userHomeDirFallback(), ".gemini", "antigravity", "skills") },
		}, true
	case "antigravity-ide":
		return addTargetSpec{
			Name:        "antigravity-ide",
			DisplayName: "Antigravity IDE",
			ConfigKind:  "json",
			DefaultConfig: func() string {
				return filepath.Join(userHomeDirFallback(), ".gemini", "antigravity-ide", "mcp_config.json")
			},
			DefaultSkill: func() string { return filepath.Join(userHomeDirFallback(), ".gemini", "antigravity-ide", "skills") },
		}, true
	case "trae":
		return addTargetSpec{
			Name:        "trae",
			DisplayName: "Trae",
			ConfigKind:  "json",
			DefaultConfig: func() string {
				return filepath.Join(userHomeDirFallback(), "Library", "Application Support", "Trae", "User", "mcp.json")
			},
			DefaultSkill: func() string { return filepath.Join(userHomeDirFallback(), ".trae", "skills") },
			JSONDisabled: true,
		}, true
	default:
		return addTargetSpec{}, false
	}
}

func mustAddTarget(name string) addTargetSpec {
	spec, ok := addTarget(name)
	if !ok {
		panic("unknown built-in add target " + name)
	}
	return spec
}

func addAllTargets() []addTargetSpec {
	return addSelectableTargets()
}

func addSelectableTargets() []addTargetSpec {
	return []addTargetSpec{
		mustAddTarget("codex"),
		mustAddTarget("claude"),
		mustAddTarget("cursor"),
		mustAddTarget("kiro"),
		mustAddTarget("gemini"),
		mustAddTarget("antigravity"),
		mustAddTarget("antigravity-ide"),
		mustAddTarget("trae"),
	}
}

func addSkillNames(source, value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "all") {
		return listLocalSkills(source)
	}
	names := splitCSV(value)
	if len(names) == 0 {
		return nil, fmt.Errorf("--skills requires at least one skill id or all")
	}
	for _, name := range names {
		if !isSafeSkillName(name) {
			return nil, fmt.Errorf("invalid skill id %q", name)
		}
		if _, err := os.Stat(filepath.Join(source, name, "SKILL.md")); err != nil {
			return nil, err
		}
	}
	sort.Strings(names)
	return names, nil
}

func defaultSohaCommand() string {
	if path, err := os.Executable(); err == nil {
		path = strings.TrimSpace(path)
		if path != "" && strings.HasPrefix(filepath.Base(path), "soha") {
			return path
		}
	}
	if path, err := exec.LookPath("soha"); err == nil {
		return path
	}
	return "soha"
}

func defaultCodexHome() string {
	if value := env("CODEX_HOME"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(".codex")
	}
	return filepath.Join(home, ".codex")
}

func defaultAgentsHome() string {
	if value := env("AGENTS_HOME"); value != "" {
		return value
	}
	home := userHomeDirFallback()
	if strings.TrimSpace(home) == "" {
		return filepath.Join(".agents")
	}
	return filepath.Join(home, ".agents")
}

func userHomeDirFallback() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return home
}

func expandHome(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return path
		}
		if path == "~" {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}

func codexMCPBlock(name, command string, args []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[mcp_servers.%s]\n", name)
	fmt.Fprintln(&b, `type = "stdio"`)
	fmt.Fprintf(&b, "command = %s\n", tomlString(command))
	fmt.Fprintf(&b, "args = %s\n", tomlStringArray(args))
	return b.String()
}

func tomlStringArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, tomlString(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func tomlString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\t", `\t`)
	return `"` + value + `"`
}

func upsertCodexMCPConfig(path, section, block string) (string, bool, error) {
	var raw []byte
	var mode fs.FileMode = 0o600
	stat, statErr := os.Stat(path)
	if statErr == nil {
		mode = stat.Mode().Perm()
		var err error
		raw, err = os.ReadFile(path)
		if err != nil {
			return "", false, err
		}
	} else if !os.IsNotExist(statErr) {
		return "", false, statErr
	}

	next := upsertTOMLSection(string(raw), section, strings.TrimRight(block, "\n"))
	if next == string(raw) {
		return "", false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", false, err
	}
	backupPath := ""
	if statErr == nil {
		backupPath = fmt.Sprintf("%s.bak.%s", path, time.Now().Format("20060102150405"))
		if err := os.WriteFile(backupPath, raw, mode); err != nil {
			return "", false, err
		}
	}
	if err := os.WriteFile(path, []byte(next), mode); err != nil {
		return "", false, err
	}
	return backupPath, true, nil
}

func upsertJSONMCPConfig(path, serverName, command string, args []string, jsonc bool, includeDisabled bool) (string, bool, error) {
	raw, mode, statErr, err := readOptionalFile(path)
	if err != nil {
		return "", false, err
	}
	root := map[string]any{}
	if len(bytes.TrimSpace(raw)) > 0 {
		decodeRaw := raw
		if jsonc {
			decodeRaw = stripJSONComments(raw)
		}
		if err := json.Unmarshal(decodeRaw, &root); err != nil {
			return "", false, fmt.Errorf("decode %s: %w", path, err)
		}
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		root["mcpServers"] = servers
	}
	entry := map[string]any{
		"command": command,
		"args":    args,
	}
	if includeDisabled {
		entry["disabled"] = false
	}
	servers[serverName] = entry
	next, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", false, err
	}
	next = append(next, '\n')
	if bytes.Equal(bytes.TrimSpace(raw), bytes.TrimSpace(next)) {
		return "", false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", false, err
	}
	backupPath := ""
	if statErr == nil {
		backupPath = fmt.Sprintf("%s.bak.%s", path, time.Now().Format("20060102150405"))
		if err := os.WriteFile(backupPath, raw, mode); err != nil {
			return "", false, err
		}
	}
	if err := os.WriteFile(path, next, mode); err != nil {
		return "", false, err
	}
	return backupPath, true, nil
}

func readOptionalFile(path string) ([]byte, fs.FileMode, error, error) {
	mode := fs.FileMode(0o600)
	stat, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, mode, statErr, nil
		}
		return nil, mode, statErr, statErr
	}
	mode = stat.Mode().Perm()
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, mode, statErr, err
	}
	return raw, mode, statErr, nil
}

func stripJSONComments(raw []byte) []byte {
	out := make([]byte, 0, len(raw))
	inString := false
	escaped := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if inString {
			out = append(out, ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			out = append(out, ch)
			continue
		}
		if ch == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			for i < len(raw) && raw[i] != '\n' {
				i++
			}
			if i < len(raw) {
				out = append(out, raw[i])
			}
			continue
		}
		if ch == '/' && i+1 < len(raw) && raw[i+1] == '*' {
			i += 2
			for i+1 < len(raw) && !(raw[i] == '*' && raw[i+1] == '/') {
				if raw[i] == '\n' {
					out = append(out, '\n')
				}
				i++
			}
			i++
			continue
		}
		out = append(out, ch)
	}
	return out
}

func upsertTOMLSection(raw, section, block string) string {
	header := "[" + section + "]"
	lines := strings.Split(raw, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			break
		}
	}
	blockLines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	if start >= 0 {
		end := len(lines)
		for i := start + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				end = i
				break
			}
		}
		next := append([]string{}, lines[:start]...)
		next = append(next, blockLines...)
		next = append(next, lines[end:]...)
		out := strings.Join(next, "\n")
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		return out
	}
	if strings.TrimSpace(raw) == "" {
		return strings.Join(blockLines, "\n") + "\n"
	}
	return strings.TrimRight(raw, "\n") + "\n\n" + strings.Join(blockLines, "\n") + "\n"
}

func installSohaSkillPackage(source, dest string, skillNames []string, target, profile, command string, overwrite bool) ([]string, error) {
	root := filepath.Join(dest, "soha")
	writes := map[string][]byte{
		filepath.Join(root, "SKILL.md"):              []byte(codexSohaSkillMarkdown(target, skillNames)),
		filepath.Join(root, "agents", "openai.yaml"): []byte(codexSohaOpenAIYAML(target, profile, command)),
	}
	indexPath := filepath.Join(filepath.Dir(source), "index.json")
	if raw, err := os.ReadFile(indexPath); err == nil {
		writes[filepath.Join(root, "references", "skills", "index.json")] = raw
	}
	for _, name := range skillNames {
		raw, err := os.ReadFile(filepath.Join(source, name, "SKILL.md"))
		if err != nil {
			return nil, err
		}
		writes[filepath.Join(root, "references", "skills", name+".md")] = raw
	}

	paths := make([]string, 0, len(writes))
	for path := range writes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeGeneratedFile(path, writes[path], overwrite); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func writeGeneratedFile(path string, raw []byte, overwrite bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}
	file, err := os.OpenFile(path, flag, fs.FileMode(0o644))
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists; pass --overwrite=true to replace it", path)
		}
		return err
	}
	defer file.Close()
	if _, err := file.Write(raw); err != nil {
		return err
	}
	return nil
}

func codexSohaSkillMarkdown(target string, skillNames []string) string {
	references := make([]string, 0, len(skillNames))
	for _, name := range skillNames {
		references = append(references, fmt.Sprintf("   - `%s.md`", name))
	}
	addCommand := "soha add " + target
	return strings.TrimSpace(`---
name: soha
description: >
  Use when a task involves OpenSoha / Soha AI Gateway through the local soha
  CLI or MCP server: configuring Codex MCP, checking profiles and capabilities,
  installing Soha skills, invoking Gateway tools/resources/prompts, or using
  delivery, release-testing, Kubernetes SRE, and security-change workflows.
allowed-tools:
  - Bash(which soha)
  - Bash(soha version*)
  - Bash(soha profile*)
  - Bash(soha context*)
  - Bash(soha capabilities*)
  - Bash(soha diagnose*)
  - Bash(soha mcp*)
  - Bash(soha skill*)
  - Bash(soha tool call*)
  - Bash(soha resource read*)
  - Bash(soha prompt get*)
  - Bash(soha approval*)
  - Bash(soha audit*)
  - Bash(soha governance*)
---

# Soha CLI

Use the local `+"`soha`"+` CLI as the source of truth for OpenSoha AI Gateway work. Do not bypass Gateway with direct Kubernetes, CI, database, runner, or deployment-target commands.

## Setup

Check the CLI and active profile first:

`+"```bash"+`
which soha
soha version --json
soha profile list
soha context show
`+"```"+`

If MCP is not configured, use `+"`"+addCommand+"`"+` to write the generated `+"`mcpServers.soha`"+` entry for this client.

## Workflow

1. Use `+"`soha capabilities --output names`"+` to see visible Gateway tools.
2. Use `+"`soha capabilities --output inputs`"+` before invoking an unfamiliar tool.
3. Use `+"`soha diagnose --tool <name>`"+` when access, scope, MCP grants, or skill bindings are unclear.
4. For product workflows, read the matching reference under `+"`references/skills/`"+`:
`+strings.Join(references, "\n")+`
5. Keep `+"`cluster`"+`, `+"`namespace`"+`, `+"`application`"+`, `+"`environment`"+`, and related IDs explicit in tool calls and final answers.
6. Treat missing capabilities, denied grants, approval-required responses, and capability warnings as real constraints.

## Guardrails

- Never request or expose tokens, kubeconfig, passwords, private keys, registry credentials, environment secrets, or raw secret-looking logs.
- Do not run `+"`kubectl`"+`, CI, Docker, PostgreSQL, or runner commands as a workaround for Gateway permissions.
- Mutating delivery or security actions must go through visible Gateway tools and explicit user intent.
`) + "\n"
}

func codexSohaOpenAIYAML(target, profile, command string) string {
	return fmt.Sprintf("interface:\n  display_name: \"Soha\"\n  short_description: \"OpenSoha AI Gateway CLI and MCP workflows\"\n  default_prompt: \"Use $soha to inspect Soha Gateway capabilities or configure MCP.\"\nmetadata:\n  target: %s\n  profile: %s\n  command: %s\n", tomlString(target), tomlString(profile), tomlString(command))
}
