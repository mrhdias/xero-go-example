//
// Xero Go Example - xero
// Author: Henrique Dias
// Last Modification: 2022-10-06 22:13:05
//

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

func (app *App) xeroGenerateTheLinkToAuthorize() {
	// Generate a link to authorize the app
	// https://developer.xero.com/documentation/guides/oauth2/pkce-flow

	codeVerifier, err := CreateCodeVerifier()
	if err != nil {
		log.Fatalln(err)
	}

	xeroUrl := "https://login.xero.com/identity/connect/authorize"
	scope := strings.Join([]string{
		"openid",
		"accounting.transactions",
		"accounting.contacts",
		"offline_access",
	}, " ")

	state := strings.ReplaceAll(uuid.NewString(), "-", "")

	codeChallenge := codeVerifier.CodeChallengeS256()
	codeChallengeMethod := "S256"

	req, err := http.NewRequest("GET", xeroUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}

	values := req.URL.Query()
	values.Add("response_type", "code")
	values.Add("client_id", app.Config.Xero.ClientID)
	values.Add("redirect_uri", app.Config.Xero.RedirectUri)
	values.Add("scope", scope)
	values.Add("state", state)
	values.Add("code_challenge", codeChallenge)
	values.Add("code_challenge_method", codeChallengeMethod)
	req.URL.RawQuery = values.Encode()

	message := `Copy/Paste this URL "%s" to the browser where you are already logged in to your Xero account.
This will redirect you to the page where you as an owner of the Xero account need to permit to allow access to your organisation/organisations.
Now once you allow the access, you will be redirected and see another page. From that, you need to grab the URL link.
From this URL, copy the value from "code=…." and define in configuration file.
Don't forget to add the Code Verifier as well: "%s".
When you finish adding the codes run the app again.`

	fmt.Printf(message, req.URL, codeVerifier.String())

	if _, err := os.Stat(app.Config.CachedToken); err == nil {
		if err := os.Remove(app.Config.CachedToken); err != nil {
			log.Fatalln(err)
		}
	}
}

func (app App) xeroCacheTheToken(respContentBytes []byte) {

	if _, err := os.Stat(app.Config.CachedToken); err == nil {
		if err := os.Remove(app.Config.CachedToken); err != nil {
			log.Fatalln(err)
		}
	}

	if err := os.WriteFile(app.Config.CachedToken, respContentBytes, 0644); err != nil {
		log.Fatalln(err)
	}
}

func (app App) xeroExchangeTheCode() {
	xeroUrl := "https://identity.xero.com/connect/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", app.Config.Xero.ClientID)
	data.Set("code", app.Config.Xero.Code)
	data.Set("redirect_uri", app.Config.Xero.RedirectUri)
	data.Set("code_verifier", app.Config.Xero.CodeVerifier)

	fmt.Println("Exchange The Code Data Encoded:", data.Encode())

	req, err := http.NewRequest("POST", xeroUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	respContentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Xero Exchange The Code response:", string(respContentBytes))

	if resp.StatusCode != 200 {
		log.Fatalf("The request post \"connect/token\" from Xero returned an error with status code %d\r\n", resp.StatusCode)
	}

	if err := json.Unmarshal([]byte(respContentBytes), &app.Xero.Token); err != nil {
		log.Fatalln(err)
	}

	if reflect.DeepEqual((XeroToken{}), app.Xero.Token) {
		log.Fatalln("The request post \"connect/token\" from Xero don't return anything!")
	}

	app.xeroCacheTheToken(respContentBytes)

}

func (app *App) xeroGetCachedAccessToken() bool {

	if _, err := os.Stat(app.Config.CachedToken); errors.Is(err, os.ErrNotExist) {
		return false
	}

	content, err := os.ReadFile(app.Config.CachedToken)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.Unmarshal([]byte(content), &app.Xero.Token); err != nil {
		log.Fatalln(err)
	}

	return true
}

func (app *App) xeroRefreshToken() bool {
	xeroUrl := "https://identity.xero.com/connect/token"

	// fmt.Println("AccessToken:", app.Xero.Token.AccessToken, "RefreshToken:", app.Xero.Token.RefreshToken)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", app.Config.Xero.ClientID)
	data.Set("refresh_token", app.Xero.Token.RefreshToken)

	fmt.Println("Refresh Token Data Encoded:", data.Encode())

	req, err := http.NewRequest("POST", xeroUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	respContentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Println("Xero Refresh Token:", string(respContentBytes))

	if resp.StatusCode != 200 {
		log.Fatalf("The request post \"connect/token\" from Xero returned an error with status code %d\r\n", resp.StatusCode)
	}

	if err := json.Unmarshal([]byte(respContentBytes), &app.Xero.Token); err != nil {
		log.Fatalln(err)
	}

	if reflect.DeepEqual((XeroToken{}), app.Xero.Token) {
		log.Fatalln("[Error] The request post \"connect/token\" from Xero don't return anything!")
	}

	app.xeroCacheTheToken(respContentBytes)

	return true
}

func (app *App) xeroCheckTheTenants() {
	// https://developer.xero.com/documentation/guides/oauth2/pkce-flow/#5-check-the-tenants-youre-authorized-to-access

	// fmt.Println("AccessToken:", app.Xero.Token.AccessToken, "RefreshToken:", app.Xero.Token.RefreshToken)

	xeroUrl := "https://api.xero.com/connections"

	client := &http.Client{}

	var xeroTenant []XeroTenant
	refreshToken := false
	for {
		req, err := http.NewRequest("GET", xeroUrl, nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", app.Xero.Token.AccessToken))
		req.Header.Add("content-type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		respContentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		// fmt.Println("Xero Check The Tenants:", string(respContentBytes))

		if resp.StatusCode != 200 {
			if resp.StatusCode == 401 && !refreshToken {
				var xeroConnections XeroConnections
				if err := json.Unmarshal([]byte(respContentBytes), &xeroConnections); err != nil {
					log.Fatalln(err)
				}
				if strings.HasPrefix(xeroConnections.Detail, "TokenExpired:") {
					refreshToken = true
					if app.xeroRefreshToken() {
						continue
					}
				}
			}
			log.Fatalf("The request get \"connections\" from Xero returned an error with status code %d\r\n", resp.StatusCode)
		}

		if err := json.Unmarshal([]byte(respContentBytes), &xeroTenant); err != nil {
			log.Fatalln(err)
		}

		break
	}

	if len(xeroTenant) == 0 {
		log.Fatalln("The request get \"connections\" from Xero don't return anything!")
	}

	for _, tenant := range xeroTenant {
		if strings.EqualFold(tenant.TenantType, "ORGANISATION") && strings.EqualFold(tenant.TenantName, app.Config.Xero.TenantName) {
			app.Xero.TenantID = tenant.TenantID
			break
		}
	}
}

func (app *App) xeroGetContacts() {
	// https://developer.xero.com/documentation/api/accounting/contacts/#get-contacts

	xeroUrl := "https://api.xero.com/api.xro/2.0/Contacts"

	client := &http.Client{}

	refreshToken := false
	for {

		req, err := http.NewRequest("GET", xeroUrl, nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Xero-tenant-id", app.Xero.TenantID)

		values := req.URL.Query()
		values.Add("summaryOnly", "true")
		// values.Add("where", fmt.Sprintf("EmailAddress=\"%s\"", "mrhdias@gmail.com"))
		values.Add("searchTerm", "mrhdias@gmail.com")
		req.URL.RawQuery = values.Encode()

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", app.Xero.Token.AccessToken))

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		respContentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Xero Get Contacts:", string(respContentBytes))

		if resp.StatusCode != 200 {
			if resp.StatusCode == 401 && !refreshToken {
				var xeroConnections XeroConnections
				if err := json.Unmarshal([]byte(respContentBytes), &xeroConnections); err != nil {
					log.Fatalln(err)
				}
				if strings.HasPrefix(xeroConnections.Detail, "TokenExpired:") {
					refreshToken = true
					if app.xeroRefreshToken() {
						continue
					}
				}
			}
			log.Fatalf("The request get \"contacts\" from Xero returned an error with status code %d\r\n", resp.StatusCode)
		}

		break
	}

}

func (app *App) xeroGetInvoices() {
	// https://developer.xero.com/documentation/api/accounting/invoices/#get-invoices

	xeroUrl := "https://api.xero.com/api.xro/2.0/Invoices"

	client := &http.Client{}

	page := 1
	refreshToken := false
	for {
		req, err := http.NewRequest("GET", xeroUrl, nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Xero-tenant-id", app.Xero.TenantID)

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", app.Xero.Token.AccessToken))

		values := req.URL.Query()
		values.Add("page", fmt.Sprintf("%d", page))
		values.Add("createdByMyApp", "true")
		values.Add("summaryOnly", "true")
		// values.Add("where", fmt.Sprintf("Reference=\"%s\"", "#IE-0001"))
		req.URL.RawQuery = values.Encode()

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		respContentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Xero Get Invoices:", string(respContentBytes))

		if resp.StatusCode != 200 {
			if resp.StatusCode == 401 && !refreshToken {
				var xeroConnections XeroConnections
				if err := json.Unmarshal([]byte(respContentBytes), &xeroConnections); err != nil {
					log.Fatalln(err)
				}
				if strings.HasPrefix(xeroConnections.Detail, "TokenExpired:") {
					refreshToken = true
					if app.xeroRefreshToken() {
						continue
					}
				}
			}
			log.Fatalf("The request get \"invoices\" from Xero returned an error with status code %d\r\n", resp.StatusCode)
		}

		page++

		break
	}
}
