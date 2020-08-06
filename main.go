package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

/*
TODO:
- update fund names and IDs from website
- write fund info to DB
- plot graph
- select funds to compare
- login with personalised default fund
*/

type Price struct {
	Date     string  `json:"date"`
	NavPrice float32 `json:"navPrice"`
}

func (p Price) IsoDate() string {
	return strings.Split(p.Date, "T")[0]
}

type PriceHistory []Price

type Fund struct {
	Url string
}

// TODO
// - Accept date ranges
// - select relative date range
func getFundData(fund Fund) string {
	// Download fund history and convert to store in variable
	resp, err := http.Get(fund.Url)
	if err != nil {
		log.Fatal(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	// Make into parsable JSON by removing angular related strings
	b := strings.Replace(string(body), "angular.callbacks._c(", "", 1)
	b = strings.Replace(b, ")", "", 1)

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
	ftse_global_all_cap := Fund{
		Url: "https://api.vanguard.com/rs/gre/gra/1.7.0/datasets/urd-product-port-specific-price-history.jsonp?vars=portId:8617,issueType:S,startDate:2020-07-22,endDate:2020-08-01&callback=angular.callbacks._c",
	}
	rw.Write([]byte(getFundData(ftse_global_all_cap)))
}

func main() {

	http.HandleFunc("/fgac", getFundDataHandler)
	http.ListenAndServe(":8080", nil)
}
