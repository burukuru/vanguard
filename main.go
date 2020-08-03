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
- display values on web pge
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

type PriceHistory []Price

type Fund struct {
	Url string
}

// TODO
// - Accept date ranges
// - select relative date range
// - change date format to ISO
func getFundData(fund Fund) {
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
	for _, price := range priceHistory {
		fmt.Printf("%s: £%.2f\n", price.Date, price.NavPrice)
	}
}

func main() {
	ftse_global_all_cap := Fund{
		Url: "https://api.vanguard.com/rs/gre/gra/1.7.0/datasets/urd-product-port-specific-price-history.jsonp?vars=portId:8617,issueType:S,startDate:2020-07-22,endDate:2020-08-01&callback=angular.callbacks._c",
	}

	getFundData(ftse_global_all_cap)
}
