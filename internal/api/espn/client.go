package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/omarshaarawi/coachbot/internal/config"
)

const baseURL = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl"

type Client struct {
	httpClient *http.Client
	Config     config.ESPNAPI
}

func NewClient(cfg config.ESPNAPI) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		Config:     cfg,
	}
}

func (c *Client) Get(endpoint string, params, headers map[string]string, result interface{}) error {
	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	q := req.URL.Query()
	for key, value := range params {
		values := strings.Split(value, ",")
		for _, v := range values {
			q.Add(key, strings.TrimSpace(v))
		}
	}
	req.URL.RawQuery = q.Encode()

	c.setCookies(req)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}

func (c *Client) setCookies(req *http.Request) {
	cookie := fmt.Sprintf("SWID=%s; espn_s2=%s", c.Config.SWID, c.Config.ESPNS2)
	req.Header.Set("Cookie", cookie)
}
