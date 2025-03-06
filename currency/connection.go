package currency

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type ApiResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

var (
	cachedRates     map[string]float64
	cachedRatesTime time.Time
	cachedRatesLock sync.RWMutex
)

func load_env() {
	// Try current directory first
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// If not found, check parent directory
		envFile = filepath.Join("..", ".env")
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			// .env file doesn't exist in either place
			return
		}
	}

	// Load the .env file that was found
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

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
	load_env()

	app := os.Getenv("API_KEY")
	if app == "" {
		return nil, fmt.Errorf("incorrect API Key")
	}

	url := fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s", app)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("response error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status not ok: %d", resp.StatusCode)
	}

	// Parse the JSON response
	var responseMessage ApiResponse
	err = json.NewDecoder(resp.Body).Decode(&responseMessage)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Update the cached rates
	cachedRatesLock.Lock()
	cachedRates = responseMessage.Rates
	cachedRatesTime = time.Now()
	cachedRatesLock.Unlock()

	return responseMessage.Rates, nil
}
