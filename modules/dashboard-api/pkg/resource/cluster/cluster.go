package cluster

import (
	"context"
	"github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"warjiang/karmada-dashboard/dashboard-api/pkg/dataselect"
	"warjiang/karmada-dashboard/errors"
	"warjiang/karmada-dashboard/helpers"
	"warjiang/karmada-dashboard/types"
)

// ClusterList contains a list of clusters.
type ClusterList struct {
	ListMeta types.ListMeta `json:"listMeta"`
	Clusters []Cluster      `json:"clusters"`
	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}
type Cluster struct {
	ObjectMeta         types.ObjectMeta          `json:"objectMeta"`
	TypeMeta           types.TypeMeta            `json:"typeMeta"`
	Ready              metav1.ConditionStatus    `json:"ready"`
	KubernetesVersion  string                    `json:"kubernetesVersion,omitempty"`
	SyncMode           v1alpha1.ClusterSyncMode  `json:"syncMode"`
	NodeSummary        *v1alpha1.NodeSummary     `json:"nodeSummary,omitempty"`
	AllocatedResources ClusterAllocatedResources `json:"allocatedResources"`
}

// GetNodeList returns a list of all Nodes in the cluster.
func GetClusterList(client karmadaclientset.Interface, dsQuery *dataselect.DataSelectQuery) (*ClusterList, error) {
	clusters, err := client.ClusterV1alpha1().Clusters().List(context.TODO(), helpers.ListEverything)
	nonCriticalErrors, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return toClusterList(client, clusters.Items, nonCriticalErrors, dsQuery), nil
}

func toClusterList(client karmadaclientset.Interface, clusters []v1alpha1.Cluster, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *ClusterList {
	clusterList := &ClusterList{
		Clusters: make([]Cluster, 0),
		ListMeta: types.ListMeta{TotalItems: len(clusters)},
		Errors:   nonCriticalErrors,
	}
	clusterCells, filteredTotal := dataselect.GenericDataSelectWithFilter(
		toCells(clusters),
		dsQuery,
	)
	clusters = fromCells(clusterCells)
	clusterList.ListMeta = types.ListMeta{TotalItems: filteredTotal}
	for _, cluster := range clusters {
		clusterList.Clusters = append(clusterList.Clusters, toCluster(cluster, nil))
	}
	return clusterList
}

func toCluster(cluster v1alpha1.Cluster, nodes []v1.Node) Cluster {
	allocatedResources, err := getclusterAllocatedResources(cluster)
	if err != nil {
		log.Printf("Couldn't get allocated resources of %s node: %s\n", cluster.Name, err)
	}

	return Cluster{
		ObjectMeta:         types.NewObjectMeta(cluster.ObjectMeta),
		TypeMeta:           types.NewTypeMeta(types.ResourceKindCluster),
		Ready:              getClusterConditionStatus(cluster, metav1.ConditionTrue),
		KubernetesVersion:  cluster.Status.KubernetesVersion,
		AllocatedResources: allocatedResources,
		SyncMode:           cluster.Spec.SyncMode,
		NodeSummary:        cluster.Status.NodeSummary,
	}
}

func getClusterConditionStatus(cluster v1alpha1.Cluster, conditionType metav1.ConditionStatus) metav1.ConditionStatus {
	for _, condition := range cluster.Status.Conditions {
		if condition.Status == conditionType {
			return condition.Status
		}
	}
	return metav1.ConditionUnknown
}

func CreateCluster(client karmadaclientset.Interface) {

}
