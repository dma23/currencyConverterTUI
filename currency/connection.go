package currency

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type ApiResponse struct {
	Disclaimer   string
	License      string
	Timestamp    int64
	BaseCurrency string
	Rates        map[string]float64
}

var (
	cachedRates     map[string]float64
	cachedRatesTime time.Time
	cachedRatesLock sync.RWMutex
)

// limit API calls since im on the free version with limits, once every hour or used cached rates
// works assumiong the program is running forever, dont restart the prgoram too many times
func GetRates() (map[string]float64, error) {
	cachedRatesLock.RLock()
	valid := !cachedRatesTime.IsZero() && time.Since(cachedRatesTime) < 60*time.Minute
	cachedRatesLock.RUnlock()

	if valid {
		cachedRatesLock.RLock()
		defer cachedRatesLock.RUnlock()
		return cachedRates, nil
	}

	return GetLatestRates()
}

func GetLatestRates() (map[string]float64, error) {
	app := os.Getenv("API_KEY")
	if app == "" {
		return nil, fmt.Errorf("incorrect API Key")
	}

	url := fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s", app)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("response error")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status not ok")
	}

	var ResponseMessage ApiResponse

	cachedRatesLock.Lock()
	cachedRates = ResponseMessage.Rates
	cachedRatesTime = time.Now()
	cachedRatesLock.Unlock()

	return ResponseMessage.Rates, nil
}
