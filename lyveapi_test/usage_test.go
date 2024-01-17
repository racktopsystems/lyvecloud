package lyveapi_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/racktopsystems/lyvecloud/lyveapi"
	"github.com/steinfletcher/apitest"
)

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

func Test_UsageBytesUsedCombined(t *testing.T) {
	t.Parallel()

	const expectedUsageBytes = 77398902000000
	usage := lyveapi.Buckets{
		lyveapi.Bucket{
			Name:    "a",
			UsageGB: 493.52,
		},
		lyveapi.Bucket{
			Name:    "b",
			UsageGB: 34.67,
		},
		lyveapi.Bucket{
			Name:    "c",
			UsageGB: 67245.034,
		},
		lyveapi.Bucket{
			Name:    "d",
			UsageGB: 9625.678,
		},
	}
	if usage.BytesUsedCombined() != expectedUsageBytes {
		t.Errorf("Unexpected usage; expected: %d, actual: %d", expectedUsageBytes, usage.BytesUsedCombined())
	}
}

func Test_UsageBytesUsedCombinedRangeLimit(t *testing.T) {
	t.Parallel()

	// Upper limit of what we can support is 'math.MaxUint64'.
	const upperLimit uint64 = math.MaxUint64
	const lowerAcceptableLimit uint64 = 1 << 63

	const expectedUsageLimit = upperLimit
	usage := lyveapi.Buckets{
		lyveapi.Bucket{
			Name:    "a",
			UsageGB: float64(upperLimit),
		},
		lyveapi.Bucket{
			Name:    "b",
			UsageGB: float64(upperLimit),
		},
		lyveapi.Bucket{
			Name:    "c",
			UsageGB: float64(upperLimit),
		},
	}

	if usage.BytesUsedCombined() < lowerAcceptableLimit {
		t.Errorf("Expected combined usage to support at least: %d bytes, actual: %d",
			lowerAcceptableLimit, usage.BytesUsedCombined())
	}
}

func Test_UsagesMonthlyTotalGB(t *testing.T) {
	t.Parallel()

	var expectedUsagesByMonth = map[lyveapi.MonthYearTuple]float64{
		{Month: 8, Year: 2023}:  36466.3927,
		{Month: 9, Year: 2023}:  36498.950999999994,
		{Month: 10, Year: 2023}: 36676.30099999999,
		{Month: 11, Year: 2023}: 34824.49,
	}
	const encodedUsageByBucket = `{
		"usageByBucket": [
		  {
			"buckets": [
			  {
				"name": "alpha",
				"usageGB": 534.52
			  },
			  {
				"name": "beta",
				"usageGB": 34567.2827
			  },
			  {
				"name": "gamma",
				"usageGB": 516.24
			  },
			  {
				"name": "delta",
				"usageGB": 848.35
			  }
			],
			"month": 8,
			"totalUsageGB": 36466.3927,
			"year": 2023
		  },
		  {
			"buckets": [
			  {
				"name": "alpha",
				"usageGB": 534.52
			  },
			  {
				"name": "beta",
				"usageGB": 34575.261
			  },
			  {
				"name": "gamma",
				"usageGB": 516.24
			  },
			  {
				"name": "delta",
				"usageGB": 872.93
			  }
			],
			"month": 9,
			"totalUsageGB": 36498.950999999994,
			"year": 2023
		  },
		  {
			"buckets": [
			  {
				"name": "alpha",
				"usageGB": 534.52
			  },
			  {
				"name": "beta",
				"usageGB": 34575.261
			  },
			  {
				"name": "gamma",
				"usageGB": 693.59
			  },
			  {
				"name": "delta",
				"usageGB": 872.93
			  }
			],
			"month": 10,
			"totalUsageGB": 36676.30099999999,
			"year": 2023
		  },
		  {
			"buckets": [
			  {
				"name": "alpha",
				"usageGB": 534.52
			  },
			  {
				"name": "beta",
				"usageGB": 32723.45
			  },
			  {
				"name": "gamma",
				"usageGB": 693.59
			  },
			  {
				"name": "delta",
				"usageGB": 872.93
			  }
			],
			"month": 11,
			"totalUsageGB": 34824.49,
			"year": 2023
		  }
		]
	  }`

	resp := &lyveapi.MonthlyUsageResp{}
	if err := json.NewDecoder(strings.NewReader(encodedUsageByBucket)).Decode(resp); err != nil {
		t.Fatal(err)
	}

	for k, v := range resp.UsageByBucket.MonthlyTotalUsageGB() {
		if expectedUsagesByMonth[k] != v {
			t.Errorf("Unable to find expected usage %v => %v", k, v)
		}
	}
}
