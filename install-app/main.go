package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultChartsPath = "/charts"
	defaultNamespace  = "default"
)

type Config struct {
	FolderName    string
	ReleaseName   string
	Namespace     string
	ChartsPath    string
	ValuesFile    string
	SetValues     string
	DryRun        bool
	Wait          bool
	Timeout       string
	CreateNS      bool
	Upgrade       bool
	KubeConfig    string
	KubeContext   string
}

func main() {
	config := parseFlags()

	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if err := installChart(config); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}

	log.Printf("Successfully installed chart from folder: %s", config.FolderName)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.FolderName, "folder", "", "Name of the folder containing Helm chart (required)")
	flag.StringVar(&config.ReleaseName, "release", "", "Helm release name (defaults to folder name)")
	flag.StringVar(&config.Namespace, "namespace", defaultNamespace, "Kubernetes namespace to install into")
	flag.StringVar(&config.ChartsPath, "charts-path", defaultChartsPath, "Base path where charts are located")
	flag.StringVar(&config.ValuesFile, "values", "", "Path to custom values file")
	flag.StringVar(&config.SetValues, "set", "", "Set values on command line (key=value,key2=value2)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Simulate installation without applying")
	flag.BoolVar(&config.Wait, "wait", true, "Wait for resources to be ready")
	flag.StringVar(&config.Timeout, "timeout", "5m", "Timeout for installation")
	flag.BoolVar(&config.CreateNS, "create-namespace", true, "Create namespace if it doesn't exist")
	flag.BoolVar(&config.Upgrade, "upgrade", true, "Use helm upgrade --install for idempotent installs (set to false to use helm install)")
	flag.StringVar(&config.KubeConfig, "kubeconfig", "", "Path to kubeconfig file")
	flag.StringVar(&config.KubeContext, "context", "", "Kubernetes context to use")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: install-app [options]\n\n")
		fmt.Fprintf(os.Stderr, "A tool to install Helm charts from the packaged repository.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Install sock-shop chart into sock-shop namespace\n")
		fmt.Fprintf(os.Stderr, "  install-app -folder sock-shop -namespace sock-shop\n\n")
		fmt.Fprintf(os.Stderr, "  # Install with custom values file\n")
		fmt.Fprintf(os.Stderr, "  install-app -folder sock-shop -values /custom/values.yaml\n\n")
		fmt.Fprintf(os.Stderr, "  # Upgrade existing release\n")
		fmt.Fprintf(os.Stderr, "  install-app -folder sock-shop -upgrade -namespace sock-shop\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry-run installation\n")
		fmt.Fprintf(os.Stderr, "  install-app -folder sock-shop -dry-run\n")
	}

	flag.Parse()

	// Default release name to folder name if not specified
	if config.ReleaseName == "" {
		config.ReleaseName = config.FolderName
	}

	return config
}

func validateConfig(config *Config) error {
	if config.FolderName == "" {
		return fmt.Errorf("folder name is required. Use -folder flag")
	}

	chartPath := filepath.Join(config.ChartsPath, config.FolderName)
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		return fmt.Errorf("chart folder not found: %s", chartPath)
	}

	// Check for Chart.yaml to verify it's a valid Helm chart
	chartYaml := filepath.Join(chartPath, "Chart.yaml")
	if _, err := os.Stat(chartYaml); os.IsNotExist(err) {
		return fmt.Errorf("not a valid Helm chart - Chart.yaml not found in: %s", chartPath)
	}

	// Validate values file if specified
	if config.ValuesFile != "" {
		if _, err := os.Stat(config.ValuesFile); os.IsNotExist(err) {
			return fmt.Errorf("values file not found: %s", config.ValuesFile)
		}
	}

	return nil
}

func installChart(config *Config) error {
	chartPath := filepath.Join(config.ChartsPath, config.FolderName)

	// Build helm command
	var args []string

	if config.Upgrade {
		args = append(args, "upgrade", "--install")
	} else {
		args = append(args, "install")
	}

	args = append(args, config.ReleaseName, chartPath)
	args = append(args, "--namespace", config.Namespace)

	if config.CreateNS {
		args = append(args, "--create-namespace")
	}

	if config.ValuesFile != "" {
		args = append(args, "-f", config.ValuesFile)
	}

	if config.SetValues != "" {
		// Parse comma-separated key=value pairs
		for _, setValue := range strings.Split(config.SetValues, ",") {
			args = append(args, "--set", setValue)
		}
	}

	if config.DryRun {
		args = append(args, "--dry-run")
	}

	if config.Wait {
		args = append(args, "--wait")
	}

	if config.Timeout != "" {
		args = append(args, "--timeout", config.Timeout)
	}

	if config.KubeConfig != "" {
		args = append(args, "--kubeconfig", config.KubeConfig)
	}

	if config.KubeContext != "" {
		args = append(args, "--kube-context", config.KubeContext)
	}

	log.Printf("Executing: helm %s", strings.Join(args, " "))

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ListAvailableCharts lists all available charts in the charts path
func ListAvailableCharts(chartsPath string) ([]string, error) {
	var charts []string

	entries, err := os.ReadDir(chartsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			chartYaml := filepath.Join(chartsPath, entry.Name(), "Chart.yaml")
			if _, err := os.Stat(chartYaml); err == nil {
				charts = append(charts, entry.Name())
			}
		}
	}

	return charts, nil
}
