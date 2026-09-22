package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultChartsPath = "/charts"
	defaultNamespace  = "default"
)

// setFlags implements flag.Value to accumulate multiple --set flags.
// Go's flag package keeps only the last value for a flag; this type
// appends each occurrence so all --set values are preserved.
type setFlags []string

func (s *setFlags) String() string { return strings.Join(*s, ",") }
func (s *setFlags) Set(val string) error {
	*s = append(*s, val)
	return nil
}

type Config struct {
	FolderName      string
	ReleaseName     string
	Namespace       string
	ChartsPath      string
	ValuesFile      string
	SetValues       setFlags // supports multiple --set flags
	DryRun          bool
	Wait            bool
	Timeout         string
	CreateNS        bool
	Upgrade         bool
	Delete          bool
	DeleteNamespace bool
	KubeConfig      string
	KubeContext     string
}

func main() {
	config := parseFlags()

	if config.Delete {
		if err := uninstallApp(config); err != nil {
			log.Fatalf("Uninstall failed: %v", err)
		}
		log.Printf("Successfully uninstalled app chart (release: %s, namespace: %s)", config.ReleaseName, config.Namespace)
		return
	}

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
	flag.Var(&config.SetValues, "set", "Set values on command line (can be repeated: --set key=value --set key2=value2)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Simulate installation without applying")
	flag.BoolVar(&config.Wait, "wait", true, "Wait for resources to be ready")
	flag.StringVar(&config.Timeout, "timeout", "20m", "Timeout for installation")
	flag.BoolVar(&config.CreateNS, "create-namespace", true, "Create namespace if it doesn't exist")
	flag.BoolVar(&config.Upgrade, "upgrade", true, "Use helm upgrade --install for idempotent installs (set to false to use helm install)")
	flag.BoolVar(&config.Delete, "delete", false, "Uninstall (helm uninstall) the chart instead of installing it")
	flag.BoolVar(&config.DeleteNamespace, "delete-namespace", false, "Also delete the target namespace after uninstalling (only used with -delete)")
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
		fmt.Fprintf(os.Stderr, "  install-app -folder sock-shop -dry-run\n\n")
		fmt.Fprintf(os.Stderr, "  # Uninstall a previously installed release\n")
		fmt.Fprintf(os.Stderr, "  install-app -delete -folder sock-shop -namespace sock-shop\n\n")
		fmt.Fprintf(os.Stderr, "  # Uninstall and also remove the namespace\n")
		fmt.Fprintf(os.Stderr, "  install-app -delete -delete-namespace -folder sock-shop -namespace sock-shop\n")
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

func applyNamespaceSetDefaults(config *Config) {
	keyByFolder := map[string]string{
		"bookinfo":  "namespaces.bookInfo",
		"otel-demo": "namespaces.otelDemo",
		"sock-shop": "namespaces.sockShop",
	}

	namespaceKey, ok := keyByFolder[config.FolderName]
	if !ok || config.Namespace == "" || hasHelmSetKey(config.SetValues, namespaceKey) {
		return
	}

	config.SetValues = append(config.SetValues, fmt.Sprintf("%s=%s", namespaceKey, config.Namespace))
}

func hasHelmSetKey(values []string, key string) bool {
	prefix := key + "="
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == key || strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func installChart(config *Config) error {
	chartPath := filepath.Join(config.ChartsPath, config.FolderName)
	applyNamespaceSetDefaults(config)

	// Must run before anything touches a namespace: a broken metrics APIService
	// makes every namespace in the cluster impossible to finalize, so waiting for
	// a Terminating namespace below would otherwise block forever.
	healStaleMetricsAPIService()

	// Calculate total steps upfront for progress display.
	totalSteps := 2 // cleanupStuckRelease + helm run always execute
	if config.CreateNS {
		totalSteps++
	}
	if config.Upgrade {
		totalSteps++
	}
	if config.Wait {
		totalSteps++
	}
	step := 0
	nextStep := func(msg string, args ...interface{}) {
		step++
		log.Printf("[%d/%d] "+msg, append([]interface{}{step, totalSteps}, args...)...)
	}

	// Pre-create namespace if requested, instead of relying on Helm's --create-namespace
	// which fails with "already exists" error on upgrade --install when namespace was
	// created outside of Helm
	if config.CreateNS {
		nextStep("Preparing namespace: %s", config.Namespace)
		if err := ensureNamespace(config.Namespace, config.ReleaseName); err != nil {
			log.Printf("Warning: failed to ensure namespace %s: %v", config.Namespace, err)
		}
	}

	// Clean up any stuck Helm release before attempting install.
	nextStep("Checking for stuck releases")
	if err := cleanupStuckRelease(config.ReleaseName, config.Namespace); err != nil {
		log.Printf("Warning: stuck release cleanup failed: %v", err)
	}

	// Adopt any pre-existing resources so Helm can manage them on upgrade --install.
	// Prevents "invalid ownership metadata" errors when resources were left behind
	// from a previous Helm release purged without deleting the underlying resources.
	if config.Upgrade {
		nextStep("Adopting existing chart resources")
		if err := adoptExistingResources(config); err != nil {
			log.Printf("Warning: failed to adopt existing resources: %v", err)
		}
	}

	// Build helm command
	var args []string

	if config.Upgrade {
		args = append(args, "upgrade", "--install")
	} else {
		args = append(args, "install")
	}

	args = append(args, config.ReleaseName, chartPath)
	args = append(args, "--namespace", config.Namespace)

	// Namespace is pre-created by ensureNamespace(), no need for --create-namespace

	if config.ValuesFile != "" {
		args = append(args, "-f", config.ValuesFile)
	}

	for _, setValue := range config.SetValues {
		args = append(args, "--set", setValue)
	}

	if config.DryRun {
		args = append(args, "--dry-run")
	}

	// NOTE: We intentionally do NOT pass --wait to Helm.
	// Helm v3.14's client-go rate limiter has a known bug that causes
	// "client rate limiter Wait returned an error: context deadline exceeded"
	// when polling pod readiness. Instead, we use kubectl rollout status below.

	if config.Timeout != "" {
		args = append(args, "--timeout", config.Timeout)
	}

	if config.KubeConfig != "" {
		args = append(args, "--kubeconfig", config.KubeConfig)
	}

	if config.KubeContext != "" {
		args = append(args, "--kube-context", config.KubeContext)
	}

	nextStep("Running: helm %s", strings.Join(args, " "))

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// If --wait was requested, use kubectl rollout status instead of Helm's
	// built-in wait which suffers from client-go rate limiter bugs in v3.14
	if config.Wait {
		nextStep("Waiting for deployments to be ready")
		if err := waitForDeployments(config.Namespace, config.ReleaseName, config.Timeout); err != nil {
			return fmt.Errorf("deployments not ready: %w", err)
		}
	}

	return nil
}

// uninstallApp uninstalls the Helm release identified by config, and — when
// config.DeleteNamespace is set — also deletes the target namespace once the
// release is gone. Mirrors install-agent's uninstallChart(), extended with
// namespace deletion since the uninstall-application fault (unlike
// uninstall-agent) requests namespaces:delete in its RBAC and is expected to
// tear the target application's namespace down entirely, not just its release.
func uninstallApp(config *Config) error {
	releaseName := config.ReleaseName
	if releaseName == "" {
		releaseName = config.FolderName
	}

	args := []string{"uninstall", releaseName, "--namespace", config.Namespace, "--ignore-not-found"}
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
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm uninstall failed: %w", err)
	}

	if !config.DeleteNamespace {
		return nil
	}

	// A namespace cannot finalize while any aggregated APIService is unavailable,
	// and this uninstall may well be what broke one. Healing here as well as at
	// install time is what keeps a teardown from leaving the cluster wedged until
	// the *next* install happens to run -- which, between experiments, can be
	// hours, and presents as an unrelated "namespace is Terminating" failure.
	healStaleMetricsAPIService()

	log.Printf("Deleting namespace: %s", config.Namespace)
	deleteArgs := []string{"delete", "namespace", config.Namespace, "--ignore-not-found"}
	if config.Timeout != "" {
		deleteArgs = append(deleteArgs, "--timeout", config.Timeout)
	}
	deleteCmd := exec.Command("kubectl", deleteArgs...)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	if err := deleteCmd.Run(); err != nil {
		return fmt.Errorf("namespace deletion failed: %w", err)
	}

	return nil
}

// waitForDeployments waits for the deployments owned by the given Helm release to be ready.
// It uses `helm get manifest` to discover which Deployments belong to the release, avoiding
// false failures from other releases sharing the namespace (e.g. a leftover agent deployment).
// If releaseName is empty or helm manifest lookup fails, it falls back to all deployments in
// the namespace.
func waitForDeployments(namespace, releaseName, timeout string) error {
	if timeout == "" {
		timeout = "15m"
	}

	var deployments []string

	if releaseName != "" {
		manifestCmd := exec.Command("helm", "get", "manifest", releaseName, "-n", namespace)
		manifestOut, manifestErr := manifestCmd.Output()
		if manifestErr == nil {
			for _, doc := range strings.Split(string(manifestOut), "---") {
				var kind, name, docNS string
				inMetadata := false
				for _, line := range strings.Split(doc, "\n") {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "kind:") && !strings.HasPrefix(line, " ") {
						kind = strings.TrimSpace(strings.TrimPrefix(trimmed, "kind:"))
					}
					if trimmed == "metadata:" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
						inMetadata = true
					} else if inMetadata && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
						inMetadata = false
					}
					if inMetadata {
						if strings.HasPrefix(trimmed, "name:") {
							indent := len(line) - len(strings.TrimLeft(line, " \t"))
							if indent <= 4 {
								name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")), `"'`)
							}
						}
						if strings.HasPrefix(trimmed, "namespace:") {
							docNS = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "namespace:")), `"'`)
						}
					}
				}
				// Only wait for deployments in the target namespace; skip cross-namespace
				// resources (e.g. monitoring components deployed to a different namespace).
				if strings.EqualFold(kind, "Deployment") && name != "" && (docNS == "" || docNS == namespace) {
					deployments = append(deployments, name)
				}
			}
			log.Printf("Scoping wait to %d deployment(s) from helm release %s: %s",
				len(deployments), releaseName, strings.Join(deployments, ", "))
		}
	}

	if len(deployments) == 0 {
		log.Printf("Falling back to listing all deployments in namespace %s", namespace)
		listCmd := exec.Command("kubectl", "get", "deployments", "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
		out, err := listCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}
		deployments = strings.Fields(string(out))
	}

	if len(deployments) == 0 {
		log.Printf("No deployments found, skipping wait")
		return nil
	}

	// Wait for all deployments concurrently so slow Java services don't serialize the wait
	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(deployments))

	for _, dep := range deployments {
		go func(d string) {
			log.Printf("Waiting for deployment %s...", d)
			waitCmd := exec.Command("kubectl", "rollout", "status", "deployment/"+d,
				"-n", namespace, "--timeout="+timeout)
			waitCmd.Stdout = os.Stdout
			waitCmd.Stderr = os.Stderr
			if err := waitCmd.Run(); err != nil {
				results <- result{d, fmt.Errorf("deployment %s not ready: %w", d, err)}
				return
			}
			log.Printf("Deployment %s is ready", d)
			results <- result{d, nil}
		}(dep)
	}

	var errs []string
	for i := 0; i < len(deployments); i++ {
		r := <-results
		if r.err != nil {
			errs = append(errs, r.err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	log.Printf("All deployments in namespace %s are ready", namespace)
	return nil
}

// cleanupStuckRelease checks if a Helm release exists in a broken state
// (pending-install, pending-upgrade, pending-rollback, or failed) and
// uninstalls it so that the next "helm upgrade --install" can succeed.
func cleanupStuckRelease(releaseName, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get release status via helm
	statusCmd := exec.CommandContext(ctx, "helm", "status", releaseName,
		"-n", namespace, "-o", "json")
	out, err := statusCmd.Output()
	if err != nil {
		// Release doesn't exist — nothing to clean up
		return nil
	}

	// Parse JSON to extract only info.status — avoid false positives from
	// strings.Contains on the full output (which includes the manifest YAML
	// where words like "failed" appear in probe descriptions, label values, etc.)
	var helmStatus struct {
		Info struct {
			Status string `json:"status"`
		} `json:"info"`
	}
	releaseStatus := ""
	if jsonErr := json.Unmarshal(out, &helmStatus); jsonErr == nil {
		releaseStatus = helmStatus.Info.Status
	} else {
		// Fallback: find the status value using a narrow pattern rather than
		// scanning the entire JSON blob.
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, `"status"`) {
				for _, candidate := range []string{"pending-install", "pending-upgrade", "pending-rollback", "failed", "deployed", "superseded", "uninstalled"} {
					if strings.Contains(line, candidate) {
						releaseStatus = candidate
						break
					}
				}
				if releaseStatus != "" {
					break
				}
			}
		}
	}
	stuckStates := []string{"pending-install", "pending-upgrade", "pending-rollback", "failed"}
	isStuck := false
	for _, state := range stuckStates {
		if releaseStatus == state {
			isStuck = true
			log.Printf("Release %s is stuck in '%s' state, cleaning up...", releaseName, state)
			break
		}
	}

	if !isStuck {
		return nil
	}

	log.Printf("Uninstalling stuck release %s in namespace %s", releaseName, namespace)
	uninstallCmd := exec.CommandContext(ctx, "helm", "uninstall", releaseName,
		"-n", namespace, "--no-hooks")
	uninstallCmd.Stdout = os.Stdout
	uninstallCmd.Stderr = os.Stderr
	if err := uninstallCmd.Run(); err != nil {
		// helm uninstall can fail when the release secret is corrupted or partially
		// deleted. Fall back to wiping the Helm state secrets directly so that
		// the subsequent "helm upgrade --install" sees no existing release and
		// performs a clean install instead of failing with
		// "'<release>' has no deployed releases".
		log.Printf("helm uninstall failed (%v), falling back to deleting Helm state secrets", err)
		if secretErr := deleteHelmStateSecrets(ctx, releaseName, namespace); secretErr != nil {
			return fmt.Errorf("failed to uninstall stuck release %s: helm uninstall: %w; secret delete: %v", releaseName, err, secretErr)
		}
		log.Printf("Successfully removed Helm state secrets for release %s", releaseName)
		return nil
	}

	log.Printf("Successfully cleaned up stuck release %s", releaseName)
	return nil
}

// deleteHelmStateSecrets removes all Helm release secrets for a given release name,
// which is the fallback when "helm uninstall" itself fails on a corrupted release.
func deleteHelmStateSecrets(ctx context.Context, releaseName, namespace string) error {
	listCmd := exec.CommandContext(ctx, "kubectl", "get", "secret",
		"-n", namespace,
		"-l", fmt.Sprintf("name=%s,owner=helm", releaseName),
		"-o", "jsonpath={.items[*].metadata.name}")
	out, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list Helm state secrets: %w", err)
	}

	names := strings.Fields(string(out))
	if len(names) == 0 {
		log.Printf("No Helm state secrets found for release %s in namespace %s", releaseName, namespace)
		return nil
	}

	log.Printf("Deleting Helm state secrets: %s", strings.Join(names, ", "))
	deleteArgs := append([]string{"delete", "secret", "-n", namespace}, names...)
	deleteCmd := exec.CommandContext(ctx, "kubectl", deleteArgs...)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	return deleteCmd.Run()
}

// waitForNamespaceNotTerminating polls until the namespace is either absent or
// Active (not Terminating). Needed because namespace deletion in Kubernetes is
// async: the previous experiment's cleanup step runs helm uninstall / deletes
// resources, but the namespace lingers in Terminating state for tens of seconds
// while the API server finalizes resource removal. If the next experiment's
// install-application runs during that window, Helm fails with "unable to
// create new content in namespace X because it is being terminated".
func waitForNamespaceNotTerminating(namespace string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := exec.Command("kubectl", "get", "ns", namespace,
			"-o", "jsonpath={.status.phase}").Output()
		if err != nil {
			return nil // namespace gone — safe to proceed
		}
		if strings.TrimSpace(string(out)) != "Terminating" {
			return nil
		}
		log.Printf("Namespace %s is Terminating, waiting...", namespace)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("namespace %s still Terminating after %v", namespace, timeout)
}

// healStaleMetricsAPIService removes the aggregated metrics APIService when it
// exists but reports Available=False, so the cluster can recover on its own.
//
// Why this is needed. v1beta1.metrics.k8s.io is cluster-scoped and owned by this
// chart, but the metrics-server backing it runs in the monitoring namespace. A
// partial teardown — typically the app namespace being deleted out from under
// `helm uninstall` — can strip metrics-server's ClusterRoleBindings while leaving
// the Deployment running. It then 403s on nodes/subjectaccessreviews, never turns
// Ready, and its Service keeps no endpoints.
//
// The damage is cluster-wide and badly disguised: with an aggregated APIService
// unavailable, the apiserver cannot complete discovery, and the namespace
// controller refuses to finalize ANY terminating namespace. Every later
// experiment then dies in install with a "namespace is Terminating" error that
// points nowhere near metrics-server, and no amount of waiting clears it.
//
// Deleting a broken one is safe: Helm recreates it, with fresh RBAC, as part of
// installing this chart. Best-effort throughout — never block an install on it.
func healStaleMetricsAPIService() {
	const apiService = "v1beta1.metrics.k8s.io"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "kubectl", "get", "apiservice", apiService,
		"-o", `jsonpath={.status.conditions[?(@.type=="Available")].status}`).Output()
	if err != nil {
		// Not present (the normal first-install case) or unreadable — nothing to do.
		return
	}
	if !strings.EqualFold(strings.TrimSpace(string(out)), "False") {
		return
	}

	// Read the backing Service's namespace before the APIService is deleted —
	// that is the only in-cluster record of where metrics-server actually runs.
	msNamespace := ""
	if nsOut, nsErr := exec.CommandContext(ctx, "kubectl", "get", "apiservice", apiService,
		"-o", "jsonpath={.spec.service.namespace}").Output(); nsErr == nil {
		msNamespace = strings.TrimSpace(string(nsOut))
	}

	log.Printf("APIService %s is present but Available=False; deleting it so namespace "+
		"finalization is not blocked cluster-wide (Helm will recreate it)", apiService)
	del := exec.CommandContext(ctx, "kubectl", "delete", "apiservice", apiService, "--ignore-not-found")
	del.Stdout = os.Stdout
	del.Stderr = os.Stderr
	if err := del.Run(); err != nil {
		log.Printf("Warning: could not delete stale APIService %s: %v", apiService, err)
	}

	if msNamespace == "" {
		return
	}

	// Deleting the APIService unblocks namespace finalization, but metrics-server
	// itself stays broken: its ClusterRole/ClusterRoleBinding were torn down with
	// the previous release, so the running pod keeps 403ing on nodes/metrics and
	// never becomes Ready. Helm recreates the RBAC as part of this install, and
	// restarting the Deployment makes the pod pick it up now instead of after an
	// indefinite client-go backoff — otherwise the APIService comes back
	// Available=False and the whole deadlock returns on the next teardown.
	restart := exec.CommandContext(ctx, "kubectl", "rollout", "restart",
		"deployment/metrics-server", "-n", msNamespace)
	restart.Stdout = os.Stdout
	restart.Stderr = os.Stderr
	if err := restart.Run(); err != nil {
		log.Printf("Warning: could not restart metrics-server in %s: %v", msNamespace, err)
	}
}

func ensureNamespace(namespace, releaseName string) error {
	log.Printf("Ensuring namespace: %s", namespace)

	// Wait out a namespace still terminating from a previous experiment's
	// teardown before trying to create it. This has to happen *before* the
	// create, not after an AlreadyExists error: once the namespace finishes
	// terminating it is gone, so the create has to be (re)attempted afterwards.
	// Previously the wait ran on the AlreadyExists branch and then fell straight
	// through to labelling a namespace that no longer existed, leaving Helm --
	// which is deliberately not given --create-namespace -- to fail with
	// "namespaces not found".
	if err := waitForNamespaceNotTerminating(namespace, 3*time.Minute); err != nil {
		return err
	}

	// Every kubectl call below gets its own deadline. A single context covering
	// the whole function expired during the wait above, so labelling and
	// annotating then failed instantly with "context deadline exceeded".
	kubectl := func(args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		var stderr bytes.Buffer
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil && stderr.Len() > 0 && !strings.Contains(stderr.String(), "AlreadyExists") {
			os.Stderr.Write(stderr.Bytes())
		}
		return stderr.String(), err
	}

	// Create the namespace; if it already exists that is fine. Immediately after
	// a namespace finishes terminating the API server can still reject the
	// create, so retry briefly.
	var createErr error
	for attempt := 0; attempt < 6; attempt++ {
		var stderr string
		stderr, createErr = kubectl("create", "namespace", namespace)
		if createErr == nil {
			break
		}
		if strings.Contains(stderr, "AlreadyExists") {
			log.Printf("Namespace %s already exists", namespace)
			createErr = nil
			break
		}
		log.Printf("Namespace %s not creatable yet (%s); retrying...", namespace, strings.TrimSpace(stderr))
		time.Sleep(5 * time.Second)
	}
	if createErr != nil {
		return fmt.Errorf("failed to create namespace: %w", createErr)
	}

	// Add Helm ownership labels and annotations so Helm can adopt the namespace
	log.Printf("Labeling namespace %s for Helm ownership", namespace)
	if _, err := kubectl("label", "namespace", namespace,
		"app.kubernetes.io/managed-by=Helm", "--overwrite"); err != nil {
		return fmt.Errorf("failed to label namespace: %w", err)
	}

	if _, err := kubectl("annotate", "namespace", namespace,
		fmt.Sprintf("meta.helm.sh/release-name=%s", releaseName),
		fmt.Sprintf("meta.helm.sh/release-namespace=%s", namespace),
		"--overwrite"); err != nil {
		return fmt.Errorf("failed to annotate namespace: %w", err)
	}

	return nil
}

// adoptExistingResources uses `helm template` to discover all resources the chart will create,
// then labels/annotates any that already exist in the cluster without Helm ownership metadata.
// This prevents "invalid ownership metadata" errors on upgrade --install when resources were
// left behind after a previous release was purged without deleting the K8s resources.
func adoptExistingResources(config *Config) error {
	chartPath := filepath.Join(config.ChartsPath, config.FolderName)

	args := []string{"template", config.ReleaseName, chartPath, "--namespace", config.Namespace}
	if config.ValuesFile != "" {
		args = append(args, "-f", config.ValuesFile)
	}
	for _, setValue := range config.SetValues {
		args = append(args, "--set", setValue)
	}

	log.Printf("Discovering chart resources via: helm %s", strings.Join(args, " "))
	cmd := exec.Command("helm", args...)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("helm template failed: %w", err)
	}

	resources := parseHelmTemplateOutput(string(out))
	if len(resources) == 0 {
		log.Printf("No resources discovered from chart template")
		return nil
	}

	total := len(resources)
	log.Printf("Discovered %d resources from chart template", total)
	adopted := 0
	for i, res := range resources {
		if adoptResource(res, config.ReleaseName, config.Namespace) {
			adopted++
		}
		if n := i + 1; n%10 == 0 || n == total {
			log.Printf("  [%d/%d] Resources checked", n, total)
		}
	}
	if adopted > 0 {
		log.Printf("Adopted %d pre-existing resources for Helm release %s", adopted, config.ReleaseName)
	}
	return nil
}

// k8sResource represents a Kubernetes resource extracted from helm template output.
type k8sResource struct {
	Kind      string
	Name      string
	Namespace string
}

// parseHelmTemplateOutput parses multi-document YAML from `helm template` and
// extracts the kind, name, and namespace of each resource.
func parseHelmTemplateOutput(output string) []k8sResource {
	docs := strings.Split(output, "---")
	var resources []k8sResource

	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var kind, name, namespace string
		inMetadata := false

		for _, line := range strings.Split(doc, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.HasPrefix(trimmed, "kind:") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				kind = strings.TrimSpace(strings.TrimPrefix(trimmed, "kind:"))
				continue
			}
			if trimmed == "metadata:" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inMetadata = true
				continue
			}
			if inMetadata && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inMetadata = false
			}
			if inMetadata {
				if strings.HasPrefix(trimmed, "name:") {
					indent := len(line) - len(strings.TrimLeft(line, " \t"))
					if indent <= 4 {
						name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")), "\"'")
					}
				}
				if strings.HasPrefix(trimmed, "namespace:") {
					namespace = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "namespace:")), "\"'")
				}
			}
		}
		if kind != "" && name != "" {
			resources = append(resources, k8sResource{Kind: kind, Name: name, Namespace: namespace})
		}
	}
	return resources
}

// adoptResource labels/annotates a pre-existing K8s resource with Helm ownership metadata.
// Returns true if the resource existed and was adopted.
func adoptResource(res k8sResource, releaseName, releaseNamespace string) bool {
	resourceType := strings.ToLower(res.Kind)
	ns := res.Namespace
	if ns == "" {
		ns = releaseNamespace
	}

	getArgs := []string{"get", resourceType, res.Name, "-n", ns, "--no-headers", "--ignore-not-found"}
	out, err := exec.Command("kubectl", getArgs...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return false
	}

	log.Printf("Adopting existing %s/%s (ns=%s) for Helm release %s", res.Kind, res.Name, ns, releaseName)

	labelCmd := exec.Command("kubectl", "label", resourceType, res.Name, "-n", ns,
		"app.kubernetes.io/managed-by=Helm", "--overwrite")
	labelCmd.Stdout = os.Stdout
	labelCmd.Stderr = os.Stderr
	if err := labelCmd.Run(); err != nil {
		log.Printf("Warning: failed to label %s/%s: %v", res.Kind, res.Name, err)
	}

	annotateCmd := exec.Command("kubectl", "annotate", resourceType, res.Name, "-n", ns,
		fmt.Sprintf("meta.helm.sh/release-name=%s", releaseName),
		fmt.Sprintf("meta.helm.sh/release-namespace=%s", releaseNamespace),
		"--overwrite")
	annotateCmd.Stdout = os.Stdout
	annotateCmd.Stderr = os.Stderr
	if err := annotateCmd.Run(); err != nil {
		log.Printf("Warning: failed to annotate %s/%s: %v", res.Kind, res.Name, err)
	}
	return true
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
