package room

import (
	"net/http"
	"strings"
	"time"
)

const (
	headerKeyContentType         = "Content-Type"
	headerKeyAccept              = "Accept"
	headerValueFormEncoded       = "application/x-www-form-urlencoded"
	headerValueApplicationJson   = "application/json"
	headerValueTextXML           = "text/xml"
	headerValueMultipartFormData = "multipart/form-data"
)

type ISend interface {
	Send() Response
}

type Request struct {
	path           string
	URI            URI
	Method         HTTPMethod
	Header         IHeader
	Query          IQuery
	BodyParser     IBodyParser
	contextBuilder IContextBuilder
	Cookies        []*http.Cookie
}

// NewRequest creates a new request
// baseUrl: the base url ex: http://localhost:8080, http://localhost/path, path/, path?lorem=ipsum
// opts: options to configure the request
func NewRequest(path string, opts ...OptionRequest) *Request {
	r := &Request{
		path: path,
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.BodyParser == nil {
		r.BodyParser = dumpBody{}
	}

	if r.Method == "" {
		r.Method = GET
	}

	return r
}

func (r *Request) Send() (Response, error) {
	c := new(http.Client)

	req := r.request()

	response, err := c.Do(req)

	if err != nil {
		return NewErrorResponse(req, err)
	}

	return NewResponse(response, req), nil
}

func (r *Request) request() *http.Request {
	var context Context

	if r.contextBuilder != nil {
		context = r.contextBuilder.Build()
	} else {
		context = NewContextBuilder(30 * time.Second).Build()
	}

	if r.Query != nil && r.Query.String() != "" {
		r.URI = NewURI(r.path + "?" + r.Query.String())
	} else {
		r.URI = NewURI(r.path)
	}

	req, _ := http.NewRequestWithContext(context.Ctx, r.Method.String(), r.URI.String(), r.BodyParser.Parse())

	if r.Header != nil {
		r.Header.Properties().Each(func(k string, v any) {
			req.Header.Add(k, v.(string))
		})
	}

	if r.BodyParser.ContentType() != "" {
		req.Header.Set("Content-Type", r.BodyParser.ContentType())
	}

	if r.Cookies != nil && len(r.Cookies) > 0 {
		for _, cookie := range r.Cookies {
			req.AddCookie(cookie)
		}
	}

	return req
}

func (r *Request) SetBaseUrl(baseUrl string) *Request {
	if strings.HasPrefix(r.path, "/") {
		r.path = r.path[1:]
	}

	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl[:len(baseUrl)-1]
	}

	if strings.HasPrefix(r.path, baseUrl) {
		r.path = r.path[len(baseUrl)+1:]
	}

	r.path = baseUrl + "/" + r.path

	return r
}

func (r *Request) MergeHeader(header IHeader) *Request {
	if header != nil {
		if r.Header == nil {
			r.Header = header
		} else {
			r.Header.Merge(header)
		}
	}

	return r
}

func (r *Request) SetContextBuilder(contextBuilder IContextBuilder) *Request {
	if contextBuilder == nil {
		return r
	}

	r.contextBuilder = contextBuilder

	return r
}

type OptionRequest func(request *Request)

func WithMethod(method HTTPMethod) OptionRequest {
	return func(request *Request) {
		request.Method = method
	}
}

func WithBody(bodyParser IBodyParser) OptionRequest {
	return func(request *Request) {
		request.BodyParser = bodyParser
	}
}

func WithQuery(query IQuery) OptionRequest {
	return func(request *Request) {
		request.Query = query
	}
}

func WithHeader(header IHeader) OptionRequest {
	return func(request *Request) {
		request.Header = header
	}
}

func WithContextBuilder(contextBuilder IContextBuilder) OptionRequest {
	return func(request *Request) {
		request.contextBuilder = contextBuilder
	}
}

func WithCookies(cookies ...*http.Cookie) OptionRequest {
	return func(request *Request) {
		request.Cookies = cookies
	}
}
