package main

import (
	"os"
	"reflect"
	"testing"
)

func TestWriteValuesJSONWritesTheDocumentForHelm(t *testing.T) {
	doc := `{"bookInfo":{"productpage":{"replicas":3}},"note":"a, b: c"}`
	path, err := writeValuesJSON(doc)
	if err != nil {
		t.Fatalf("writeValuesJSON() error = %v", err)
	}
	defer os.Remove(path)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Written verbatim: commas and colons need no --set escaping in a file.
	if string(got) != doc {
		t.Fatalf("file holds %q, want %q", got, doc)
	}
}

func TestWriteValuesJSONRejectsAnythingButAnObject(t *testing.T) {
	for _, doc := range []string{`[]`, `null`, `"x"`, `{"agent":`, `not json`} {
		if path, err := writeValuesJSON(doc); err == nil {
			os.Remove(path)
			t.Errorf("writeValuesJSON(%q) succeeded, want an error", doc)
		}
	}
}

func TestValuesJSONKeysListsLeafPaths(t *testing.T) {
	values := map[string]interface{}{
		"bookInfo": map[string]interface{}{"productpage": map[string]interface{}{"replicas": 3.0}, "loadGenerator": map[string]interface{}{"enabled": false}},
		"global":   "x",
	}
	want := []string{"bookInfo.loadGenerator.enabled", "bookInfo.productpage.replicas", "global"}
	if got := valuesJSONKeys(values); !reflect.DeepEqual(got, want) {
		t.Fatalf("valuesJSONKeys() = %v, want %v", got, want)
	}
}

func TestHelmValuesArgsAppliesUserSettingsBeforeInstallerValues(t *testing.T) {
	config := &Config{
		ValuesFile:     "/custom/values.yaml",
		ValuesJSONFile: "/tmp/values-json-1.json",
		SetValues:      setFlags{"namespaces.bookInfo=book-info"},
	}
	want := []string{
		"-f", "/custom/values.yaml",
		"-f", "/tmp/values-json-1.json",
		"--set", "namespaces.bookInfo=book-info",
	}
	if got := helmValuesArgs(config); !reflect.DeepEqual(got, want) {
		t.Fatalf("helmValuesArgs() = %v, want %v", got, want)
	}
	if got := helmValuesArgs(&Config{}); len(got) != 0 {
		t.Fatalf("helmValuesArgs() with no values = %v, want none", got)
	}
}
