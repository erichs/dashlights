package install

import (
	"strings"
	"testing"
)

func TestGetShellTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		wantContains string
		wantEmpty    bool
	}{
		{
			name:         "bash template",
			templateType: TemplateBash,
			wantContains: "PS1=",
		},
		{
			name:         "zsh template",
			templateType: TemplateZsh,
			wantContains: "PROMPT=",
		},
		{
			name:         "fish template",
			templateType: TemplateFish,
			wantContains: "--on-event fish_prompt",
		},
		{
			name:         "p10k template",
			templateType: TemplateP10k,
			wantContains: "p10k segment",
		},
		{
			name:         "unknown template",
			templateType: TemplateType("unknown"),
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetShellTemplate(tt.templateType)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GetShellTemplate(%q) = %q, want empty", tt.templateType, got)
				}
				return
			}

			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("GetShellTemplate(%q) does not contain %q", tt.templateType, tt.wantContains)
			}

			// All templates should have sentinel markers
			if !strings.Contains(got, SentinelBegin) {
				t.Errorf("GetShellTemplate(%q) missing SentinelBegin", tt.templateType)
			}
			if !strings.Contains(got, SentinelEnd) {
				t.Errorf("GetShellTemplate(%q) missing SentinelEnd", tt.templateType)
			}
		})
	}
}

func TestSentinelMarkers(t *testing.T) {
	// Verify sentinel markers are correct
	if SentinelBegin != "# BEGIN dashlights" {
		t.Errorf("SentinelBegin = %q, want %q", SentinelBegin, "# BEGIN dashlights")
	}
	if SentinelEnd != "# END dashlights" {
		t.Errorf("SentinelEnd = %q, want %q", SentinelEnd, "# END dashlights")
	}
}

func TestTemplateConstants(t *testing.T) {
	// Verify all templates contain the dashlights command
	templates := []struct {
		name     string
		template string
	}{
		{"BashTemplate", BashTemplate},
		{"ZshTemplate", ZshTemplate},
		{"FishTemplate", FishTemplate},
		{"P10kTemplate", P10kTemplate},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.template, "dashlights") {
				t.Errorf("%s does not contain 'dashlights'", tt.name)
			}
			if !strings.Contains(tt.template, SentinelBegin) {
				t.Errorf("%s does not contain SentinelBegin", tt.name)
			}
			if !strings.Contains(tt.template, SentinelEnd) {
				t.Errorf("%s does not contain SentinelEnd", tt.name)
			}
		})
	}
}

func TestDashlightsCommand(t *testing.T) {
	if DashlightsCommand != "dashlights --agentic" {
		t.Errorf("DashlightsCommand = %q, want %q", DashlightsCommand, "dashlights --agentic")
	}
}
