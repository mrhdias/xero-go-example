//
// Xero Go Example - types
// Author: Henrique Dias
// Last Modification: 2022-10-06 22:14:41
//

package main

type XeroToken struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type XeroTenant struct {
	ID             string `json:"id"`
	AuthEventID    string `json:"authEventId"`
	TenantID       string `json:"tenantId"`
	TenantType     string `json:"tenantType"`
	TenantName     string `json:"tenantName"`
	CreatedDateUtc string `json:"createdDateUtc"`
	UpdatedDateUtc string `json:"updatedDateUtc"`
}

type XeroConnections struct {
	Type       interface{} `json:"Type"`
	Title      string      `json:"Title"`
	Status     int         `json:"Status"`
	Detail     string      `json:"Detail"`
	Instance   string      `json:"Instance"`
	Extensions struct {
	} `json:"Extensions"`
}
