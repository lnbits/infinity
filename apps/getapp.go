package apps

import (
	"fmt"
	"io/ioutil"

	"github.com/asaskevich/govalidator"
	"github.com/lnbits/lnbits/apps/runlua"
)

func getAppSettings(url string) (code string, settings Settings, err error) {
	code, err = getAppCode(url)
	if err != nil {
		err = fmt.Errorf("failed to fetch app code from %s: %w", url, err)
		return
	}

	_, err = runlua.RunLua(runlua.Params{
		AppID:            url,
		AppCode:          code,
		ExtractedGlobals: &settings,
	})
	if err != nil {
		err = fmt.Errorf("failed to get settings from app code: %w", err)
		return
	}

	if valid, validationErr := govalidator.ValidateStruct(settings); !valid {
		err = fmt.Errorf("app returned invalid definitions: %w", validationErr)
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
