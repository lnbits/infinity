package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// should only be used inside lua code
func HTTP(
	method string,
	url string,
	data interface{},
	headers map[string]string,
) (interface{}, int, error) {
	var baseHeaders = map[string]string{
		"Accept": "application/json",
	}
	var body = &bytes.Buffer{}
	if data != nil {
		switch v := data.(type) {
		case string:
			body = bytes.NewBuffer([]byte(v))
			baseHeaders["Content-Type"] = "text/plain"
		case int, int64, float64:
			body = bytes.NewBuffer([]byte(fmt.Sprint(v)))
			baseHeaders["Content-Type"] = "text/plain"
		default:
			j, err := JSONMarshal(data)
			if err != nil {
				return nil, 0,
					fmt.Errorf("given data (%v) is not json serializable: %w",
						data, err)
			}
			body = bytes.NewBuffer(j)
			baseHeaders["Content-Type"] = "application/json"
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build http request to %s: %w", url, err)
	}

	for k, v := range baseHeaders {
		req.Header.Set(k, v)
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make http request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	var result interface{}
	b, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		if err := json.Unmarshal(b, &result); err != nil {
			result = string(b)
		}
	} else {
		result = map[string]interface{}{}
	}

	return result, resp.StatusCode, nil
}

func HTTPGet(url string) (interface{}, int, error) {
	return HTTP("GET", url, nil, nil)
}
func HTTPPost(url string, data interface{}) (interface{}, int, error) {
	return HTTP("POST", url, data, nil)
}
func HTTPPut(url string, data interface{}) (interface{}, int, error) {
	return HTTP("PUT", url, data, nil)
}
func HTTPPatch(url string, data interface{}) (interface{}, int, error) {
	return HTTP("PATCH", url, data, nil)
}
func HTTPDelete(url string) (interface{}, int, error) {
	return HTTP("DELETE", url, nil, nil)
}

func ParseQueryString(query string) (map[string]string, error) {
	parsed, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	qs := make(map[string]string)
	for k, vs := range parsed {
		qs[k] = vs[0]
	}

	return qs, nil
}

func EncodeQueryString(qs map[string]interface{}) string {
	values := url.Values{}

	for k, v := range qs {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	return values.Encode()
}
