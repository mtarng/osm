package cds

import (
	mapset "github.com/deckarep/golang-set"
	xds_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	xds_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/envoy"
	"github.com/openservicemesh/osm/pkg/envoy/registry"
)

// NewResponse creates a new Cluster Discovery Response.
func NewResponse(meshCatalog catalog.MeshCataloger, proxy *envoy.Proxy, _ *xds_discovery.DiscoveryRequest, cfg configurator.Configurator, _ certificate.Manager, proxyRegistry *registry.ProxyRegistry) ([]types.Resource, error) {
	svcList, err := proxyRegistry.ListProxyServices(proxy)
	if err != nil {
		log.Error().Err(err).Msgf("Error looking up MeshService for Envoy with SerialNumber=%s on Pod with UID=%s", proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
		return nil, err
	}

	var clusters []*xds_cluster.Cluster

	proxyIdentity, err := envoy.GetServiceAccountFromProxyCertificate(proxy.GetCertificateCommonName())
	if err != nil {
		log.Error().Err(err).Msgf("Error looking up proxy identity for proxy with SerialNumber=%s on Pod with UID=%s",
			proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
		return nil, err
	}

	// Build remote clusters based on allowed outbound services
	for _, dstService := range meshCatalog.ListAllowedOutboundServicesForIdentity(proxyIdentity.ToServiceIdentity()) {
		cluster, err := getUpstreamServiceCluster(proxyIdentity.ToServiceIdentity(), dstService, cfg)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to construct service cluster for service %s for proxy with XDS Certificate SerialNumber=%s on Pod with UID=%s",
				dstService.Name, proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
			return nil, err
		}

		clusters = append(clusters, cluster)
	}

	// Create a local cluster for each service behind the proxy.
	// The local cluster will be used to handle incoming traffic.
	for _, proxyService := range svcList {
		localClusterName := envoy.GetLocalClusterNameForService(proxyService)
		localCluster, err := getLocalServiceCluster(meshCatalog, proxyService, localClusterName)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get local cluster config for proxy %s", proxyService)
			return nil, err
		}
		clusters = append(clusters, localCluster)
	}

	// Add egress clusters based on applied policies
	if egressTrafficPolicy, err := meshCatalog.GetEgressTrafficPolicy(proxyIdentity.ToServiceIdentity()); err != nil {
		log.Error().Err(err).Msgf("Error retrieving egress policies for proxy with identity %s, skipping egress clusters", proxyIdentity)
	} else {
		if egressTrafficPolicy != nil {
			clusters = append(clusters, getEgressClusters(egressTrafficPolicy.ClustersConfigs)...)
		}
	}

	outboundPassthroughCluser, err := getOriginalDestinationEgressCluster(envoy.OutboundPassthroughCluster)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to passthrough cluster for egress for proxy %s", envoy.OutboundPassthroughCluster)
		return nil, err
	}

	// Add an outbound passthrough cluster for egress if global mesh-wide Egress is enabled
	if cfg.IsEgressEnabled() {
		clusters = append(clusters, outboundPassthroughCluser)
	}

	// Add an inbound prometheus cluster (from Prometheus to localhost)
	if cfg.IsPrometheusScrapingEnabled() {
		clusters = append(clusters, getPrometheusCluster())
	}

	// Add an outbound tracing cluster (from localhost to tracing sink)
	if cfg.IsTracingEnabled() {
		clusters = append(clusters, getTracingCluster(cfg))
	}

	alreadyAdded := mapset.NewSet()
	var cdsResources []types.Resource
	for _, cluster := range clusters {
		if alreadyAdded.Contains(cluster.Name) {
			log.Error().Msgf("Found duplicate clusters with name %s; Duplicate will not be sent to Envoy with XDS Certificate SerialNumber=%s on Pod with UID=%s",
				cluster.Name, proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
			continue
		}
		alreadyAdded.Add(cluster.Name)
		cdsResources = append(cdsResources, cluster)
	}

	return cdsResources, nil
}
