package luaskills

import (
	"encoding/json"
	"testing"
)

// TestSkillInstallSourceTypeMatchesRustProtocol verifies install source JSON protocol values.
// TestSkillInstallSourceTypeMatchesRustProtocol 校验安装来源 JSON 协议值。
func TestSkillInstallSourceTypeMatchesRustProtocol(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SkillInstallSourceType
		expected   string
	}{
		{
			name:       "github",
			sourceType: SkillInstallSourceGithub,
			expected:   "github",
		},
		{
			name:       "official hub",
			sourceType: SkillInstallSourceOfficialHub,
			expected:   "official_hub",
		},
		{
			name:       "url",
			sourceType: SkillInstallSourceURL,
			expected:   "url",
		},
		{
			name:       "private url manifest",
			sourceType: SkillInstallSourcePrivateURLManifest,
			expected:   "private_url_manifest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := SkillInstallRequest{SourceType: test.sourceType}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("marshal skill install request: %v", err)
			}
			var payload map[string]string
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatalf("unmarshal skill install request: %v", err)
			}
			if payload["source_type"] != test.expected {
				t.Fatalf("unexpected source_type: %#v", payload["source_type"])
			}
		})
	}
}
