package proxy

import (
	"sync"
	"time"
)

// Proxy contains information about an specific instance of a proxy (envoy sidecar, gateway,
// etc). The Proxy is initialized when a sidecar connects to Pilot, and populated from
// 'node' info in the protocol as well as data extracted from registries.
//
// In current Istio implementation nodes use a 4-parts '~' delimited ID.
// Type~IPAddress~ID~Domain
type Proxy struct {
	sync.RWMutex

	// Type specifies the node type. First part of the ID.
	Type NodeType

	// IPAddresses is the IP addresses of the proxy used to identify it and its
	// co-located service instances. Example: "10.60.1.6". In some cases, the host
	// where the proxy and service instances reside may have more than one IP address
	IPAddresses []string

	// ID is the unique platform-specific sidecar proxy ID. For k8s it is the pod ID and
	// namespace <podName.namespace>.
	ID string

	// Locality is the location of where Envoy proxy runs. This is extracted from
	// the registry where possible. If the registry doesn't provide a locality for the
	// proxy it will use the one sent via ADS that can be configured in the Envoy bootstrap
	Locality *core.Locality

	// DNSDomain defines the DNS domain suffix for short hostnames (e.g.
	// "default.svc.cluster.local")
	DNSDomain string

	// ConfigNamespace defines the namespace where this proxy resides
	// for the purposes of network scoping.
	// NOTE: DO NOT USE THIS FIELD TO CONSTRUCT DNS NAMES
	ConfigNamespace string

	// Metadata key-value pairs extending the Node identifier
	Metadata *NodeMetadata

	// the sidecarScope associated with the proxy
	SidecarScope *SidecarScope

	// the sidecarScope associated with the proxy previously
	PrevSidecarScope *SidecarScope

	// The merged gateways associated with the proxy if this is a Router
	MergedGateway *MergedGateway

	// service instances associated with the proxy
	ServiceInstances []*ServiceInstance

	// Istio version associated with the Proxy
	IstioVersion *IstioVersion

	// VerifiedIdentity determines whether a proxy had its identity verified. This
	// generally occurs by JWT or mTLS authentication. This can be false when
	// connecting over plaintext. If this is set to true, we can verify the proxy has
	// access to ConfigNamespace namespace. However, other options such as node type
	// are not part of an Istio identity and thus are not verified.
	VerifiedIdentity *spiffe.Identity

	// IPMode of proxy.
	ipMode IPMode

	// GlobalUnicastIP stores the global unicast IP if available, otherwise nil
	GlobalUnicastIP string

	// XdsResourceGenerator is used to generate resources for the node, based on the PushContext.
	// If nil, the default networking/core v2 generator is used. This field can be set
	// at connect time, based on node metadata, to trigger generation of a different style
	// of configuration.
	XdsResourceGenerator XdsResourceGenerator

	// WatchedResources contains the list of watched resources for the proxy, keyed by the DiscoveryRequest TypeUrl.
	WatchedResources map[string]*WatchedResource

	// XdsNode is the xDS node identifier
	XdsNode *core.Node

	AutoregisteredWorkloadEntryName string

	// LastPushContext stores the most recent push context for this proxy. This will be monotonically
	// increasing in version. Requests should send config based on this context; not the global latest.
	// Historically, the latest was used which can cause problems when computing whether a push is
	// required, as the computed sidecar scope version would not monotonically increase.
	LastPushContext *PushContext
	// LastPushTime records the time of the last push. This is used in conjunction with
	// LastPushContext; the XDS cache depends on knowing the time of the PushContext to determine if a
	// key is stale or not.
	LastPushTime time.Time
}
