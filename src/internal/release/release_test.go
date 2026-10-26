package release

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNativeReleaseMatchesPackage(t *testing.T) {
	data, err := os.ReadFile("../../../VERSION.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if Version != manifest.Version {
		t.Fatalf("native %s does not match package %s", Version, manifest.Version)
	}
	if Changes == "" {
		t.Fatal("release notes cannot be empty")
	}
}
