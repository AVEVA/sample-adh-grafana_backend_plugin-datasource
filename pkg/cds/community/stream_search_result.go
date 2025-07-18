package community

type StreamSearchResult struct {
	TypeId      string `json:"TypeId"`
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Self        string `json:"Self"`
	TenantId    string `json:"TenantId"`
	NamespaceId string `json:"NamespaceId"`
	CommunityId string `json:"CommunityId"`
}
