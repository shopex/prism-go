package prism

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	Client        http.Client
	Key           string
	Server        string
	OAuthToken    string
	AlwaysUseSign bool
	server_url    *url.URL
	secret        string
}

type Response struct {
	Raw []byte
}

func NewClient(server, key, secret string) (c *Client, err error) {
	c = &Client{
		Key:    key,
		Server: server,
		secret: secret,
	}
	c.server_url, err = url.Parse(server)
	return
}

func (r *Response) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.Raw, v)
}

func (c *Client) Get(api string, params *map[string]interface{}) (rsp *Response, err error) {
	return c.do("GET", api, params)
}

func (c *Client) Post(api string, params *map[string]interface{}) (rsp *Response, err error) {
	return c.do("POST", api, params)
}

func (c *Client) do(method, api string, params *map[string]interface{}) (rsp *Response, err error) {
	r, err := c.get_request(method, api, params)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(r)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	return &Response{data}, err
}

func (c *Client) get_request(method, api string, params *map[string]interface{}) (req *http.Request, err error) {
	vals := url.Values{}

	if params != nil {
		for k, v := range *params {
			vals.Set(k, param_to_str(v))
		}
	}

	r, err := http.NewRequest(method, c.Server+"/"+api, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "Prism/Go")
	if c.OAuthToken != "" {
		r.Header.Set("Authorization", "Bearer "+c.OAuthToken)
	}

	use_url_query := method != "POST"

	vals.Set("client_id", c.Key)
	if !c.AlwaysUseSign && r.URL.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DisableCompression: true,
		}
		c.Client.Transport = tr

		vals.Set("grant_type", "password")
		vals.Set("client_secret", c.secret)
	} else {
		vals.Set("sign_time", strconv.FormatInt(time.Now().Unix(), 10))
		if use_url_query {
			r.URL.RawQuery = vals.Encode()
		} else {
			r.PostForm = vals
		}
		vals.Set("sign", Sign(r, c.secret))
	}

	query_string := vals.Encode()

	if use_url_query {
		r.URL.RawQuery = query_string
	} else {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ContentLength = int64(len(query_string))
		r.Body = &closebuf{bytes.NewBufferString(query_string)}
	}

	return r, nil
}

func param_to_str(v interface{}) (v2 string) {
	switch v.(type) {
	case string:
		v2 = v.(string)
	default:
		buf, _ := json.Marshal(v)
		v2 = string(buf)
	}
	return
}

type closebuf struct {
	*bytes.Buffer
}

func (cb *closebuf) Close() error {
	return nil
}
