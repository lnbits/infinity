package apps

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/rif/cache2go"
)

var settingsCache = cache2go.New(AppCacheSize, time.Minute*45)

func GetAppSettings(url string) (*Settings, error) {
	if settings, ok := settingsCache.Get(url); ok {
		return settings.(*Settings), nil
	}

	code, err := getAppCode(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get app code: %w", err)
	}

	var settings Settings
	_, err = runlua(RunluaParams{
		Code:             code,
		AppURL:           url,
		ExtractedGlobals: &settings,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run lua: %w", err)
	}

	if validationErr := settings.validate(); validationErr != nil {
		return nil, fmt.Errorf("app exported invalid definitions: %w", validationErr)
	}

	// transform data
	settings.normalize()

	// also add the code
	settings.Code = code

	if AppCacheSize > 0 {
		settingsCache.Set(url, &settings)
	}

	return &settings, nil
}

var codeCache = cache2go.New(AppCacheSize/2, time.Minute*45)

func getAppCode(url string) (string, error) {
	if code, ok := codeCache.Get(url); ok {
		return code.(string), nil
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("http call errored: %w", err)
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("http call returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	code := string(body)

	if AppCacheSize > 0 {
		codeCache.Set(url, code)
	}

	return code, nil
}
