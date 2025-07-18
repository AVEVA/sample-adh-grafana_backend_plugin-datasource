package cds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aveva/connect-data-services/pkg/cds/community"
	"github.com/aveva/connect-data-services/pkg/cds/sds"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type CdsClient struct {
	resource        string
	apiVersion      string
	tenantId        string
	clientId        string
	clientSecret    string
	token           string
	tokenExpiration int64
	client          *http.Client
}

func NewCdsClient(resource string, apiVersion string, tenantId string, clientId string, clientSecret string) CdsClient {
	return CdsClient{
		resource:     resource,
		apiVersion:   apiVersion,
		tenantId:     tenantId,
		clientId:     clientId,
		clientSecret: clientSecret,
		client:       &http.Client{},
	}
}

func GetClientToken(d *CdsClient) (string, error) {
	if (d.tokenExpiration - time.Now().Unix()) > (5 * 60) {
		return ("Bearer " + d.token), nil
	}

	wellKnownEndpoint := d.resource + "/identity/.well-known/openid-configuration"
	req, err := http.NewRequest("GET", wellKnownEndpoint, nil)
	if err != nil {
		log.DefaultLogger.Warn("Error forming request", err.Error())
		return "", err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		log.DefaultLogger.Warn("Error requesting well known endpoints", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.DefaultLogger.Warn("Error reading response", err.Error())
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("Status: " + resp.Status + "\nBody: " + string(body))
		log.DefaultLogger.Warn("Error making request", err)
		return "", err
	}

	var openIdConfig map[string]interface{}

	err = json.Unmarshal(body, &openIdConfig)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		return "", err
	}

	tokenEndpoint := openIdConfig["token_endpoint"].(string)

	resp, err = d.client.PostForm(tokenEndpoint,
		url.Values{
			"client_id":     {d.clientId},
			"client_secret": {d.clientSecret},
			"grant_type":    {"client_credentials"}})

	if err != nil {
		log.DefaultLogger.Warn("Error requesting token", err.Error())
		return "", err
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.DefaultLogger.Warn("Error requesting token", err.Error())
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("Status: " + resp.Status + "\nBody: " + string(body))
		log.DefaultLogger.Warn("Error making request", err)
		return "", err
	}

	var tokenInformation map[string]interface{}

	err = json.Unmarshal(body, &tokenInformation)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		return "", err
	}

	d.token = tokenInformation["access_token"].(string)
	d.tokenExpiration = int64(tokenInformation["expires_in"].(float64)) + time.Now().Unix()

	return ("Bearer " + d.token), nil
}

func SdsRequest(d *CdsClient, token string, path string, headers map[string]string) ([]byte, error) {
	log.DefaultLogger.Debug("Making query to", path)

	// request data or collection items
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.DefaultLogger.Warn("Error forming request", err.Error())
		return nil, err
	}

	req.Header.Add("Authorization", token)

	// add optional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		log.DefaultLogger.Warn("Error making request", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.DefaultLogger.Warn("Error reading request body", err.Error())
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("Status: " + resp.Status + "\nBody: " + string(body))
		log.DefaultLogger.Warn("Error making request", err)
		return nil, err
	}

	return body, nil
}

func StreamsQuery(d *CdsClient, namespaceId string, token string, query string) (*data.Frame, error) {
	basePath := d.resource + "/api/" + d.apiVersion + "/tenants/" + url.QueryEscape(d.tenantId) + "/namespaces/" + url.QueryEscape(namespaceId)
	path := (basePath + "/streams?query=" + url.QueryEscape(query))

	body, err := SdsRequest(d, token, path, nil)
	if err != nil {
		return nil, err
	}

	var streams []sds.SdsStream

	err = json.Unmarshal(body, &streams)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	// create a dataframe
	frame := data.NewFrame("response")

	// create property lists from streams list
	ids := make([]string, len(streams))
	names := make([]string, len(streams))
	for i := 0; i < len(streams); i++ {
		ids[i] = streams[i].Id
		names[i] = streams[i].Name
	}

	// add fields
	frame.Fields = append(frame.Fields,
		data.NewField("Id", nil, ids),
		data.NewField("Name", nil, names),
	)

	return frame, nil
}

func CommunityStreamsQuery(d *CdsClient, communityId string, token string, query string) (*data.Frame, error) {
	basePath := d.resource + "/api/" + d.apiVersion + "/search/communities/" + url.QueryEscape(communityId)

	path := (basePath + "/streams?query=" + url.QueryEscape(query))

	body, err := SdsRequest(d, token, path, nil)
	if err != nil {
		return nil, err
	}

	var streams []community.StreamSearchResult

	err = json.Unmarshal(body, &streams)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	// create a dataframe
	frame := data.NewFrame("response")

	// create property lists from streams list
	ids := make([]string, len(streams))
	names := make([]string, len(streams))
	for i := 0; i < len(streams); i++ {
		// replace api version for compatibility with preview route
		// this can be removed once community features are released
		ids[i] = strings.Replace(streams[i].Self, "/v1/", "/"+d.apiVersion+"/", 1)
		names[i] = streams[i].Name
	}

	// add fields
	frame.Fields = append(frame.Fields,
		data.NewField("Id", nil, ids),
		data.NewField("Name", nil, names),
	)

	return frame, nil
}

func StreamsDataQuery(d *CdsClient, namespaceId string, token string, id string, startIndex string, endIndex string) (*data.Frame, error) {
	basePath := d.resource + "/api/" + d.apiVersion + "/tenants/" + url.QueryEscape(d.tenantId) + "/namespaces/" + url.QueryEscape(namespaceId)

	// get type Id
	path := (basePath + "/streams/" + url.QueryEscape(id))
	body, err := SdsRequest(d, token, path, nil)
	if err != nil {
		return nil, err
	}

	var stream sds.SdsStream
	err = json.Unmarshal(body, &stream)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	// get type info
	path = (basePath + "/types/" + url.QueryEscape(stream.TypeId))
	body, err = SdsRequest(d, token, path, nil)
	if err != nil {
		return nil, err
	}

	var sdsType sds.SdsType
	err = json.Unmarshal(body, &sdsType)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	log.DefaultLogger.Info(fmt.Sprint(sdsType))

	// get data
	path = (basePath + "/streams/" + url.QueryEscape(id) + "/Data?startIndex=" + url.QueryEscape(startIndex) + "&endIndex=" + url.QueryEscape(endIndex))
	body, err = SdsRequest(d, token, path, nil)
	if err != nil {
		return nil, err
	}

	var sdsData []map[string]interface{}
	err = json.Unmarshal(body, &sdsData)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	return createDataFrameFromSdsData(stream.Name, sdsType, sdsData)
}

func CommunityStreamsDataQuery(d *CdsClient, communityId string, token string, self string, startIndex string, endIndex string) (*data.Frame, error) {

	// make a community header
	communityHeader := map[string]string{
		"Community-Id": url.QueryEscape(communityId),
	}

	// get stream
	path := self
	body, err := SdsRequest(d, token, path, communityHeader)
	if err != nil {
		return nil, err
	}

	var stream sds.SdsStream
	err = json.Unmarshal(body, &stream)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	// get resolved type info
	path = (self + "/resolved")
	body, err = SdsRequest(d, token, path, communityHeader)
	if err != nil {
		return nil, err
	}

	var sdsResolvedStream sds.SdsResolvedStream
	err = json.Unmarshal(body, &sdsResolvedStream)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	// get data
	path = (self + "/Data?startIndex=" + url.QueryEscape(startIndex) + "&endIndex=" + url.QueryEscape(endIndex))
	body, err = SdsRequest(d, token, path, communityHeader)
	if err != nil {
		return nil, err
	}

	var sdsData []map[string]interface{}
	err = json.Unmarshal(body, &sdsData)
	if err != nil {
		log.DefaultLogger.Warn("Error parsing json", err.Error())
		log.DefaultLogger.Warn(fmt.Sprint(string(body)))
		return nil, err
	}

	return createDataFrameFromSdsData(stream.Name, sdsResolvedStream.SdsType, sdsData)
}

func createDataFrameFromSdsData(dataFrameName string, sdsType sds.SdsType, sdsData []map[string]interface{}) (*data.Frame, error) {
	// create a dataframe
	frame := data.NewFrame(dataFrameName)

	// create columns in dataframe
	for i := 0; i < len(sdsType.Properties); i++ {
		typeCodeString := sdsType.Properties[i].SdsType.SdsTypeCode
		frame.Fields = append(frame.Fields,
			data.NewField(sdsType.Properties[i].Id, nil, createSdsValueList(typeCodeString)))
	}

	// add data to rows
	for i := 0; i < len(sdsData); i++ {
		row := make([]interface{}, len(sdsType.Properties))
		for j := 0; j < len(sdsType.Properties); j++ {
			row[j] = convertSdsValue(sdsType.Properties[j].SdsType.SdsTypeCode, sdsData[i][string(sdsType.Properties[j].Id)])
		}
		frame.AppendRow(row...)
	}

	return frame, nil
}

func createSdsValueList(sdsTypeCode sds.SdsTypeCode) interface{} {
	switch t := sdsTypeCode; t {
	case "DateTime":
		return []time.Time{}
	case "NullableDateTime":
		return []*time.Time{}
	case "Boolean":
		return []bool{}
	case "NullableBoolean":
		return []*bool{}
	case "Int16":
		return []int16{}
	case "NullableInt16":
		return []*int16{}
	case "UInt16":
		return []uint16{}
	case "NullableUInt16":
		return []*uint16{}
	case "Int32":
		return []int32{}
	case "NullableInt32":
		return []*int32{}
	case "UInt32":
		return []uint32{}
	case "NullableUInt32":
		return []*uint32{}
	case "Int64":
		return []int64{}
	case "NullableInt64":
		return []*int64{}
	case "UInt64":
		return []uint64{}
	case "NullableUInt64":
		return []*uint32{}
	case "Single":
		return []float32{}
	case "NullableSingle":
		return []*float32{}
	case "Double":
		return []float64{}
	case "NullableDouble":
		return []*float64{}
	default:
		return []*string{}
	}
}

func convertSdsValue(sdsTypeCode sds.SdsTypeCode, value interface{}) interface{} {

	switch t := sdsTypeCode; t {
	case "DateTime":
		if value == nil {
			return value
		}
		timestamp, _ := time.Parse(time.RFC3339, value.(string))
		return timestamp
	case "NullableDateTime":
		if value == nil {
			return value
		}
		timestamp, _ := time.Parse(time.RFC3339, value.(string))
		return &timestamp
	case "Boolean":
		if value == nil {
			return false
		}
		return true
	case "NullableBoolean":
		if value == nil {
			return value
		}
		valuePointer := true
		return &valuePointer
	case "Int16":
		if value == nil {
			return int16(0)
		}
		return int16(value.(float64))
	case "NullableInt16":
		if value == nil {
			return value
		}
		valuePointer := int16(value.(float64))
		return &valuePointer
	case "UInt16":
		if value == nil {
			return uint16(0)
		}
		return uint16(value.(float64))
	case "NullableUInt16":
		if value == nil {
			return value
		}
		valuePointer := uint16(value.(float64))
		return &valuePointer
	case "Int32":
		if value == nil {
			return int32(0)
		}
		return int32(value.(float64))
	case "NullableInt32":
		if value == nil {
			return value
		}
		valuePointer := int32(value.(float64))
		return &valuePointer
	case "UInt32":
		if value == nil {
			return uint32(0)
		}
		return uint32(value.(float64))
	case "NullableUInt32":
		if value == nil {
			return value
		}
		valuePointer := uint32(value.(float64))
		return &valuePointer
	case "Int64":
		if value == nil {
			return int64(0)
		}
		return int64(value.(float64))
	case "NullableInt64":
		if value == nil {
			return value
		}
		valuePointer := int64(value.(float64))
		return &valuePointer
	case "UInt64":
		if value == nil {
			return uint64(0)
		}
		return uint64(value.(float64))
	case "NullableUInt64":
		if value == nil {
			return value
		}
		valuePointer := uint64(value.(float64))
		return &valuePointer
	case "Single":
		if value == nil {
			return float32(0)
		}
		return float32(value.(float64))
	case "NullableSingle":
		if value == nil {
			return value
		}
		valuePointer := float32(value.(float64))
		return &valuePointer
	case "Double":
		if value == nil {
			return float64(0)
		}
		return value.(float64)
	case "NullableDouble":
		if value == nil {
			return value
		}
		valuePointer := value.(float64)
		return &valuePointer
	default:
		log.DefaultLogger.Info("Default")
		if value == nil {
			return value
		}
		valuePointer := value.(string)
		return &valuePointer
	}
}
