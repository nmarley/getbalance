package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"gopkg.in/yaml.v2"
)

// InsightAddressResponse is the format of a response from the Insight-API
// addr endpoint.
type InsightAddressResponse struct {
	Address                    string   `json:"addrStr"`
	Balance                    float64  `json:"balance"`
	BalanceSatoshis            int64    `json:"balanceSat"`
	TotalReceived              float64  `json:"totalReceived"`
	TotalReceivedSatoshis      int64    `json:"totalReceivedSat"`
	TotalSent                  float64  `json:"totalSent"`
	TotalSentSatoshis          int64    `json:"totalSentSat"`
	UnconfirmedBalance         float64  `json:"unconfirmedBalance"`
	UnconfirmedBalanceSatoshis int64    `json:"unconfirmedBalanceSat"`
	UnconfirmedTxApperances    int      `json:"unconfirmedTxApperances"`
	TxApperances               int      `json:"txApperances"`
	Transactions               []string `json:"transactions"`
}

// mainnetInsightAPI is the URL for the mainnet Insight deployment
const mainnetInsightAPI = "https://mainnet-insight.dashevo.org/insight-api"

// fetchBalance gets a Dash address balance via querying the official Insight
// deployment. Currently mainnet only.
func fetchBalance(addr string) (float64, error) {
	// TODO: template snippet
	// The full path to the address endpoint
	addrEndpoint := mainnetInsightAPI + "/addr/" + addr

	resp, err := http.Get(addrEndpoint)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	target := &InsightAddressResponse{}
	err = json.NewDecoder(resp.Body).Decode(target)

	return target.Balance, err
}

// AddrEntry is used to unmarshal YAML address entries.
type AddrEntry struct {
	Label   string `yaml:"label"`
	Address string `yaml:"addr"`
}

func main() {
	fn := "addresses.yaml"
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	var entries []AddrEntry
	err = yaml.Unmarshal(data, &entries)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	// Bound concurrent goroutines at about 8 or so
	guard := make(chan struct{}, 8)
	defer close(guard)

	total := float64(0)
	for _, e := range entries {
		wg.Add(1)
		guard <- struct{}{}

		go func(entry AddrEntry) {
			defer wg.Done()
			bal, err := fetchBalance(entry.Address)
			if err != nil {
				panic(err)
			}
			total += bal
			fmt.Printf("Balance for %v (%v) : %v\n", entry.Label, entry.Address, bal)
			<-guard
		}(e)
	}

	wg.Wait()
	fmt.Printf("Total: %v\n", total)
}
