package currency

func ConvertCurrencies(p1 float64, p2 string) float64 {

	price := p1 / Rates[p2]

	return price * Rates[p2]
}

func GetExchangeRates(p1, p2 string) float64 {

	return Rates[p1] / Rates[p2]
}
