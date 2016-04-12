package env

import (
	"os"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	// AuthBearerToken is a global token required for all requests except GET/POST command/
	AuthBearerToken string
	// SlackToken contains the expected Slack custom command token.
	SlackToken string

	// ServerHost is passed to gin and configures the listen address of the server
	ServerHost string

	// K8sServiceHost is the Kubernetes host
	K8sServiceHost string

	// K8sServicePort is the Kubernetes port
	K8sServicePort string

	// K8sNamespace is the namespace used by Broadway's deployments
	K8sNamespace string

	// EtcdHost is the Etcd host
	EtcdHost string

	// EtcdPath is the root directory for Broadway objects
	EtcdPath string

	// SlackWebhook is the Slack incoming webhook
	SlackWebhook string

	// ManifestsPath is the absolute path where manifest files are read from
	ManifestsPath string

	// PlaybooksPath is the absolute path where playbook files are read from
	PlaybooksPath string
)

var (
	cwd                  string
	defaultManifestsPath string
	defaultPlaybooksPath string
)

// LoadEnvs sets env.* variables to their OS-provided value
func LoadEnvs() {
	ManifestsPath = loadw("BROADWAY_MANIFESTS_PATH")
	if ManifestsPath == "" {
		ManifestsPath = defaultManifestsPath
	}
	PlaybooksPath = loadw("BROADWAY_PLAYBOOKS_PATH")
	if PlaybooksPath == "" {
		PlaybooksPath = defaultPlaybooksPath
	}
	AuthBearerToken = loadw("BROADWAY_AUTH_TOKEN")

	ServerHost = loadw("HOST")

	SlackWebhook = loadw("SLACK_WEBHOOK")
	SlackToken = loadw("SLACK_VERIFICATION_TOKEN")

	K8sServiceHost = loadw("KUBERNETES_SERVICE_HOST")
	K8sServicePort = loadw("KUBERNETES_PORT_443_TCP_PORT")
	K8sNamespace = loadf("KUBERNETES_NAMESPACE")

	EtcdHost = loadw("ETCD_HOST")
	EtcdPath = loadw("ETCD_PATH")
}

func loadw(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		glog.Warningf("Environment variable %s unset; defaulting to empty string\n", key)
	}
	return val
}

func loadf(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		glog.Fatalf("Environment variable %s unset; Exiting.n", key)
	}
	return val
}

func init() {
	if d, err := os.Getwd(); err == nil {
		cwd = d
	}
	// Caution, when runnning tests, cwd will be a subfolder, not namely/broadway
	defaultManifestsPath = filepath.Join(cwd, "manifests")
	defaultPlaybooksPath = filepath.Join(cwd, "playbooks")
	LoadEnvs()
}
