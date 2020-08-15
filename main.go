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
	AssetClass                          string `json:"assetClass"`
	Benchmark                           string `json:"benchmark"`
	BenchmarkNameFromECS                string `json:"benchmarkNameFromECS"`
	CloseIndicator                      string `json:"closeIndicator"`
	CurrencyCode                        string `json:"currencyCode"`
	CurrencySymbol                      string `json:"currencySymbol"`
	CutOffTime                          string `json:"cutOffTime"`
	DisplayName                         string `json:"displayName"`
	DistributionStrategyType            string `json:"distributionStrategyType"`
	DistributionStrategyTypeDescription string `json:"distributionStrategyTypeDescription"`
	ExtendedFundType                    string `json:"extendedFundType"`
	FundType                            string `json:"fundType"`
	Id                                  string `json:"id"`
	InceptionDate                       string `json:"inceptionDate"`
	InvestmentStrategyDescription       string `json:"investmentStrategyDescription"`
	IsGlobalBalanced                    string `json:"isGlobalBalanced"`
	Isin                                string `json:"isin"`
	IssueType                           string `json:"issueType"`
	ManagementType                      string `json:"managementType"`
	Name                                string `json:"Name"`
	OCF                                 string `json:"OCF"`
	ParentPortId                        string `json:"parentPortId"`
	PortId                              string `json:"PortId"`
	PurchaseFee                         string `json:"purchaseFee"`
	RedemptionFee                       string `json:"redemptionFee"`
	RetailDirectAvailability            string `json:"retailDirectAvailability"`
	Sedol                               string `json:"sedol"`
	ShareClassCode                      string `json:"shareClassCode"`
	ShareClassCodeDescription           string `json:"shareClassCodeDescription"`
	ShareclassCode                      string `json:"shareclassCode"`
	ShareclassDescription               string `json:"shareclassDescription"`
	StampDutyReserveTax                 string `json:"stampDutyReserveTax"`
	ValidityCode                        string `json:"validityCode"`
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

func getFundPriceHandler(rw http.ResponseWriter, r *http.Request) {
	DayDiff := -7
	today := time.Now().Format(TIME_FORMAT)
	start := time.Now().AddDate(0, 0, DayDiff).Format(TIME_FORMAT)
	validPath := regexp.MustCompile("^/fundprice/([0-9]+)$")
	portId := validPath.FindStringSubmatch(r.URL.Path)[1]

	fund := Fund{
		PortId: portId,
	}
	fundUrl := fund.Url(start, today)
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
		response = response + fmt.Sprintf("<a href=fundprice/%s>%s %s</a></br>", fund.PortId, fund.Name, fund.ShareClassCodeDescription)
	}
	rw.Write([]byte(response))
}

func main() {

	http.HandleFunc("/fundprice/", getFundPriceHandler)
	http.HandleFunc("/fundslist", getFundsListHandler)
	http.Handle("/", http.RedirectHandler("/fundslist", 302))
	http.ListenAndServe(":8080", nil)
}
