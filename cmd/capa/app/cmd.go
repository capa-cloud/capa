package app

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"group.rxcloud/capa/pkg/cmd"
)

// NewRootCommand returns the root cobra command of pilot-discovery.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "capa",
		Short:        "Capa Sidecar.",
		Long:         "Capa Sidecar provides cloud application api capabilities in the Mecha Mesh.",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}

	discoveryCmd := newDiscoveryCommand()
	addFlags(discoveryCmd)
	rootCmd.AddCommand(discoveryCmd)
	rootCmd.AddCommand(version.CobraCommand())
	rootCmd.AddCommand(collateral.CobraCommand(rootCmd, &doc.GenManHeader{
		Title:   "Istio Pilot Discovery",
		Section: "pilot-discovery CLI",
		Manual:  "Istio Pilot Discovery",
	}))
	rootCmd.AddCommand(requestCmd)

	return rootCmd
}

func newSidecarCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sidecar",
		Short: "Start Capa Sidecar.",
		Args:  cobra.ExactArgs(0),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		RunE: func(c *cobra.Command, args []string) error {
			flags := c.Flags()
			cmd.PrintFlags(flags)

			proxy, err := initProxy(args)
			if err != nil {
				return err
			}
			proxyConfig, err := config.ConstructProxyConfig(meshConfigFile, serviceCluster, options.ProxyConfigEnv, concurrency, proxy)
			if err != nil {
				return fmt.Errorf("failed to get proxy config: %v", err)
			}
			if out, err := protomarshal.ToYAML(proxyConfig); err != nil {
				log.Infof("Failed to serialize to YAML: %v", err)
			} else {
				log.Infof("Effective config: %s", out)
			}

			envoyOptions := envoy.ProxyConfig{
				LogLevel:          proxyLogLevel,
				ComponentLogLevel: proxyComponentLogLevel,
				LogAsJSON:         loggingOptions.JSONEncoding,
				NodeIPs:           proxy.IPAddresses,
				Sidecar:           proxy.Type == model.SidecarProxy,
				OutlierLogPath:    outlierLogPath,
			}
			agentOptions := options.NewAgentOptions(proxy, proxyConfig)
			// 初始化 EnvoyProxy 对象
			agent := istio_agent.NewAgent(proxyConfig, agentOptions, secOpts, envoyOptions)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// 启动 status server
			// If a status port was provided, start handling status probes.
			if proxyConfig.StatusPort > 0 {
				if err := initStatusServer(ctx, proxy, proxyConfig, agentOptions.EnvoyPrometheusPort, agent); err != nil {
					return err
				}
			}

			go iptableslog.ReadNFLOGSocket(ctx)

			// On SIGINT or SIGTERM, cancel the context, triggering a graceful shutdown
			go cmd.WaitSignalFunc(cancel)

			// 启动 EnvoyProxy
			// Start in process SDS, dns server, xds proxy, and Envoy.
			wait, err := agent.Run(ctx)
			if err != nil {
				return err
			}
			wait()
			return nil
		},
	}
}
