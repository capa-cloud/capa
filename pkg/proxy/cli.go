package proxy

import (
	"fmt"
)

func initProxy(args []string) (*Proxy, error) {
	proxy := &Proxy{
		Type: model.SidecarProxy,
	}
	if len(args) > 0 {
		proxy.Type = model.NodeType(args[0])
		if !model.IsApplicationNodeType(proxy.Type) {
			return nil, fmt.Errorf("Invalid proxy Type: " + string(proxy.Type))
		}
	}

	// Extract pod variables.
	podName := options.PodNameVar.Get()
	podNamespace := options.PodNamespaceVar.Get()
	proxy.ID = podName + "." + podNamespace

	// If not set, set a default based on platform - podNamespace.svc.cluster.local for
	// K8S
	proxy.DNSDomain = getDNSDomain(podNamespace, dnsDomain)
	log.WithLabels("ips", proxy.IPAddresses, "type", proxy.Type, "id", proxy.ID, "domain", proxy.DNSDomain).Info("Proxy role")

	return proxy, nil
}
