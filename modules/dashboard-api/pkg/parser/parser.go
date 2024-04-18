package parser

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"warjiang/karmada-dashboard/dashboard-api/pkg/dataselect"
)

func parsePaginationPathParameter(request *gin.Context) *dataselect.PaginationQuery {
	itemsPerPage, err := strconv.ParseInt(request.Query("itemsPerPage"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	page, err := strconv.ParseInt(request.Query("page"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	// Frontend pages start from 1 and backend starts from 0
	return dataselect.NewPaginationQuery(int(itemsPerPage), int(page-1))
}

func parseFilterPathParameter(request *gin.Context) *dataselect.FilterQuery {
	return dataselect.NewFilterQuery(strings.Split(request.Query("filterBy"), ","))
}

// Parses query parameters of the request and returns a SortQuery object
func parseSortPathParameter(request *gin.Context) *dataselect.SortQuery {
	return dataselect.NewSortQuery(strings.Split(request.Query("sortBy"), ","))
}

/*
// Parses query parameters of the request and returns a MetricQuery object
func parseMetricPathParameter(request *restful.Request) *dataselect.MetricQuery {
	metricNamesParam := request.QueryParameter("metricNames")
	var metricNames []string
	if metricNamesParam != "" {
		metricNames = strings.Split(metricNamesParam, ",")
	} else {
		metricNames = nil
	}
	aggregationsParam := request.QueryParameter("aggregations")
	var rawAggregations []string
	if aggregationsParam != "" {
		rawAggregations = strings.Split(aggregationsParam, ",")
	} else {
		rawAggregations = nil
	}
	aggregationModes := metricapi.AggregationModes{}
	for _, e := range rawAggregations {
		aggregationModes = append(aggregationModes, metricapi.AggregationMode(e))
	}
	return dataselect.NewMetricQuery(metricNames, aggregationModes)
}
*/

// ParseDataSelectPathParameter parses query parameters of the request and returns a DataSelectQuery object
func ParseDataSelectPathParameter(request *gin.Context) *dataselect.DataSelectQuery {
	paginationQuery := parsePaginationPathParameter(request)
	sortQuery := parseSortPathParameter(request)
	filterQuery := parseFilterPathParameter(request)
	return dataselect.NewDataSelectQuery(paginationQuery, sortQuery, filterQuery)
}
