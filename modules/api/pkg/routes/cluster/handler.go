package cluster

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"net/http"
	"warjiang/karmada-dashboard/api/pkg/parser"
	"warjiang/karmada-dashboard/api/pkg/resource/cluster"
	"warjiang/karmada-dashboard/api/pkg/router"
	"warjiang/karmada-dashboard/client"
)

func handleGetClusterList(c *gin.Context) {
	karmadaClient, err := client.Client(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	dataSelect := parser.ParseDataSelectPathParameter(c)
	result, err := cluster.GetClusterList(karmadaClient, dataSelect)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func handleGetClusterDetail(c *gin.Context) {
	karmadaClient, err := client.Client(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	name := c.Param("name")
	karmadaClient = karmadaClient
	name = name
	return
}

func init() {
	r := router.V1()
	r.GET("/cluster", handleGetClusterList)
	r.GET("/cluster/{name}", handleGetClusterDetail)
}
