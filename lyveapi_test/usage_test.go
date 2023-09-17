package lyveapi_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/szaydel/lyvecloud/lyveapi"
)

// var monthFromNumericString = map[string]lyveapi.Month{
// 	"0":  lyveapi.INVALID,
// 	"1":  lyveapi.JAN,
// 	"2":  lyveapi.FEB,
// 	"3":  lyveapi.MAR,
// 	"4":  lyveapi.APR,
// 	"5":  lyveapi.MAY,
// 	"6":  lyveapi.JUN,
// 	"7":  lyveapi.JUL,
// 	"8":  lyveapi.AUG,
// 	"9":  lyveapi.SEP,
// 	"10": lyveapi.OCT,
// 	"11": lyveapi.NOV,
// 	"12": lyveapi.DEC,
// }

func TestUsage(t *testing.T) {
	const currentUsageRespBody = `{
		"usageByBucket": {
			"numBuckets": 5,
			"totalUsageGB": 24032.63,
			"buckets": [
				{
					"name": "alpha",
					"usageGB": 0
				},
				{
					"name": "beta",
					"usageGB": 0
				},
				{
					"name": "gamma",
					"usageGB": 534.51
				},
				{
					"name": "delta",
					"usageGB": 3957.45
				},
				{
					"name": "epsilon",
					"usageGB": 19540.67
				}
			]
		}
	}`

	const currentUsageDecoded = `{
		"usageByBucket": {
			"buckets": [
				{
					"name": "alpha",
					"usageGB": 0
				},
				{
					"name": "beta",
					"usageGB": 0
				},
				{
					"name": "gamma",
					"usageGB": 534.51
				},
				{
					"name": "delta",
					"usageGB": 3957.45
				},
				{
					"name": "epsilon",
					"usageGB": 19540.67
				}
			],
			"numBuckets": 5,
			"totalUsageGB": 24032.63
		},
			"usageBySubAccount": {
			"totalUsageGB": 0
		}
	}`

	const monthlyUsageRespBody = `{
		"usageByBucket": [
			{
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 0
					},
					{
						"name": "beta",
						"usageGB": 0
					},
					{
						"name": "gamma",
						"usageGB": 0
					}
				],
				"month": 3,
				"totalUsageGB": 0,
				"year": 2023
			},
			{
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 23.20
					},
					{
						"name": "beta",
						"usageGB": 8.56
					},
					{
						"name": "gamma",
						"usageGB": 3.92
					}
				],
				"month": 4,
				"totalUsageGB": 35.68,
				"year": 2023
			},
			{
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 23.32
					},
					{
						"name": "beta",
						"usageGB": 14.27
					},
					{
						"name": "gamma",
						"usageGB": 4.67
					},
					{
						"name": "delta",
						"usageGB": 56.78
					}
	
				],
				"month": 5,
				"totalUsageGB": 99.04,
				"year": 2023
			}
		]
	}`

	const monthlyUsageDecoded = `{
		"usageByBucket": [
			{
				"year": 2023,
				"month": 3,
				"totalUsageGB": 0,
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 0
					},
					{
						"name": "beta",
						"usageGB": 0
					},
					{
						"name": "gamma",
						"usageGB": 0
					}
				]
			},
			{
				"year": 2023,
				"month": 4,
				"totalUsageGB": 35.68,
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 23.2
					},
					{
						"name": "beta",
						"usageGB": 8.56
					},
					{
						"name": "gamma",
						"usageGB": 3.92
					}
				]
			},
			{
				"year": 2023,
				"month": 5,
				"totalUsageGB": 99.04,
				"buckets": [
					{
						"name": "alpha",
						"usageGB": 23.32
					},
					{
						"name": "beta",
						"usageGB": 14.27
					},
					{
						"name": "gamma",
						"usageGB": 4.67
					},
					{
						"name": "delta",
						"usageGB": 56.78
					}
				]
			}
		]
	}`

	var currentUsageMock = apitest.NewMock().
		Get(mockCurrentUsageUri).
		RespondWith().
		Body(currentUsageRespBody).
		Status(http.StatusOK).
		End()

	var monthlyUsageQueryParams = map[string]string{
		"fromMonth": lyveapi.MAR.String(),
		"fromYear":  "2023",
		"toMonth":   lyveapi.MAY.String(),
		"toYear":    "2023"}

	var monthlyUsageMock = apitest.NewMock().
		Get(mockMonthlyUsageUri).
		QueryParams(monthlyUsageQueryParams).
		RespondWith().
		Body(monthlyUsageRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(currentUsageMock).
		Handler(usageHandler()).
		Get(mockCurrentUsageUri).
		Expect(t).
		Body(currentUsageDecoded).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(monthlyUsageMock).
		Handler(usageHandler()).
		Get(mockMonthlyUsageUri).
		QueryParams(monthlyUsageQueryParams).
		Expect(t).
		Body(monthlyUsageDecoded).
		Status(http.StatusOK).
		End()
}

func usageHandler() *http.ServeMux {
	var handler = http.NewServeMux()

	handler.HandleFunc(mockCurrentUsageUri, func(w http.ResponseWriter, r *http.Request) {
		var currentUsage lyveapi.CurrentUsageResp

		switch r.Method {
		case http.MethodGet:
			if err := usageGet(mockCurrentUsageUri, nil, &currentUsage); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			bytes, _ := json.Marshal(currentUsage)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	handler.HandleFunc(mockMonthlyUsageUri, func(w http.ResponseWriter, r *http.Request) {
		var monthlyUsage lyveapi.MonthlyUsageResp

		log.Print(r)
		switch r.Method {
		case http.MethodGet:
			if err := usageGet(mockMonthlyUsageUri, r.URL, &monthlyUsage); err != nil {
				errResp := map[string]interface{}{
					"error": err.Error(),
				}
				bytes, _ := json.Marshal(errResp)
				w.Write(bytes)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			bytes, _ := json.Marshal(monthlyUsage)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	return handler
}

func usageGet(path string, url *url.URL, response interface{}) error {
	var err error
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockCurrentUsageUri:
		var u *lyveapi.CurrentUsageResp
		if u, err = client.GetCurrentUsage(); err != nil {
			return err
		}
		response.(*lyveapi.CurrentUsageResp).
			UsageByBucket = u.UsageByBucket
		response.(*lyveapi.CurrentUsageResp).
			UsageBySubAccount = u.UsageBySubAccount
		return nil

	case mockMonthlyUsageUri:
		var u *lyveapi.MonthlyUsageResp

		fromMonth, _ := strconv.Atoi(url.Query().Get("fromMonth"))
		fromYear, _ := strconv.Atoi(url.Query().Get("fromYear"))
		toMonth, _ := strconv.Atoi(url.Query().Get("toMonth"))
		toYear, _ := strconv.Atoi(url.Query().Get("toYear"))

		fm := lyveapi.Month(fromMonth)
		fy := uint(fromYear)
		tm := lyveapi.Month(toMonth)
		ty := uint(toYear)

		if u, err = client.GetMonthlyUsage(fm, fy, tm, ty); err != nil {
			return err
		}
		response.(*lyveapi.MonthlyUsageResp).
			UsageByBucket = u.UsageByBucket
		response.(*lyveapi.MonthlyUsageResp).
			UsageBySubAccount = u.UsageBySubAccount
		return nil
	}

	return pathUnmatchedErr
}
