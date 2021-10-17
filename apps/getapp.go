package apps

import (
	"fmt"
	"io/ioutil"
)

func GetAppSettings(url string) (code string, settings Settings, err error) {
	_, err = runlua(RunluaParams{
		AppID:            url,
		ExtractedGlobals: &settings,
	})
	if err != nil {
		err = fmt.Errorf("failed to run lua: %w", err)
		return
	}

	if validationErr := settings.validate(); validationErr != nil {
		err = fmt.Errorf("app exported invalid definitions: %w", validationErr)
		return
	}

	return
}

func getAppCode(url string) (string, error) {
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

	return string(body), nil
}
