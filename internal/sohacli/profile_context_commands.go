package sohacli

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

func runProfile(args []string, rt Runtime) error {
	if len(args) == 0 {
		args = []string{"show"}
	}
	cfg, err := loadConfig(rt.ConfigPath)
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		names := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			marker := " "
			if name == cfg.CurrentProfile {
				marker = "*"
			}
			fmt.Fprintf(rt.Out, "%s %s\t%s\n", marker, name, cfg.Profiles[name].ServerURL)
		}
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("profile use requires a profile name")
		}
		name := profileName(args[1])
		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("profile %q is not configured", name)
		}
		cfg.CurrentProfile = name
		if err := saveConfig(rt.ConfigPath, cfg); err != nil {
			return err
		}
		fmt.Fprintf(rt.Out, "Current profile: %s\n", name)
	case "show":
		name := profileName(firstArg(args[1:], cfg.CurrentProfile))
		profile, ok := cfg.Profiles[name]
		if !ok {
			return fmt.Errorf("profile %q is not configured", name)
		}
		view := profile
		view.AccessToken = redactToken(view.AccessToken)
		view.RefreshToken = redactToken(view.RefreshToken)
		return writePrettyJSON(rt.Out, map[string]any{"name": name, "profile": view})
	default:
		return fmt.Errorf("unknown profile command %q", args[0])
	}
	return nil
}

func runContext(_ context.Context, args []string, rt Runtime) error {
	if len(args) == 0 {
		args = []string{"show"}
	}
	switch args[0] {
	case "show":
		fs := newRuntimeFlagSet("context show", args[1:], rt)
		profileFlag := fs.String("profile", "", "profile name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		_, name, profile, err := loadLocalRuntimeProfile(rt, *profileFlag)
		if err != nil {
			return err
		}
		return writePrettyJSON(rt.Out, map[string]any{
			"profile":      name,
			"serverUrl":    profile.ServerURL,
			"aiClientId":   profile.AIClientID,
			"aiClientName": profile.AIClientName,
			"skillId":      profile.SkillID,
			"source":       profile.Source,
		})
	case "set":
		fs := newRuntimeFlagSet("context set", args[1:], rt)
		profileFlag := fs.String("profile", "", "profile name")
		aiClientID := fs.String("ai-client-id", "", "AI client id")
		aiClientName := fs.String("ai-client", "", "AI client display name")
		skillID := fs.String("skill-id", "", "skill id")
		source := fs.String("source", "", "request source label")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		cfg, name, profile, err := loadLocalRuntimeProfile(rt, *profileFlag)
		if err != nil {
			return err
		}
		if *aiClientID != "" {
			profile.AIClientID = strings.TrimSpace(*aiClientID)
		}
		if *aiClientName != "" {
			profile.AIClientName = strings.TrimSpace(*aiClientName)
		}
		if *skillID != "" {
			profile.SkillID = strings.TrimSpace(*skillID)
		}
		if *source != "" {
			profile.Source = strings.TrimSpace(*source)
		}
		cfg.Profiles[name] = profile
		if err := saveConfig(rt.ConfigPath, cfg); err != nil {
			return err
		}
		fmt.Fprintf(rt.Out, "Updated context for profile %s\n", name)
		return nil
	default:
		return fmt.Errorf("unknown context command %q", args[0])
	}
}

func loadLocalRuntimeProfile(rt Runtime, requested string) (Config, string, ProfileConfig, error) {
	cfg, err := loadConfig(rt.ConfigPath)
	if err != nil {
		return Config{}, "", ProfileConfig{}, err
	}
	name, profile, err := resolveProfile(cfg, requested)
	if err != nil {
		return Config{}, "", ProfileConfig{}, err
	}
	if strings.TrimSpace(profile.Source) == "" {
		profile.Source = "soha"
	}
	return cfg, name, profile, nil
}
