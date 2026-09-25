package main

import (
	"strings"
	"testing"
)

func TestApplyNamespaceSetDefaults(t *testing.T) {
	tests := []struct {
		name      string
		folder    string
		namespace string
		setValues setFlags
		want      []string
	}{
		{
			name:      "bookinfo namespace follows requested namespace",
			folder:    "bookinfo",
			namespace: "bookinfo",
			want:      []string{"namespaces.bookInfo=bookinfo"},
		},
		{
			name:      "otel demo namespace follows requested namespace",
			folder:    "otel-demo",
			namespace: "custom-otel",
			want:      []string{"namespaces.otelDemo=custom-otel"},
		},
		{
			name:      "sock shop namespace follows requested namespace",
			folder:    "sock-shop",
			namespace: "custom-sock",
			want:      []string{"namespaces.sockShop=custom-sock"},
		},
		{
			name:      "explicit chart namespace is preserved",
			folder:    "bookinfo",
			namespace: "bookinfo",
			setValues: setFlags{"namespaces.bookInfo=book-info"},
			want:      []string{"namespaces.bookInfo=book-info"},
		},
		{
			name:      "unknown chart is unchanged",
			folder:    "custom-app",
			namespace: "custom",
			setValues: setFlags{"foo=bar"},
			want:      []string{"foo=bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{FolderName: tc.folder, Namespace: tc.namespace, SetValues: tc.setValues}
			applyNamespaceSetDefaults(config)

			if len(config.SetValues) != len(tc.want) {
				t.Fatalf("SetValues = %#v, want %#v", []string(config.SetValues), tc.want)
			}
			for i := range tc.want {
				if config.SetValues[i] != tc.want[i] {
					t.Fatalf("SetValues = %#v, want %#v", []string(config.SetValues), tc.want)
				}
			}
		})
	}
}

func TestFormatHelmArgsRedactsSensitiveSetValues(t *testing.T) {
	args := []string{
		"upgrade", "agent",
		"--set", "agent.secret.OPENAI_API_KEY=top-secret",
		"--set-string=credentials.access-token=token-value",
		"--set", "agent.config.MODEL_ALIAS=qwen",
	}
	got := formatHelmArgs(args)
	if strings.Contains(got, "top-secret") || strings.Contains(got, "token-value") {
		t.Fatalf("formatHelmArgs leaked a secret: %s", got)
	}
	if !strings.Contains(got, "OPENAI_API_KEY=<redacted>") ||
		!strings.Contains(got, "access-token=<redacted>") ||
		!strings.Contains(got, "MODEL_ALIAS=qwen") {
		t.Fatalf("formatHelmArgs redaction is incorrect: %s", got)
	}
	if args[3] != "agent.secret.OPENAI_API_KEY=top-secret" {
		t.Fatal("formatHelmArgs mutated command arguments")
	}
}
