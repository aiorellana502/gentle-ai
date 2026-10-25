package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a small local helper so each fixture below stays one line.
func writeOrphanFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDoctor_OrphanedManagedContent_NoRecordedSelection(t *testing.T) {
	home := t.TempDir()
	writeOrphanFixtureFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "<!-- gentle-ai:persona -->\nbody\n<!-- /gentle-ai:persona -->\n")
	output := runDoctorAgentConfigFixture(t, home, `{}`)
	if !strings.Contains(output, "  [ok]  sync:orphaned_content          no recorded agent selection — not applicable\n") {
		t.Fatalf("expected the check to stay silent without a recorded selection:\n%s", output)
	}
}

func TestRunDoctor_OrphanedManagedContent_ExcludedAgentStillCarriesASection(t *testing.T) {
	home := t.TempDir()
	writeOrphanFixtureFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "<!-- gentle-ai:persona -->\nbody\n<!-- /gentle-ai:persona -->\n")
	output := runDoctorAgentConfigFixture(t, home, `{"installed_agents":["pi"],"selection_configured":true}`)
	claudePromptPath := filepath.Join(home, ".claude", "CLAUDE.md")
	want := "  [!!]  sync:orphaned_content          1 excluded agent(s) still carry a gentle-ai managed section in their system prompt: claude-code (" + claudePromptPath + ")\n"
	if !strings.Contains(output, want) {
		t.Fatalf("missing orphan warning %q:\n%s", want, output)
	}
	if !strings.Contains(output, "Remedy: Review and remove the gentle-ai managed section(s)") {
		t.Fatalf("missing orphan remedy:\n%s", output)
	}
}

func TestRunDoctor_OrphanedManagedContent_ExcludedAgentWithPlainUserContent(t *testing.T) {
	home := t.TempDir()
	writeOrphanFixtureFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "# My own rules\n- Answer in English.\n")
	output := runDoctorAgentConfigFixture(t, home, `{"installed_agents":["pi"],"selection_configured":true}`)
	if !strings.Contains(output, "  [ok]  sync:orphaned_content          no excluded agent carries leftover gentle-ai managed content\n") {
		t.Fatalf("plain user content must never be reported as an orphan:\n%s", output)
	}
}

func TestRunDoctor_OrphanedManagedContent_ExcludedAgentWithNoPromptFile(t *testing.T) {
	home := t.TempDir()
	output := runDoctorAgentConfigFixture(t, home, `{"installed_agents":["pi"],"selection_configured":true}`)
	if !strings.Contains(output, "  [ok]  sync:orphaned_content          no excluded agent carries leftover gentle-ai managed content\n") {
		t.Fatalf("an absent prompt file must never be reported as an orphan:\n%s", output)
	}
}

func TestRunDoctor_OrphanedManagedContent_SelectedAgentIsNeverReported(t *testing.T) {
	home := t.TempDir()
	writeOrphanFixtureFile(t, filepath.Join(home, ".claude", "CLAUDE.md"), "<!-- gentle-ai:persona -->\nbody\n<!-- /gentle-ai:persona -->\n")
	output := runDoctorAgentConfigFixture(t, home, `{"installed_agents":["pi","claude-code"],"selection_configured":true}`)
	if !strings.Contains(output, "  [ok]  sync:orphaned_content          no excluded agent carries leftover gentle-ai managed content\n") {
		t.Fatalf("a selected agent's own managed content must never be reported as an orphan:\n%s", output)
	}
}
