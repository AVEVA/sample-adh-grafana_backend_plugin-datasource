package cds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aveva/connect-data-services/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*CdsDataSource)(nil)
	_ backend.CheckHealthHandler    = (*CdsDataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*CdsDataSource)(nil)
)

type CdsDataSource struct {
	cdsClient *CdsClient
	settings  *models.CdsSettings
}

type CdsDataSourceOptions struct {
	Resource      string `json:"resource"`
	ApiVersion    string `json:"apiVersion"`
	TenantId      string `json:"tenantId"`
	NamespaceId   string `json:"namespaceId"`
	UseCommunity  bool   `json:"useCommunity"`
	CommunityId   string `json:"communityId"`
	ClientId      string `json:"clientId"`
	OauthPassThru bool   `json:"oauthPassThru"`
}

type QueryModel struct {
	Collection string `json:"collection"`
	Query      string `json:"queryText"`
	Id         string `json:"id"`
}

type CheckHealthResponseBody struct {
	Id string `json:"Id"`
}

// Creates a new datasource instance.
func NewCdsDataSource(dis backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Error("NEW DATA SOURCE CALLED")
	settings, err := models.LoadPluginSettings(dis)
	if err != nil {
		return nil, err
	}

	client := NewCdsClient(settings.Resource, settings.ApiVersion, settings.TenantId, settings.ClientId, settings.Secrets.ClientSecret)
	return &CdsDataSource{
		cdsClient: &client,
		settings:  settings,
	}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using the new instance factory function.
func (d *CdsDataSource) Dispose() {
	// Clean up datasource instance resources.
}

// Handles multiple queries and returns multiple responses.
func (d *CdsDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// retrieve token
	var token string
	if d.settings.OauthPassThru {
		token = req.Headers["Authorization"]
		if len(token) == 0 {
			return nil, fmt.Errorf("Unable to retrieve token")
		}
	} else {
		var err error
		token, err = GetClientToken(d.cdsClient)
		if err != nil {
			log.DefaultLogger.Warn("Unable to retrieve token", err.Error())
			return nil, err
		}
	}

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res, err := d.query(ctx, req.PluginContext, q, token)

		if err != nil {
			return nil, err
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// Handles the individual queries from QueryData.
func (d *CdsDataSource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery, token string) (backend.DataResponse, error) {
	log.DefaultLogger.Info("Running query", "query", query)
	response := backend.DataResponse{}

	// unmarshal the JSON into our QueryModel.
	var qm QueryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response, nil
	}

	// determine what type of query to use
	frame := data.NewFrame("response")
	var err error
	if d.settings.UseCommunity {
		if strings.EqualFold(qm.Collection, "streams") && qm.Id != "" {
			log.DefaultLogger.Debug("Community stream data query")
			frame, err = CommunityStreamsDataQuery(d.cdsClient,
				d.settings.CommunityId,
				token,
				qm.Id,
				query.TimeRange.From.Format(time.RFC3339),
				query.TimeRange.To.Format(time.RFC3339))
		} else if strings.EqualFold(qm.Collection, "streams") {
			log.DefaultLogger.Debug("Community stream query")
			frame, err = CommunityStreamsQuery(d.cdsClient, d.settings.CommunityId, token, qm.Query)
		}
	} else {
		if strings.EqualFold(qm.Collection, "streams") && qm.Id != "" {
			log.DefaultLogger.Debug("Stream data query")
			frame, err = StreamsDataQuery(d.cdsClient,
				d.settings.NamespaceId,
				token,
				qm.Id,
				query.TimeRange.From.Format(time.RFC3339),
				query.TimeRange.To.Format(time.RFC3339))
		} else if strings.EqualFold(qm.Collection, "streams") {
			log.DefaultLogger.Debug("Stream query")
			frame, err = StreamsQuery(d.cdsClient, d.settings.NamespaceId, token, qm.Query)
		}
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	log.DefaultLogger.Info("We made it")

	return response, err
}

// Handles health checks sent from Grafana to the plugin.
func (d *CdsDataSource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Error("CHECK HEALTH CALLED")
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	// Retrieve token
	var token string
	if d.settings.OauthPassThru {
		return &backend.CheckHealthResult{
			Status:  status,
			Message: message,
		}, nil
	} else {
		var err error
		token, err = GetClientToken(d.cdsClient)
		if err != nil {
			log.DefaultLogger.Warn("Error unable to get token health check", err.Error())
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Unable to retrieve token",
			}, nil
		}
	}

	// Make a request to test the token
	var path string
	if d.settings.UseCommunity {
		path = d.cdsClient.resource + "/api/" + d.cdsClient.apiVersion + "/tenants/" + d.cdsClient.tenantId + "/communities/" + d.settings.CommunityId
	} else {
		path = d.cdsClient.resource + "/api/" + d.cdsClient.apiVersion + "/tenants/" + d.cdsClient.tenantId + "/namespaces/" + d.settings.NamespaceId
	}

	body, err := SdsRequest(d.cdsClient, token, path, nil)
	if err != nil {
		log.DefaultLogger.Warn("Error test request health check", err.Error())
		status = backend.HealthStatusError
		message = "Invalid Configuration"
	}

	var responseJson CheckHealthResponseBody

	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing resonse health check", err.Error())
		status = backend.HealthStatusError
		message = "Invalid Configuration"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
