package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type (
	symbolPriceResponse struct {
		Status string                `json:"status"`
		Stats  map[string]PriceStats `json:"Stats"`
	}
	PriceStats struct {
		BestSell  string `json:"bestSell"`
		DayOpen   string `json:"dayOpen"`
		DayHigh   string `json:"dayHigh"`
		BestBuy   string `json:"bestBuy"`
		VolumeSrc string `json:"volumeSrc"`
		DayLow    string `json:"dayLow"`
		Latest    string `json:"latest"`
		VolumeDst string `json:"volumeDst"`
		DayChange string `json:"dayChange"`
		DayClose  string `json:"dayClose"`
		IsClosed  bool   `json:"IsClosed"`
	}
	httpClient struct {
		baseUrl string
		apiKey  string
		cli     *http.Client
	}
)

func newClientHttp(url, apiKey string, timeout time.Duration) *httpClient {
	return &httpClient{
		baseUrl: url,
		cli:     &http.Client{Timeout: timeout * time.Second},
		apiKey:  apiKey,
	}
}

func (c *httpClient) tokenPrice(srcCurrency string, dstCurrency string) (symbolPriceResponse, error) {

	u := fmt.Sprintf("%s/market/stats", c.baseUrl)
	fmt.Printf("%s\n", u)
	body, err := json.Marshal(map[string]interface{}{
		"srcCurrency": srcCurrency,
		"dstCurrency": dstCurrency,
	})
	var v symbolPriceResponse
	err = c.do(http.MethodPost, u, c.authHeader(), bytes.NewReader(body), &v)
	if err != nil {
		return v, err
	}

	return v, nil
}
func (c *httpClient) do(method, url string, header http.Header, body io.Reader, v interface{}) error {
	req, err := http.NewRequest(method, url, body)
	fmt.Println(header)
	req.Header = header
	if err != nil {
		return err
	}
	res, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println("Response Status Code:", res.StatusCode)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println("Response Body:", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return err
	}
	return nil
}
func (c *httpClient) authHeader() http.Header {
	header := http.Header{}
	header.Set("Authorization", "Currency "+c.apiKey)
	header.Add("content-type", "application/json")
	return header
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan symbolPriceResponse)
	httpClient := newClientHttp("https://api.nobitex.ir", "5f77fb19b6f633ff1d9ceab185f27c8fe5611ec6", 30)
	tokenSymbols := []string{"btc", "ETH", "LTC", "XRP", "ADA"}

	// resp, err := httpClient.tokenPrice("btc", "rls")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// fmt.Println(resp)

	for _, symbol := range tokenSymbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			resp, err := httpClient.tokenPrice(sym, "rls")
			if err != nil {
				fmt.Println("Error:", err)
			}
			ch <- resp
		}(symbol)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for resp := range ch {
		// process the response
		fmt.Println(resp.Stats)
	}
}
