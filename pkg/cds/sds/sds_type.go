package sds

type SdsType struct {
	Id          string            `json:"Id"`
	SdsTypeCode SdsTypeCode       `json:"SdsTypeCode"`
	Name        string            `json:"Name"`
	Properties  []SdsTypeProperty `json:"Properties"`
}
