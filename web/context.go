package web

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
)

var (
	// RouteCtxKey is the context.Context key to store the request context.
	RouteCtxKey = &contextKey{"WebContext"}
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

type Context interface {
	Reset()
	PathParam(key string) string
	GetParam(key string) (string, error)
	GetParamArray(key string) ([]string, error)
	PostParam(key string) (string, error)
	PostParamArray(key string) ([]string, error)
	Cookie(key string) (*http.Cookie, error)
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	BodyJson(s interface{}) error
	SetResponse(s interface{})
}

type contextInline struct {
	Request       *http.Request
	responseWrite http.ResponseWriter
	context       *chi.Context
	Response      interface{}

	queryForm url.Values
}

func NewContext() Context {
	return &contextInline{}
}

func RouteContext(ctx context.Context) Context {
	return ctx.Value(RouteCtxKey).(Context)
}

func (c *contextInline) Reset() {
	c.queryForm = nil
}

func (c *contextInline) PathParam(key string) string {
	return c.context.URLParam(key)
}

func (c *contextInline) GetParam(key string) (string, error) {
	if vs, err := c.GetParamArray(key); err != nil {
		return "", fmt.Errorf("empty")
	} else {
		return vs[0], nil
	}
}

func (c *contextInline) GetParamArray(key string) ([]string, error) {
	if c.queryForm == nil {
		queryForm, err := url.ParseQuery(c.Request.URL.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("url Parse error:%s", err)
		} else {

			c.queryForm = queryForm
		}
	}
	if vs := c.queryForm["id"]; len(vs) > 0 {
		return vs, nil
	} else {
		return nil, fmt.Errorf("%s is exists", key)
	}
}

func (c *contextInline) PostParam(key string) (string, error) {
	if vs, err := c.PostParamArray(key); err != nil {
		return "", fmt.Errorf("empty")
	} else {
		return vs[0], nil
	}
}

func (c *contextInline) PostParamArray(key string) ([]string, error) {
	if c.Request.Form == nil {
		c.Request.ParseMultipartForm(defaultMaxMemory)
	}
	if vs := c.Request.Form[key]; len(vs) > 0 {
		return vs, nil
	} else {
		return nil, fmt.Errorf("%s is exists", key)
	}
}

func (c *contextInline) Cookie(key string) (*http.Cookie, error) {
	return c.Request.Cookie(key)
}

func (c *contextInline) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(key)
}

func (c *contextInline) BodyJson(s interface{}) error {
	if err := json.NewDecoder(c.Request.Body).Decode(s); err != nil {
		return err
	}
	return nil
}

func (c *contextInline) SetResponse(s interface{}) {
	c.Response = s
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "chi context value " + k.name
}
