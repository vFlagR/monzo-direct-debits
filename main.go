package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	Rule `json:"rule,"`
}

type Rule struct {
	Ranges      `json:"ranges,"`
	BooleanRule `json:"booleanRule,"`
}

type Ranges struct {
	SheetId          int `json:"sheetId"`
	StartRowIndex    int `json:"startRowIndex"`
	EndRowIndex      int `json:"endRowIndex"`
	StartColumnIndex int `json:"startColumnIndex"`
	EndColumnIndex   int `json:"endColumnIndex"`
}

type BooleanRule struct {
	Format struct {
		TextFormat struct {
			Strikethrough bool `json:"strikethrough"`
		} `json:"textFormat,"`
	} `json:"format,"`

	Condition struct {
		Type string `json:"type,"`
	} `json:"condition,"`
}

func main() {
	monthIndex := mapMonthToCell()
	debitIndex := mapDebitsToCell("d")

	payload := &FinalRequest{}
	payload.Ranges.SheetId = 0
	payload.Ranges.StartRowIndex = debitIndex      // Direct Debit Name Starting
	payload.Ranges.EndRowIndex = debitIndex + 1    // Direct Debit Name Ending
	payload.Ranges.StartColumnIndex = monthIndex   // Month of the year starting
	payload.Ranges.EndColumnIndex = monthIndex + 1 // Month of the year ending
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

	spreadsheetId := "185D2SRLy7huAeprPCvmZTSbOr0fH5oQ2Rdtn-9ODj4o"

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

func mapMonthToCell() int {
	months := map[string]int{
		"August":    1,
		"September": 2,
		"October":   3,
		"November":  4,
		"December":  5,
	}

	currentMonth := time.Now().Format("January")

	monthIndex := months[currentMonth]

	return monthIndex
}

func mapDebitsToCell(debitReference string) int {
	debits := map[string]int{
		"a": 2,  // Netflix
		"b": 3,  // Ikea
		"c": 4,  // Prime
		"d": 5,  // Barclays
		"e": 6,  // Usenet
		"f": 7,  // GSuite
		"g": 8,  // Vodafone
		"h": 9,  // PSN
		"i": 10, // Hertzner
		"j": 11, // AWS
		"k": 12, // O2
		"l": 14, // Tesco
		"m": 15, // BarclayCard
		"n": 16, // Train
		"o": 17, // Bus
	}

	if debitReference, ok := debits[debitReference]; ok {
		return debitReference
	}

	return 0

}
