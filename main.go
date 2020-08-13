package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

/*
TODO:
- update fund names and IDs from website
- write fund info to DB
- plot graph
- select funds to compare
- login with personalised default fund
*/

const (
	TIME_FORMAT = "2006-01-02"
)

type Price struct {
	Date     string  `json:"date"`
	NavPrice float32 `json:"navPrice"`
}

func (p Price) IsoDate() string {
	return strings.Split(p.Date, "T")[0]
}

type PriceHistory []Price

type Fund struct {
	Name   string `json:"Name"`
	PortId string `json:"PortId"`
}

func (f Fund) Url(start string, today string) string {
	return fmt.Sprintf("https://api.vanguard.com/rs/gre/gra/1.7.0/datasets/urd-product-port-specific-price-history.jsonp?vars=portId:%s,issueType:S,startDate:%s,endDate:%s&callback=angular.callbacks._c", f.PortId, start, today)
}

func callVanguardApi(Url string) string {
	// Download fund history and convert to store in variable
	resp, err := http.Get(Url)
	if err != nil {
		log.Fatal(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	// Make into parsable JSON by removing angular related strings
	var re = regexp.MustCompile(`^angular.callbacks._[A-Za-z0-9]*\(`)
	b := re.ReplaceAllString(string(body), "")
	var re2 = regexp.MustCompile(`\)$`)
	b = re2.ReplaceAllString(b, "")

	return b
}

// TODO
// - Accept date ranges
// - select relative date range
func getFundData(fundUrl string) string {

	b := callVanguardApi(fundUrl)

	// Read response from API into a JSON object
	var priceHistory PriceHistory
	json.Unmarshal([]byte(b), &priceHistory)

	// Print daily prices
	var prices string
	for _, price := range priceHistory {
		prices = prices + fmt.Sprintf("%s: £%.2f\n", price.IsoDate(), price.NavPrice)
	}
	return prices
}

func getFundDataHandler(rw http.ResponseWriter, r *http.Request) {
	DayDiff := -7
	today := time.Now().Format(TIME_FORMAT)
	start := time.Now().AddDate(0, 0, DayDiff).Format(TIME_FORMAT)

	ftse_global_all_cap := Fund{
		PortId: "8617",
	}
	fundUrl := ftse_global_all_cap.Url(start, today)
	rw.Write([]byte(getFundData(fundUrl)))
}

func getFundsList() []byte {
	fundsListUrl := "https://api.vanguard.com/rs/gre/gra/1.7.0/datasets/urd-identifiers.jsonp?callback=angular.callbacks._0"
	fundsList := callVanguardApi(fundsListUrl)

	return []byte(fundsList)
}

func getFundsListHandler(rw http.ResponseWriter, r *http.Request) {
	fundsList := getFundsList()
	var FundsList []Fund

	json.Unmarshal(fundsList, &FundsList)
	var response string

	for _, fund := range FundsList {
		response = response + fmt.Sprintf("%s: %s\n", fund.Name, fund.PortId)
	}
	rw.Write([]byte(response))
}

func main() {

	http.HandleFunc("/fgac", getFundDataHandler)
	http.HandleFunc("/fundslist", getFundsListHandler)
	http.ListenAndServe(":8080", nil)
}
