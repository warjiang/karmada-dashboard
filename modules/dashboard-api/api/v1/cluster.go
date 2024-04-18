package v1

import "github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"

type PostClusterRequest struct {
	MemberClusterKubeConfig string                   `json:"member_cluster_kubeconfig" binding:"required"`
	SyncMode                v1alpha1.ClusterSyncMode `json:"sync_mode" binding:"required"`
	MemberClusterName       string                   `json:"member_cluster_name" binding:"required"`
	MemberClusterEndpoint   string                   `json:"member_cluster_endpoint"`
	MemberClusterNamespace  string                   `json:"member_cluster_namespace"`
	ClusterProvider         string                   `json:"cluster_provider"`
	ClusterRegion           string                   `json:"cluster_region"`
	ClusterZones            []string                 `json:"cluster_zones"`
}

type PostClusterResponse struct {
}

type DeleteClusterRequest struct {
	MemberClusterName string `uri:"member_cluster_name" binding:"required"`
}
type DeleteClusterResponse struct {
}
