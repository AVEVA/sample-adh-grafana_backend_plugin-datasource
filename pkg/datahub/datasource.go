package datahub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var (
	_ backend.QueryDataHandler      = (*DataHubDataSource)(nil)
	_ backend.CheckHealthHandler    = (*DataHubDataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*DataHubDataSource)(nil)
)

type DataHubDataSource struct {
	dataHubClient *DataHubClient
	namespaceId   string
	communityId   string
	oauthPassThru bool
	useCommunity  bool
}

type DataHubDataSourceOptions struct {
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
func NewDataHubDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Get JSON Data to read datasource settings
	var options DataHubDataSourceOptions
	err := json.Unmarshal(settings.JSONData, &options)
	if err != nil {
		log.DefaultLogger.Warn("error marshalling", "err", err)
		return nil, err
	}

	// Read Client Secret from Secure JSON Data
	var secureData = settings.DecryptedSecureJSONData
	clientSecret, _ := secureData["clientSecret"]

	client := NewDataHubClient(options.Resource, options.ApiVersion, options.TenantId, options.ClientId, clientSecret)
	return &DataHubDataSource{
		dataHubClient: &client,
		namespaceId:   options.NamespaceId,
		communityId:   options.CommunityId,
		oauthPassThru: options.OauthPassThru,
		useCommunity:  options.UseCommunity,
	}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using the new instance factory function.
func (d *DataHubDataSource) Dispose() {
	// Clean up datasource instance resources.
}

// Handles multiple queries and returns multiple responses.
func (d *DataHubDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// retrieve token
	var token string
	if d.oauthPassThru {
		token = req.Headers["Authorization"]
		if len(token) == 0 {
			return nil, fmt.Errorf("Unable to retrieve token")
		}
	} else {
		var err error
		token, err = GetClientToken(d.dataHubClient)
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
func (d *DataHubDataSource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery, token string) (backend.DataResponse, error) {
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
	if d.useCommunity {
		if strings.EqualFold(qm.Collection, "streams") && qm.Id != "" {
			log.DefaultLogger.Debug("Community stream data query")
			frame, err = CommunityStreamsDataQuery(d.dataHubClient,
				d.communityId,
				token,
				qm.Id,
				query.TimeRange.From.Format(time.RFC3339),
				query.TimeRange.To.Format(time.RFC3339))
		} else if strings.EqualFold(qm.Collection, "streams") {
			log.DefaultLogger.Debug("Community stream query")
			frame, err = CommunityStreamsQuery(d.dataHubClient, d.communityId, token, qm.Query)
		}
	} else {
		if strings.EqualFold(qm.Collection, "streams") && qm.Id != "" {
			log.DefaultLogger.Debug("Stream data query")
			frame, err = StreamsDataQuery(d.dataHubClient,
				d.namespaceId,
				token,
				qm.Id,
				query.TimeRange.From.Format(time.RFC3339),
				query.TimeRange.To.Format(time.RFC3339))
		} else if strings.EqualFold(qm.Collection, "streams") {
			log.DefaultLogger.Debug("Stream query")
			frame, err = StreamsQuery(d.dataHubClient, d.namespaceId, token, qm.Query)
		}
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	log.DefaultLogger.Info("We made it")

	return response, err
}

// Handles health checks sent from Grafana to the plugin.
func (d *DataHubDataSource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	// Retrieve token
	var token string
	if d.oauthPassThru {
		return &backend.CheckHealthResult{
			Status:  status,
			Message: message,
		}, nil
	} else {
		var err error
		token, err = GetClientToken(d.dataHubClient)
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
	if d.useCommunity {
		path = d.dataHubClient.resource + "/api/" + d.dataHubClient.apiVersion + "/tenants/" + d.dataHubClient.tenantId + "/communities/" + d.communityId
	} else {
		path = d.dataHubClient.resource + "/api/" + d.dataHubClient.apiVersion + "/tenants/" + d.dataHubClient.tenantId + "/namespaces/" + d.namespaceId
	}

	body, err := SdsRequest(d.dataHubClient, token, path, nil)
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
