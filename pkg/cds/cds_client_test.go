package cds

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Tests struct {
	name          string
	server        *httptest.Server
	response      *data.Frame
	expectedError error
}

var apiVersion = "v1"
var tenantId = "default"
var namespaceId = "default"
var communityId = "default"

func TestStreamsQuery(t *testing.T) {
	tests := []Tests{
		{
			name: "streams-query",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"TypeId": "StreamType1",
						"Id": "StreamId1",
						"Name": "StreamName1",
						"Description": "",
						"InterpolationMode": null,
						"ExtrapolationMode": null
					},
					{
						"TypeId": "StreamType2",
						"Id": "StreamId2",
						"Name": "StreamName2",
						"Description": "",
						"InterpolationMode": null,
						"ExtrapolationMode": null
					},
					{
						"TypeId": "StreamType3",
						"Id": "StreamId3",
						"Name": "StreamName3",
						"Description": "",
						"InterpolationMode": null,
						"ExtrapolationMode": null
					}
				]`))
			})),
			response: data.NewFrame("response",
				data.NewField("Id", nil, []string{"StreamId1", "StreamId2", "StreamId3"}),
				data.NewField("Name", nil, []string{"StreamName1", "StreamName2", "StreamName3"}),
			),
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			client := NewCdsClient(test.server.URL, apiVersion, tenantId, "", "")
			resp, err := StreamsQuery(&client, namespaceId, "token", "")

			if !reflect.DeepEqual(resp, test.response) {
				t.Errorf("FAILED: expected %v, got %v\n", test.response, resp)
			}
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Expected error FAILED: expected %v, got %v\n", test.expectedError, err)
			}
		})
	}
}

func TestStreamsDataQuery(t *testing.T) {
	basePath := "/api/" + apiVersion + "/tenants/" + tenantId + "/namespaces/" + namespaceId
	mux := http.NewServeMux()

	mux.HandleFunc(basePath+"/streams/StreamId1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			{
				"TypeId": "StreamType1",
				"Id": "StreamId1",
				"Name": "StreamName1",
				"Description": "",
				"InterpolationMode": null,
				"ExtrapolationMode": null
			}`))
	})

	mux.HandleFunc(basePath+"/types/StreamType1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
		{
			"Id": "StreamType1",
			"Name": "StreamType1",
			"Description": "",
			"SdsTypeCode": 1,
			"IsGenericType": false,
			"IsReferenceType": false,
			"GenericArguments": null,
			"Properties": [
				{
					"Id": "Timestamp",
					"Name": "Timestamp",
					"Description": null,
					"Order": 0,
					"IsKey": true,
					"FixedSize": 0,
					"SdsType": {
						"Id": "PropertyId1",
						"Name": "DateTime",
						"Description": null,
						"SdsTypeCode": 16,
						"IsGenericType": false,
						"IsReferenceType": false,
						"GenericArguments": null,
						"Properties": null,
						"BaseType": null,
						"DerivedTypes": null,
						"InterpolationMode": 0,
						"ExtrapolationMode": 0
					},
					"Value": null,
					"Uom": null,
					"InterpolationMode": null,
					"IsQuality": false
				},
				{
					"Id": "Value",
					"Name": "Value",
					"Description": null,
					"Order": 0,
					"IsKey": false,
					"FixedSize": 0,
					"SdsType": {
						"Id": "PropertyId2",
						"Name": "Single",
						"Description": null,
						"SdsTypeCode": 13,
						"IsGenericType": true,
						"IsReferenceType": false,
						"GenericArguments": null,
						"Properties": null,
						"BaseType": null,
						"DerivedTypes": null,
						"InterpolationMode": 0,
						"ExtrapolationMode": 0
					},
					"Value": null,
					"Uom": null,
					"InterpolationMode": null,
					"IsQuality": false
				}
			],
			"BaseType": null,
			"DerivedTypes": null,
			"InterpolationMode": 4,
			"ExtrapolationMode": 2
		}`))
	})

	mux.HandleFunc(basePath+"/streams/StreamId1/Data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"Timestamp": "2022-06-04T00:00:00Z",
				"Value": 0
			},
			{
				"Timestamp": "2022-06-05T00:00:00Z",
				"Value": 1
			}
		]`))
	})

	tests := []Tests{
		{
			name:   "streams-data-query",
			server: httptest.NewServer(mux),
			response: data.NewFrame("StreamName1",
				data.NewField("Timestamp", nil, []time.Time{time.Date(2022, 6, 4, 0, 0, 0, 0, time.UTC), time.Date(2022, 6, 5, 0, 0, 0, 0, time.UTC)}),
				data.NewField("Value", nil, []float32{float32(0), float32(1)}),
			),
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			client := NewCdsClient(test.server.URL, apiVersion, tenantId, "", "")
			resp, err := StreamsDataQuery(&client, namespaceId, "token", "StreamId1", "", "")

			if !reflect.DeepEqual(resp, test.response) {
				t.Errorf("FAILED: expected %v, got %v\n", test.response, resp)
			}
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Expected error FAILED: expected %v, got %v\n", test.expectedError, err)
			}
		})
	}
}

func TestCommunityStreamsQuery(t *testing.T) {
	tests := []Tests{
		{
			name: "community-streams-query",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[
					{
						"Name": "StreamName1",
						"Id": "StreamId1",
						"TypeId": "StreamType1",
						"Description": "",
						"Self": "http://host/StreamId1",
						"TenantId": "tenantId1",
						"TenantName": "tenantName1",
						"NamespaceId": "namespaceId1",
						"CommunityId": "communityId1"
					},
					{
						"Name": "StreamName2",
						"Id": "StreamId2",
						"TypeId": "StreamType2",
						"Description": "",
						"Self": "http://host/StreamId2",
						"TenantId": "tenantId2",
						"TenantName": "tenantName2",
						"NamespaceId": "namespaceId2",
						"CommunityId": "communityId2"
					},
					{
						"Name": "StreamName3",
						"Id": "StreamId3",
						"TypeId": "StreamType3",
						"Description": "",
						"Self": "http://host/StreamId3",
						"TenantId": "tenantId3",
						"TenantName": "tenantName3",
						"NamespaceId": "namespaceId3",
						"CommunityId": "communityId3"
					}
				]`))
			})),
			response: data.NewFrame("response",
				data.NewField("Id", nil, []string{"http://host/StreamId1", "http://host/StreamId2", "http://host/StreamId3"}),
				data.NewField("Name", nil, []string{"StreamName1", "StreamName2", "StreamName3"}),
			),
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			client := NewCdsClient(test.server.URL, apiVersion, tenantId, "", "")
			resp, err := CommunityStreamsQuery(&client, namespaceId, "token", "")

			if !reflect.DeepEqual(resp, test.response) {
				t.Errorf("FAILED: expected %v, got %v\n", test.response, resp)
			}
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Expected error FAILED: expected %v, got %v\n", test.expectedError, err)
			}
		})
	}
}

func TestCommunityStreamsDataQuery(t *testing.T) {
	basePath := "/api/" + apiVersion + "/tenants/" + tenantId + "/namespaces/" + namespaceId
	mux := http.NewServeMux()

	mux.HandleFunc(basePath+"/streams/StreamId1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			{
				"TypeId": "StreamType1",
				"Id": "StreamId1",
				"Name": "StreamName1",
				"Description": "",
				"InterpolationMode": null,
				"ExtrapolationMode": null
			}
		`))
	})

	mux.HandleFunc(basePath+"/streams/StreamId1/resolved", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
		{
			"Id": "StreamId1",
			"Name": "StreamName1",
			"Description": "",
			"Resolved": true,
			"Type": {
				"Properties": [
					{
						"SdsType": {
							"Properties": [],
							"SdsTypeCode": "DateTime",
							"ExtrapolationMode": "All",
							"InterpolationMode": "Continuous",
							"Id": "PropertyId1",
							"Name": "DateTime",
							"Description": null,
							"IsGenericType": false,
							"IsReferenceType": false,
							"GenericArguments": null
						},
						"InterpolationMode": null,
						"Id": "Timestamp",
						"Name": "Timestamp",
						"Description": null,
						"Order": 0,
						"IsKey": true,
						"FixedSize": 0,
						"Value": null,
						"Uom": null,
						"IsQuality": false
					},
					{
						"SdsType": {
							"Properties": [],
							"SdsTypeCode": "Single",
							"ExtrapolationMode": "All",
							"InterpolationMode": "Continuous",
							"Id": "PropertyId2",
							"Name": "Single",
							"Description": null,
							"IsGenericType": true,
							"IsReferenceType": false,
							"GenericArguments": null
						},
						"InterpolationMode": null,
						"Id": "Value",
						"Name": "Value",
						"Description": null,
						"Order": 0,
						"IsKey": false,
						"FixedSize": 0,
						"Value": null,
						"Uom": null,
						"IsQuality": false
					}
				],
				"SdsTypeCode": "Object",
				"ExtrapolationMode": "Forward",
				"InterpolationMode": "ContinuousNullableLeading",
				"Id": "StreamType1",
				"Name": "StreamType1",
				"Description": "",
				"IsGenericType": false,
				"IsReferenceType": false,
				"GenericArguments": null
			}
		}`))
	})

	mux.HandleFunc(basePath+"/streams/StreamId1/Data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"Timestamp": "2022-06-04T00:00:00Z",
				"Value": 0
			},
			{
				"Timestamp": "2022-06-05T00:00:00Z",
				"Value": 1
			}
		]`))
	})

	tests := []Tests{
		{
			name:   "community-streams-data-query",
			server: httptest.NewServer(mux),
			response: data.NewFrame("StreamName1",
				data.NewField("Timestamp", nil, []time.Time{time.Date(2022, 6, 4, 0, 0, 0, 0, time.UTC), time.Date(2022, 6, 5, 0, 0, 0, 0, time.UTC)}),
				data.NewField("Value", nil, []float32{float32(0), float32(1)}),
			),
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			client := NewCdsClient(test.server.URL, apiVersion, tenantId, "", "")
			resp, err := CommunityStreamsDataQuery(&client, communityId, "token", test.server.URL+basePath+"/streams/StreamId1", "", "")

			if !reflect.DeepEqual(resp, test.response) {
				t.Errorf("FAILED: expected %v, got %v\n", test.response, resp)
			}
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Expected error FAILED: expected %v, got %v\n", test.expectedError, err)
			}
		})
	}
}
