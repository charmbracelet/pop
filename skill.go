package main

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

// popSkillDescription is the description used in skill frontmatter.
const popSkillDescription = `Send emails from the terminal. Use when the user asks to send, draft, or compose an email, attach files to an email, configure email delivery via Resend or SMTP, or asks about pop commands and environment variables. Triggers on mentions of "pop" in the context of email, mailing, SMTP, or Resend.`

//go:embed skill.md
var popSkillBody string

// popSkill is the full Crush-compatible skill definition (frontmatter + body).
var popSkill = "---\n" +
	"name: pop\n" +
	"description: '" + popSkillDescription + "'\n" +
	"---\n\n" +
	popSkillBody

// SkillCmd prints a Crush-compatible skill definition to stdout so AI agents
// can discover and use pop non-interactively.
var SkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print a skill definition for AI agents",
	Long:  `Print a Crush-compatible skill that teaches AI agents how to use pop non-interactively.`,
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Print(popSkill)
		return nil
	},
}
