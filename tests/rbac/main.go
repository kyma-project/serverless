package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type Rule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

type Config struct {
	Admin []Rule `json:"admin"`
	Edit  []Rule `json:"edit"`
	View  []Rule `json:"view"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.yaml>\n", os.Args[0])
		os.Exit(1)
	}

	cfg, err := loadConfig(os.Args[1])
	if err != nil {
		fatal("loading config: %v", err)
	}

	clientset, err := buildClientset()
	if err != nil {
		fatal("building kubernetes client: %v", err)
	}

	// Kubernetes aggregation hierarchy: admin ⊃ edit ⊃ view
	view := cfg.View
	edit := append(cfg.Edit, view...)
	admin := append(cfg.Admin, edit...)

	var failed bool
	for _, c := range []struct {
		name  string
		rules []Rule
	}{
		{"admin", admin},
		{"edit", edit},
		{"view", view},
	} {
		if err := verifyClusterRole(clientset, c.name, c.rules); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: clusterrole/%s: %v\n", c.name, err)
			failed = true
		} else {
			fmt.Printf("PASS: clusterrole/%s\n", c.name)
		}
	}

	if failed {
		os.Exit(1)
	}
	fmt.Println("ALL PASSED")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(data, &cfg)
}

func buildClientset() (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func keyFor(apiGroups, resources, verbs []string) string {
	sorted := func(s []string) string {
		c := make([]string, len(s))
		copy(c, s)
		sort.Strings(c)
		return strings.Join(c, ",")
	}
	return sorted(apiGroups) + "/" + sorted(resources) + "/" + sorted(verbs)
}

func verifyClusterRole(clientset *kubernetes.Clientset, name string, expected []Rule) error {
	cr, err := clientset.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get clusterrole %q: %w", name, err)
	}

	// Collect relevant API groups from expected rules.
	groups := map[string]bool{}
	expectedSet := map[string]bool{}
	for _, r := range expected {
		for _, g := range r.APIGroups {
			groups[g] = true
		}
		expectedSet[keyFor(r.APIGroups, r.Resources, r.Verbs)] = true
	}

	// Filter actual rules to only relevant API groups.
	actualSet := map[string]bool{}
	for _, r := range cr.Rules {
		for _, g := range r.APIGroups {
			if groups[g] {
				actualSet[keyFor(r.APIGroups, r.Resources, r.Verbs)] = true
				break
			}
		}
	}

	var errs []string
	for k := range expectedSet {
		if !actualSet[k] {
			errs = append(errs, "missing: "+k)
		}
	}
	for k := range actualSet {
		if !expectedSet[k] {
			errs = append(errs, "unexpected: "+k)
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("rules mismatch:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
