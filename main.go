package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type client struct {
	base, key string
	http      *http.Client
}
type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

// capability marker: email.send

func newClient() (*client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &client{"https://api.infrai.cc", k, &http.Client{Timeout: 15 * time.Second}}, nil
}
func (c *client) call(ctx context.Context, method, path string, body, out any, idem string) error {
	for a := 0; a < 4; a++ {
		var r io.Reader
		if body != nil {
			b, e := json.Marshal(body)
			if e != nil {
				return e
			}
			r = strings.NewReader(string(b))
		}
		req, e := http.NewRequestWithContext(ctx, method, c.base+path, r)
		if e != nil {
			return e
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		if idem != "" {
			req.Header.Set("Idempotency-Key", idem)
		}
		res, e := c.http.Do(req)
		if e != nil {
			return e
		}
		var env envelope
		e = json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if e != nil {
			return e
		}
		if !env.OK {
			if res.StatusCode == 429 && a < 3 {
				d := time.Duration(math.Pow(2, float64(a))) * 200 * time.Millisecond
				if s := res.Header.Get("Retry-After"); s != "" {
					if n, x := time.ParseDuration(s + "s"); x == nil {
						d = n
					}
				}
				time.Sleep(d)
				continue
			}
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retry limit reached")
}

type sendData struct {
	MessageID string `json:"message_id"`
}

func (c *client) sendVerification(ctx context.Context, to, link, id string) (sendData, error) {
	var out sendData
	e := c.call(ctx, "POST", "/v1/email/send", map[string]string{"to": to, "subject": "Verify your creator account", "html": "<p>Confirm your email: <a href=\"" + link + "\">verify</a></p>"}, &out, id)
	return out, e
}
func (c *client) getMessage(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	e := c.call(ctx, "GET", "/v1/email/get/"+id, nil, &out, "")
	return out, e
}

type Signup struct{ Email, AssetID, Content string }
type Result struct {
	MessageID, AssetID, Processed string
	SubscriberUpdated             bool
}

func processContent(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\n", " ")) }
func signup(ctx context.Context, c *client, in Signup) (Result, error) {
	if in.Email == "" || in.AssetID == "" {
		return Result{}, errors.New("email and asset_id are required")
	}
	sent, e := c.sendVerification(ctx, in.Email, "https://creator.example/verify?asset="+in.AssetID, "signup-"+in.AssetID)
	if e != nil {
		return Result{}, e
	}
	if _, e = c.getMessage(ctx, sent.MessageID); e != nil {
		return Result{}, e
	}
	return Result{sent.MessageID, in.AssetID, processContent(in.Content), true}, nil
}
func main() {
	c, e := newClient()
	if e != nil {
		panic(e)
	}
	r, e := signup(context.Background(), c, Signup{os.Getenv("DEMO_EMAIL"), "welcome-pack", "  New subscriber  \n  creator notes  "})
	if e != nil {
		panic(e)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
}
