// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package client

import (
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/option"
)

const OptUserAgent = "user-agent"

type Config struct {
	ClientParams
	SharedConfig
}

type SharedConfig struct {
	// EnableK8s is a flag that, when set to false, forcibly disables the clientset, to let cilium
	// operates with CNI-compatible orchestrators other than Kubernetes. Default to true.
	EnableK8s bool

	// K8sAPIServerURLs is the list of API server instances
	K8sAPIServerURLs []string

	// K8sKubeConfigPath is the absolute path of the kubernetes kubeconfig file
	K8sKubeConfigPath string

	// K8sClientConnectionTimeout configures the timeout for K8s client connections.
	K8sClientConnectionTimeout time.Duration

	// K8sClientConnectionKeepAlive configures the keep alive duration for K8s client connections.
	K8sClientConnectionKeepAlive time.Duration

	// K8sHeartbeatTimeout configures the timeout for apiserver heartbeat
	K8sHeartbeatTimeout time.Duration

	// K8sClientConnectionRetryTimeout configures how long the agent keeps
	// retrying to establish its initial connection to the API server at
	// startup (cold start) before giving up and exiting. Increase to avoid
	// terminating too soon during a control-plane / API server outage. The
	// effective wait is min(this, --hive-start-timeout), so raise
	// --hive-start-timeout as well for values above its default.
	K8sClientConnectionRetryTimeout time.Duration

	// EnableAPIDiscovery enables Kubernetes API discovery
	EnableK8sAPIDiscovery bool
}

type ClientParams struct {
	// K8sClientQPS is the queries per second limit for the K8s client. Defaults to k8s client defaults.
	K8sClientQPS float32

	// K8sClientBurst is the burst value allowed for the K8s client. Defaults to k8s client defaults.
	K8sClientBurst int
}

var defaultClientParams = ClientParams{
	K8sClientQPS:   defaults.K8sClientQPSLimit,
	K8sClientBurst: defaults.K8sClientBurst,
}

func (def ClientParams) Flags(flags *pflag.FlagSet) {
	flags.Float32(option.K8sClientQPSLimit, def.K8sClientQPS, "Queries per second limit for the K8s client")
	flags.Int(option.K8sClientBurst, def.K8sClientBurst, "Burst value allowed for the K8s client")
}

var defaultSharedConfig = SharedConfig{
	EnableK8s:                    true,
	K8sAPIServerURLs:             []string{},
	K8sKubeConfigPath:            "",
	K8sClientConnectionTimeout:   30 * time.Second,
	K8sClientConnectionKeepAlive: 30 * time.Second,
	K8sHeartbeatTimeout:          30 * time.Second,
	// NOTE: temporarily very large for control-plane-outage testing so the
	// agent does not terminate at cold start when the API server is
	// unreachable. Revert to 1 minute (the previous hardcoded connTimeout)
	// before merging upstream. The default HiveStartTimeout is raised to
	// match so this is not capped at start-hook timeout.
	K8sClientConnectionRetryTimeout: 168 * time.Hour,
	EnableK8sAPIDiscovery:           defaults.K8sEnableAPIDiscovery,
}

func (def SharedConfig) Flags(flags *pflag.FlagSet) {
	flags.Bool(option.EnableK8s, def.EnableK8s, "Enable the k8s clientset")
	flags.StringSlice(option.K8sAPIServerURLs, def.K8sAPIServerURLs, "Kubernetes API server URLs")
	flags.String(option.K8sKubeConfigPath, def.K8sKubeConfigPath, "Absolute path of the kubernetes kubeconfig file")
	flags.Duration(option.K8sClientConnectionTimeout, def.K8sClientConnectionTimeout, "Configures the timeout of K8s client connections. K8s client is disabled if the value is set to 0")
	flags.Duration(option.K8sClientConnectionKeepAlive, def.K8sClientConnectionKeepAlive, "Configures the keep alive duration of K8s client connections. K8 client is disabled if the value is set to 0")
	flags.Duration(option.K8sHeartbeatTimeout, def.K8sHeartbeatTimeout, "Configures the timeout for api-server heartbeat, set to 0 to disable")
	flags.Duration("k8s-client-connection-retry-timeout", def.K8sClientConnectionRetryTimeout, "Maximum time the agent keeps retrying its initial API server connection at startup before exiting. Increase to avoid terminating too soon during a control-plane / API server outage. Effective wait is min(this, --hive-start-timeout). A non-positive value falls back to the built-in default")
	flags.Bool(option.K8sEnableAPIDiscovery, def.EnableK8sAPIDiscovery, "Enable discovery of Kubernetes API groups and resources with the discovery API")
}

func NewClientConfig(cfg SharedConfig, params ClientParams) Config {
	return Config{
		SharedConfig: cfg,
		ClientParams: params,
	}
}

func (cfg Config) IsEnabled() bool {
	if !cfg.EnableK8s {
		return false
	}
	return len(cfg.K8sAPIServerURLs) >= 1 ||
		cfg.K8sKubeConfigPath != "" ||
		(os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
			os.Getenv("KUBERNETES_SERVICE_PORT") != "") ||
		os.Getenv("K8S_NODE_NAME") != ""
}
