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

func TestParseUnavailableAPIServices(t *testing.T) {
	out := "v1.apps\tTrue\nv1beta1.metrics.k8s.io\tFalse\nv1.custom.example.io\t\n\n"
	got := parseUnavailableAPIServices(out)
	if len(got) != 1 || got[0] != "v1beta1.metrics.k8s.io" {
		t.Fatalf("parseUnavailableAPIServices() = %v, want [v1beta1.metrics.k8s.io]", got)
	}
}

func TestReleaseWorkloads(t *testing.T) {
	list := []byte(`{"items":[
	  {"kind":"Deployment","metadata":{"name":"front-end","annotations":{"meta.helm.sh/release-name":"sock-shop"}}},
	  {"kind":"StatefulSet","metadata":{"name":"carts-db","annotations":{"meta.helm.sh/release-name":"sock-shop"}}},
	  {"kind":"Deployment","metadata":{"name":"flash-agent","annotations":{"meta.helm.sh/release-name":"flash-agent"}}},
	  {"kind":"DaemonSet","metadata":{"name":"node-exporter"}}
	]}`)
	got, err := releaseWorkloads(list, "sock-shop")
	if err != nil {
		t.Fatalf("releaseWorkloads() error = %v", err)
	}
	want := []string{"deployment/front-end", "statefulset/carts-db"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("releaseWorkloads() = %v, want %v", got, want)
	}

	if _, err := releaseWorkloads([]byte("not json"), "sock-shop"); err == nil {
		t.Fatal("releaseWorkloads() accepted malformed JSON")
	}
}
