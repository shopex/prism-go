package prism

import (
	"bytes"
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
	secret        string
}

type Response struct {
	Raw []byte
}

func NewClient(server, key, secret string) *Client {
	return &Client{
		Key:    key,
		Server: server,
		secret: secret,
	}
}

func (r *Response) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.Raw, v)
}

func (me *Client) Get(api string, params *map[string]interface{}) (result interface{}, err error) {
	return me.do("GET", api, params)
}

func (me *Client) Post(api string, params *map[string]interface{}) (result interface{}, err error) {
	return me.do("POST", api, params)
}

func (me *Client) do(method, api string, params *map[string]interface{}) (result interface{}, err error) {
	r, err := me.get_request(method, api, params)
	if err != nil {
		return nil, err
	}
	res, err := me.Client.Do(r)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	return Response{data}, err
}

func (me *Client) get_request(method, api string, params *map[string]interface{}) (req *http.Request, err error) {
	vals := url.Values{}

	if params != nil {
		for k, v := range *params {
			vals.Set(k, param_to_str(v))
		}
	}

	r, err := http.NewRequest(method, me.Server+"/"+api, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "Prism/Go")
	if me.OAuthToken != "" {
		r.Header.Set("Authorization", "Bearer "+me.OAuthToken)
	}

	use_url_query := method != "POST"

	vals.Set("client_id", me.Key)
	if !me.AlwaysUseSign && r.URL.Scheme == "https" {
		vals.Set("secret", me.secret)
	} else {
		vals.Set("sign_time", strconv.FormatInt(time.Now().Unix(), 10))
		if use_url_query {
			r.URL.RawQuery = vals.Encode()
		} else {
			r.PostForm = vals
		}
		vals.Set("sign", Sign(r, me.secret))
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
