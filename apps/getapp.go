package apps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/rif/cache2go"
)

var CachedFailure = errors.New("cached-failure")

var settingsCache = cache2go.New(AppCacheSize, time.Minute*45)

func getCachedAppSettings(url string) *Settings {
	if settings, ok := settingsCache.Get(url); ok && settings != nil {
		return settings.(*Settings)
	}
	return nil
}

func GetAppSettings(url string) (*Settings, error) {
	if settings, ok := settingsCache.Get(url); ok {
		if settings == nil {
			return nil, CachedFailure
		} else {
			return settings.(*Settings), nil
		}
	}

	code, err := getAppCode(url)
	if err != nil {
		settingsCache.Set(url, nil)
		return nil, fmt.Errorf("failed to get app code: %w", err)
	}

	var settings Settings
	_, err = runlua(RunluaParams{
		Code:             code,
		AppURL:           url,
		ExtractedGlobals: &settings,
	})

	if err != nil {
		settingsCache.Set(url, nil)
		return nil, fmt.Errorf("failed to run lua: %w", err)
	}

	if validationErr := settings.validate(); validationErr != nil {
		settingsCache.Set(url, nil)
		return nil, fmt.Errorf("app exported invalid definitions: %w", validationErr)
	}

	// also add the code and URL
	settings.Code = code
	settings.URL = url

	// transform data
	settings.normalize()

	if AppCacheSize > 0 {
		settingsCache.Set(url, &settings)
	}

	return &settings, nil
}

var codeCache = cache2go.New(AppCacheSize/3, time.Minute*45)

func getAppCode(url string) (string, error) {
	if code, ok := codeCache.Get(url); ok {
		if code == nil {
			return "", CachedFailure
		} else {
			return code.(string), nil
		}
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		codeCache.Set(url, nil)
		return "", fmt.Errorf("http call errored: %w", err)
	}

	if resp.StatusCode >= 300 {
		codeCache.Set(url, nil)
		return "", fmt.Errorf("http call returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		codeCache.Set(url, nil)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	code := string(body)
	if AppCacheSize > 0 {
		codeCache.Set(url, code)
	}

	return code, nil
}
