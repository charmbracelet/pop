package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSecretStringVarPDoesNotLeakEnvValue(t *testing.T) {
	const secret = "hunter2"
	t.Setenv(PopSMTPPassword, secret)

	var password string
	cmd := &cobra.Command{Use: "pop"}
	secretStringVarP(cmd, &password, "smtp.password", "p", PopSMTPPassword, "Password of the SMTP server")

	if password != secret {
		t.Fatalf("expected value from env %q, got %q", secret, password)
	}
	if usage := cmd.Flags().FlagUsages(); strings.Contains(usage, secret) {
		t.Fatalf("flag usage leaks secret:\n%s", usage)
	}

	if err := cmd.Flags().Parse([]string{"--smtp.password", "override"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if password != "override" {
		t.Fatalf("expected flag to override env value, got %q", password)
	}
}

func TestRootCmdSecretFlagsHaveNoDefault(t *testing.T) {
	for _, name := range []string{"smtp.password", "resend.key"} {
		f := rootCmd.Flags().Lookup(name)
		if f == nil {
			t.Fatalf("flag %q not found", name)
		}
		if f.DefValue != "" {
			t.Errorf("flag %q has non-empty default %q", name, f.DefValue)
		}
	}
}
