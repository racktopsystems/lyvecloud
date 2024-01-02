package lyveapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Month represents months in a year in their numeric form
type Month uint8

const (
	JAN Month = iota
	FEB
	MAR
	APR
	MAY
	JUN
	JUL
	AUG
	SEP
	OCT
	NOV
	DEC
)

func (m Month) String() string {
	return map[Month]string{
		JAN: "0",
		FEB: "1",
		MAR: "2",
		APR: "3",
		MAY: "4",
		JUN: "5",
		JUL: "6",
		AUG: "7",
		SEP: "8",
		OCT: "9",
		NOV: "10",
		DEC: "11",
	}[m]
}

type Year uint16

type MonthYearTuple struct {
	Month Month
	Year  Year
}

// GetMonthlyUsage retrieves historical data, in monthly increments
// within the provided range of time. This range is limited to a maximum of six
// (6) months and a query spanning a larger range will fail with an
// InvalidTimeRange error. This validation happens on the Lyve Cloud side. If
// the information is queried with a sub-account, then no information is
// returned in the UsageBySubAccount field, since sub-accounts cannot have
// their own sub-accounts.
func (client *Client) GetMonthlyUsage(
	fromMonth Month,
	fromYear uint,
	toMonth Month,
	toYear uint,
) (*MonthlyUsageResp, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/usage/monthly"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	// Build our query string with the supplied parameters
	params := url.Values{}
	params.Set("fromMonth", fromMonth.String())
	params.Set("fromYear", fmt.Sprintf("%d", fromYear))
	params.Set("toMonth", toMonth.String())
	params.Set("toYear", fmt.Sprintf("%d", toYear))

	url := endpoint + "?" + params.Encode()

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, url, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	respBodyDecoder := json.NewDecoder(rdr)

	usageResp := &MonthlyUsageResp{}

	if err := respBodyDecoder.Decode(usageResp); err != nil {
		return nil, err
	}

	return usageResp, nil
}

// GetCurrentUsage retrieves current bucket usage data across all buckets under
// the sub-account. If the information is queried with a sub-account, then
// no information is returned in the UsageBySubAccount field, since sub-accounts
// cannot have their own sub-accounts.
func (client *Client) GetCurrentUsage() (*CurrentUsageResp, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/usage/current"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, endpoint, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	respBodyDecoder := json.NewDecoder(rdr)

	usageResp := &CurrentUsageResp{}

	if err := respBodyDecoder.Decode(usageResp); err != nil {
		return nil, err
	}

	return usageResp, nil
}
