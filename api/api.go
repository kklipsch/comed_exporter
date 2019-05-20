package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type (
	response struct {
		MillisUTC string `json:"millisUTC"`
		Price     string `json:"price"`
	}

	today []response

	Price struct {
		AsOf        time.Time
		CentsPerKWh float64
	}
)

const Address = "https://hourlypricing.comed.com/api"

func GetLastPrice(client *http.Client, address string) (Price, error) {
	resp, err := client.Get(fmt.Sprintf("%s?type=5minutefeed", address))
	if err != nil {
		return Price{}, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Price{}, fmt.Errorf("cannot read body: %v", err)
	}

	var t today
	err = json.Unmarshal(b, &t)
	if err != nil {
		return Price{}, fmt.Errorf("unable to unmarshal response: %v\n%s", err, b)
	}

	if len(t) < 1 {
		return Price{}, fmt.Errorf("no prices available: %s", b)
	}

	last := t[len(t)-1]

	millis, err := strconv.ParseInt(last.MillisUTC, 10, 64)
	if err != nil {
		return Price{}, fmt.Errorf("bad milli format: %v\n%s", last, b)
	}

	p, err := strconv.ParseFloat(last.Price, 2)
	if err != nil {
		return Price{}, fmt.Errorf("bad price format: %v\n%s", last, b)
	}

	return Price{AsOf: time.Unix(0, millis*1000*1000), CentsPerKWh: p}, nil
}
