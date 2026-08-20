package version

import "testing"

func TestCurrentIncludesJSONEnvelope(t *testing.T) {
	info := Current()
	if info.SchemaVersion != SchemaVersion || info.Command != "version" || !info.Success || info.State != "current" {
		t.Fatalf("invalid version envelope: %+v", info)
	}
}
