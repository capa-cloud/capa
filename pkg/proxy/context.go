package proxy

// NodeType decides the responsibility of the proxy serves in the mesh
type NodeType string

const (
	// SidecarProxy type is used for sidecar proxies in the application containers
	SidecarProxy NodeType = "sidecar"

	// ProxyLess type is used for proxyless in the application containers
	ProxyLess NodeType = "proxyless"
)

var NodeTypes = [...]NodeType{SidecarProxy, ProxyLess}
