package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Skill install targets.
const (
	targetCrush  = "crush"
	targetClaude = "claude"
	targetCodex  = "codex"
	targetCursor = "cursor"
)

// skillInstaller resolves the destination path and formats content for a given agent.
type skillInstaller struct {
	path    func() (string, error)
	content func() string
}

// crushSkillContent returns the standard Crush/Claude skill format.
func crushSkillContent() string {
	return popSkill
}

// cursorSkillContent returns the Cursor .mdc format with cursor-compatible frontmatter.
func cursorSkillContent() string {
	return "---\n" +
		"description: '" + popSkillDescription + "'\n" +
		"globs:\n" +
		"alwaysApply: false\n" +
		"---\n\n" +
		popSkillBody
}

var skillInstallers = map[string]skillInstaller{
	targetCrush: {
		path: func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolving home directory: %w", err)
			}
			return filepath.Join(home, ".config", "crush", "skills", "pop", "SKILL.md"), nil
		},
		content: crushSkillContent,
	},
	targetClaude: {
		path: func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolving home directory: %w", err)
			}
			return filepath.Join(home, ".claude", "skills", "pop", "SKILL.md"), nil
		},
		content: crushSkillContent,
	},
	targetCursor: {
		path: func() (string, error) {
			return filepath.Join(".cursor", "rules", "pop.mdc"), nil
		},
		content: cursorSkillContent,
	},
	targetCodex: {
		path: func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolving home directory: %w", err)
			}
			return filepath.Join(home, ".codex", "AGENTS.md"), nil
		},
		content: func() string { return popSkillBody },
	},
}

// skillDisplayNames maps internal target names to their product display names.
var skillDisplayNames = map[string]string{
	targetCrush:  "Charm Crush",
	targetClaude: "Claude Code",
	targetCodex:  "OpenAI Codex",
	targetCursor: "Cursor",
}

// skillTargetOrder is the order targets appear in help output.
var skillTargetOrder = []string{targetCrush, targetClaude, targetCodex, targetCursor}

var installSkillForce bool

// InstallSkillCmd is the parent command for installing the pop skill to AI agents.
var InstallSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the Pop skill for AI agents",
	Long:  `Install the Pop skill definition for an AI agent.`,
	Args:  cobra.NoArgs,
}

// newInstallSkillTargetCmd creates a subcommand for installing to a specific agent.
func newInstallSkillTargetCmd(target string) *cobra.Command {
	name := skillDisplayNames[target]
	short := fmt.Sprintf("Install the Pop skill for %s", name)
	if target == targetCursor {
		short = fmt.Sprintf("Install the Pop skill for %s (in the current directory)", name)
	}
	return &cobra.Command{
		Use:   target,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return installSkill(target)
		},
	}
}

func installSkill(target string) error {
	name := skillDisplayNames[target]
	inst := skillInstallers[target]
	path, err := inst.path()
	if err != nil {
		fmt.Printf("  %s Failed to resolve path for %s: %s\n",
			errorHeaderStyle.String(), name, err)
		return fmt.Errorf("resolving path: %w", err)
	}

	if _, err := os.Stat(path); err == nil && !installSkillForce {
		fmt.Printf("  %s Already installed at %s (use --force to overwrite)\n",
			errorHeaderStyle.String(),
			path)
		return fmt.Errorf("skill already installed")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		fmt.Printf("  %s Failed to create directory for %s: %s\n",
			errorHeaderStyle.String(), name, err)
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(inst.content()), 0o600); err != nil {
		fmt.Printf("  %s Failed to write %s: %s\n",
			errorHeaderStyle.String(), name, err)
		return fmt.Errorf("writing skill: %w", err)
	}

	fmt.Printf("  %s Installed skill to %s\n",
		activeLabelStyle.Render("OK"),
		path)
	return nil
}

func init() {
	InstallSkillCmd.PersistentFlags().BoolVarP(&installSkillForce, "force", "f", false, "Overwrite existing skill files")
	cobra.EnableCommandSorting = false
	for _, target := range skillTargetOrder {
		InstallSkillCmd.AddCommand(newInstallSkillTargetCmd(target))
	}
}
