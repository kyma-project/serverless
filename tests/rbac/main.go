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

type RoleConfig struct {
	Rules     []Rule `json:"rules"`
	Forbidden []Rule `json:"forbidden"`
}

type Config struct {
	Admin RoleConfig `json:"admin"`
	Edit  RoleConfig `json:"edit"`
	View  RoleConfig `json:"view"`
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
	// Rules merge upward; forbidden rules are per-role only.
	viewRules := cfg.View.Rules
	editRules := append(cfg.Edit.Rules, viewRules...)
	adminRules := append(cfg.Admin.Rules, editRules...)

	var failed bool
	for _, c := range []struct {
		name      string
		rules     []Rule
		forbidden []Rule
	}{
		{"admin", adminRules, cfg.Admin.Forbidden},
		{"edit", editRules, cfg.Edit.Forbidden},
		{"view", viewRules, cfg.View.Forbidden},
	} {
		if err := verifyClusterRole(clientset, c.name, c.rules, c.forbidden); err != nil {
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

// permission is a single (apiGroup, resource, verb) entry.
type permission struct{ group, resource, verb string }

func (p permission) String() string {
	return p.group + "/" + p.resource + "/" + p.verb
}

// expand flattens Rules into individual permissions.
func expand(rules []Rule) map[permission]bool {
	set := make(map[permission]bool)
	for _, r := range rules {
		for _, g := range r.APIGroups {
			for _, res := range r.Resources {
				for _, v := range r.Verbs {
					set[permission{g, res, v}] = true
				}
			}
		}
	}
	return set
}

func verifyClusterRole(clientset *kubernetes.Clientset, name string, expected []Rule, forbidden []Rule) error {
	cr, err := clientset.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get clusterrole %q: %w", name, err)
	}

	expectedSet := expand(expected)
	forbiddenSet := expand(forbidden)

	// Build actual permission set from the ClusterRole.
	actualSet := make(map[permission]bool)
	for _, r := range cr.Rules {
		for _, g := range r.APIGroups {
			for _, res := range r.Resources {
				for _, v := range r.Verbs {
					actualSet[permission{g, res, v}] = true
				}
			}
		}
	}

	var errs []string
	for p := range expectedSet {
		if !actualSet[p] {
			errs = append(errs, "missing expected: "+p.String())
		}
	}
	for p := range forbiddenSet {
		if actualSet[p] {
			errs = append(errs, "forbidden found: "+p.String())
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("rules mismatch:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
