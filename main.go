package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"

	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile := "credentials/token.json"
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type FinalRequest struct {
	Requests `json:"requests"`
}

type Requests struct {
	addConditionalFormatRule `json:"addConditionalFormatRule"`
}

type addConditionalFormatRule struct {
	Rule `json:"rule,omitempty"`
}

type Rule struct {
	Ranges `json:"ranges,omitempty"`
	BooleanRule `json:"booleanRule,omitempty"`
}

type Ranges struct {
	SheetId int `json:"sheetId"`
	StartRowIndex int `json:"startRowIndex"`
	EndRowIndex int `json:"endRowIndex"`
	StartColumnIndex int `json:"startColumnIndex"`
	EndColumnIndex int `json:"endColumnIndex"`
}

type BooleanRule struct {
	Format struct {
		TextFormat struct {
			Strikethrough bool `json:"strikethrough,omitempty"`
		} `json:"textFormat,omitempty"`
	} `json:"format,omitempty"`

	Condition struct {
		Type string `json:"type,omitempty"`
	}`json:"condition,omitempty"`
}

func main() {

	payload := &FinalRequest{}
	payload.Ranges.SheetId = 0
	payload.Ranges.StartRowIndex = 0
	payload.Ranges.EndRowIndex = 1
	payload.Ranges.StartColumnIndex = 0
	payload.Ranges.EndColumnIndex = 1
	payload.BooleanRule.Format.TextFormat.Strikethrough = true
	payload.BooleanRule.Condition.Type = "NOT_BLANK"

	//b, err := json.Marshal(c)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return
	}


	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	spreadsheetId := "1E0Zs4lrPkJuFVk0G3j7p0DIMnxIPYoz_1qE5nyliMS4"

	url := "https://sheets.googleapis.com/v4/spreadsheets/" + spreadsheetId + ":batchUpdate"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(jsonPayload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}