//
// Xero Go Example - main
// Author: Henrique Dias
// Last Modification: 2022-10-13 11:52:21
//

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type App struct {
	Config struct {
		Xero struct {
			ClientID     string `json:"client_id"`
			RedirectUri  string `json:"redirect_uri"`
			Code         string `json:"code"`
			CodeVerifier string `json:"code_verifier"`
			Tenant       struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"tenant"`
		} `json:"xero"`
		CachedToken string `json:"cached_token"`
	}
	Xero struct {
		Token    XeroToken
		TenantID string
	}
}

func (app *App) runApp() {
	if app.Config.Xero.Code == "" || app.Config.Xero.CodeVerifier == "" {
		app.xeroGenerateTheLinkToAuthorize()
		os.Exit(0)
	}

	if !app.xeroGetCachedAccessToken() {
		app.xeroExchangeTheCode()
	}

	// fmt.Println("AccessToken:", app.Xero.Token.AccessToken, "RefreshToken:", app.Xero.Token.RefreshToken)

	app.xeroCheckTheTenants()
	app.xeroGetContacts()
	app.xeroGetInvoices()
}

func main() {

	if _, err := os.Stat("config.json"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("The json configuration file not exists!")
		os.Exit(0)
	}

	jsonConfigFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer jsonConfigFile.Close()

	byteJsonConfig, err := io.ReadAll(jsonConfigFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(byteJsonConfig))

	var app App

	if err := json.Unmarshal([]byte(byteJsonConfig), &app.Config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app.runApp()

	os.Exit(0)
}
