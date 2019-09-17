// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signer

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/WICG/webpackage/go/signedexchange"
	"github.com/WICG/webpackage/go/signedexchange/structuredheader"
	"github.com/ampproject/amppackager/packager/accept"
	"github.com/ampproject/amppackager/packager/mux"
	"github.com/ampproject/amppackager/packager/rtv"
	pkgt "github.com/ampproject/amppackager/packager/testing"
	"github.com/ampproject/amppackager/packager/util"
	"github.com/ampproject/amppackager/transformer"
	rpb "github.com/ampproject/amppackager/transformer/request"
	"github.com/stretchr/testify/suite"
)

var fakePath = "/amp/secret-life-of-pine-trees.html"
var fakeBody = []byte("<html amp><body>They like to OPINE. Get it? (Is he fir real? Yew gotta be kidding me.)")
var transformedBody = []byte("<html amp><head></head><body>They like to OPINE. Get it? (Is he fir real? Yew gotta be kidding me.)</body></html>")

func headerNames(headers http.Header) []string {
	names := make([]string, len(headers))
	i := 0
	for name := range headers {
		names[i] = strings.ToLower(name)
		i++
	}
	sort.Strings(names)
	return names
}

type fakeCertHandler struct {
}

func (this fakeCertHandler) GetLatestCert() *x509.Certificate {
	return pkgt.Certs[0]
}

type SignerSuite struct {
	suite.Suite
	httpServer, tlsServer *httptest.Server
	httpsClient           *http.Client
	shouldPackage         bool
	fakeHandler           func(resp http.ResponseWriter, req *http.Request)
	lastRequest           *http.Request
}

func (this *SignerSuite) new(urlSets []util.URLSet) http.Handler {
	forwardedRequestHeaders := []string{"Host", "X-Foo"}
	handler, err := New(fakeCertHandler{}, pkgt.Key, urlSets, &rtv.RTVCache{}, func() bool { return this.shouldPackage }, nil, true, forwardedRequestHeaders)
	this.Require().NoError(err)
	// Accept the self-signed certificate generated by the test server.
	handler.client = this.httpsClient
	return mux.New(nil, handler, nil)
}

func (this *SignerSuite) get(t *testing.T, handler http.Handler, target string) *http.Response {
	return pkgt.GetH(t, handler, target, http.Header{
		"AMP-Cache-Transform": {"google"}, "Accept": {"application/signed-exchange;v=" + accept.AcceptedSxgVersion}})
}

func (this *SignerSuite) getFRH(t *testing.T, handler http.Handler, target string, host string, header http.Header) *http.Response {
	return pkgt.GetHH(t, handler, target, host, header)
}

func (this *SignerSuite) getB(t *testing.T, handler http.Handler, target string, body string) *http.Response {
	return pkgt.GetBHH(t, handler, target, "", strings.NewReader(body), http.Header{
		"AMP-Cache-Transform": {"google"}, "Accept": {"application/signed-exchange;v=" + accept.AcceptedSxgVersion}})
}

func (this *SignerSuite) httpURL() string {
	return this.httpServer.URL
}

func (this *SignerSuite) httpHost() string {
	u, err := url.Parse(this.httpURL())
	this.Require().NoError(err)
	return u.Host
}

// Same port as httpURL, but with an HTTPS scheme.
func (this *SignerSuite) httpSignURL() string {
	u, err := url.Parse(this.httpURL())
	this.Require().NoError(err)
	u.Scheme = "https"
	return u.String()
}

func (this *SignerSuite) certSubjectCN() string {
	return pkgt.Certs[0].Subject.CommonName
}

func (this *SignerSuite) httpSignURL_CertSubjectCN() string {
	u, err := url.Parse(this.certSubjectCN())
	this.Require().NoError(err)
	u.Scheme = "https"
	return u.String()
}

func (this *SignerSuite) httpsURL() string {
	return this.tlsServer.URL
}

func (this *SignerSuite) httpsHost() string {
	u, err := url.Parse(this.httpsURL())
	this.Require().NoError(err)
	return u.Host
}

func (this *SignerSuite) SetupSuite() {
	// Mock out example.com endpoint.
	this.httpServer = httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		this.fakeHandler(resp, req)
	}))

	this.tlsServer = httptest.NewTLSServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		this.fakeHandler(resp, req)
	}))
	this.httpsClient = this.tlsServer.Client()
	// Configure the test httpsClient to have the same redirect policy as production.
	this.httpsClient.CheckRedirect = noRedirects
}

func (this *SignerSuite) TearDownSuite() {
	this.httpServer.Close()
	this.tlsServer.Close()
}

func (this *SignerSuite) SetupTest() {
	this.shouldPackage = true
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		this.lastRequest = req
		resp.Header().Set("Content-Type", "text/html")
		resp.Write(fakeBody)
	}
	// Don't actually do any transforms. Only parse & print.
	getTransformerRequest = func(r *rtv.RTVCache, s, u string) *rpb.Request {
		return &rpb.Request{Html: string(s), DocumentUrl: u, Config: rpb.Request_NONE,
			AllowedFormats: []rpb.Request_HtmlFormat{rpb.Request_AMP}}
	}
}

func (this *SignerSuite) TestSimple() {
	urlSets := []util.URLSet{{
		Sign:  &util.URLPattern{[]string{"https"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
		Fetch: &util.URLPattern{[]string{"http"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, boolPtr(true)},
	}}
	resp := this.get(this.T(), this.new(urlSets),
		"/priv/doc?fetch="+url.QueryEscape(this.httpURL()+fakePath)+
			"&sign="+url.QueryEscape(this.httpSignURL()+fakePath))

	this.Assert().Equal(fakePath, this.lastRequest.URL.String())
	this.Assert().Equal(userAgent, this.lastRequest.Header.Get("User-Agent"))
	this.Assert().Equal("1.1 amppkg", this.lastRequest.Header.Get("Via"))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	this.Assert().Equal(fmt.Sprintf(`google;v="%d"`, transformer.SupportedVersions[0].Max), resp.Header.Get("AMP-Cache-Transform"))
	this.Assert().Equal("nosniff", resp.Header.Get("X-Content-Type-Options"))
	this.Assert().Equal("Accept, AMP-Cache-Transform", resp.Header.Get("Vary"))

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(this.httpSignURL()+fakePath, exchange.RequestURI)
	this.Assert().Equal("GET", exchange.RequestMethod)
	this.Assert().Equal(http.Header{}, exchange.RequestHeaders)
	this.Assert().Equal(200, exchange.ResponseStatus)
	this.Assert().Equal(
		[]string{"content-encoding", "content-length", "content-security-policy", "content-type", "date", "digest", "x-content-type-options"},
		headerNames(exchange.ResponseHeaders))
	this.Assert().Equal("text/html", exchange.ResponseHeaders.Get("Content-Type"))
	this.Assert().Equal("nosniff", exchange.ResponseHeaders.Get("X-Content-Type-Options"))
	this.Assert().Contains(exchange.SignatureHeaderValue, "validity-url=\""+this.httpSignURL()+"/amppkg/validity\"")
	this.Assert().Contains(exchange.SignatureHeaderValue, "integrity=\"digest/mi-sha256-03\"")
	this.Assert().Contains(exchange.SignatureHeaderValue, "cert-url=\""+this.httpSignURL()+"/amppkg/cert/"+pkgt.CertName+"\"")
	certHash, _ := base64.RawURLEncoding.DecodeString(pkgt.CertName)
	this.Assert().Contains(exchange.SignatureHeaderValue, "cert-sha256=*"+base64.StdEncoding.EncodeToString(certHash[:])+"*")
	// TODO(twifkak): Control date, and test for expires and sig.

	signatures, err := structuredheader.ParseParameterisedList(exchange.SignatureHeaderValue)
	this.Require().NoError(err)
	this.Require().NotEmpty(signatures)
	date, ok := signatures[0].Params["date"].(int64)
	this.Require().True(ok)
	expires, ok := signatures[0].Params["expires"].(int64)
	this.Require().True(ok)
	this.Assert().Equal(int64(604800), expires-date)

	// The response header values are untested here, as that is covered by signedexchange tests.

	// For small enough bodies, the only thing that MICE does is add a record size prefix.
	var payloadPrefix bytes.Buffer
	binary.Write(&payloadPrefix, binary.BigEndian, uint64(miRecordSize))
	this.Assert().Equal(append(payloadPrefix.Bytes(), transformedBody...), exchange.Payload)
}

func (this *SignerSuite) TestFetchSignWithForwardedRequestHeaders() {
	urlSets := []util.URLSet{{
		Sign:  &util.URLPattern{[]string{"https"}, "", this.certSubjectCN(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
		Fetch: &util.URLPattern{[]string{"http"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, boolPtr(true)},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		// Host and X-Foo headers are forwarded with forwardedRequestHeaders
		this.Assert().Equal("www.example.com", req.Host)
		this.Assert().Equal("foo", req.Header.Get("X-Foo"))
		this.Assert().Equal("", req.Header.Get("X-Bar"))
		this.lastRequest = req
		resp.Header().Set("Content-Type", "text/html")
		resp.Write(fakeBody)
	}
	// Request with ForwardedRequestHeaders
	header := http.Header{"AMP-Cache-Transform": {"google"}, "Accept": {"application/signed-exchange;v=" + accept.AcceptedSxgVersion},
		"X-Foo": {"foo"}, "X-Bar": {"bar"}}
	resp := this.getFRH(this.T(), this.new(urlSets),
		"/priv/doc?fetch="+url.QueryEscape(this.httpURL()+fakePath)+"&sign="+url.QueryEscape(this.httpSignURL_CertSubjectCN()+fakePath),
		"www.example.com", header)
	this.Assert().Equal(fakePath, this.lastRequest.URL.String())
	this.Assert().Equal(userAgent, this.lastRequest.Header.Get("User-Agent"))
	this.Assert().Equal("1.1 amppkg", this.lastRequest.Header.Get("Via"))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	this.Assert().Equal(fmt.Sprintf(`google;v="%d"`, transformer.SupportedVersions[0].Max), resp.Header.Get("AMP-Cache-Transform"))
	this.Assert().Equal("nosniff", resp.Header.Get("X-Content-Type-Options"))
	this.Assert().Equal("Accept, AMP-Cache-Transform", resp.Header.Get("Vary"))
	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(this.httpSignURL_CertSubjectCN()+fakePath, exchange.RequestURI)
	this.Assert().Equal("GET", exchange.RequestMethod)
	this.Assert().Equal(http.Header{}, exchange.RequestHeaders)
	this.Assert().Equal(200, exchange.ResponseStatus)
	this.Assert().Equal(
		[]string{"content-encoding", "content-length", "content-security-policy", "content-type", "date", "digest", "x-content-type-options"},
		headerNames(exchange.ResponseHeaders))
	this.Assert().Equal("text/html", exchange.ResponseHeaders.Get("Content-Type"))
	this.Assert().Equal("text/html", exchange.ResponseHeaders.Get("Content-Type"))
	this.Assert().Equal("nosniff", exchange.ResponseHeaders.Get("X-Content-Type-Options"))
	this.Assert().Contains(exchange.SignatureHeaderValue, "validity-url=\""+this.httpSignURL_CertSubjectCN()+"/amppkg/validity\"")
	this.Assert().Contains(exchange.SignatureHeaderValue, "integrity=\"digest/mi-sha256-03\"")
	this.Assert().Contains(exchange.SignatureHeaderValue, "cert-url=\""+this.httpSignURL_CertSubjectCN()+"/amppkg/cert/"+pkgt.CertName+"\"")
	certHash, _ := base64.RawURLEncoding.DecodeString(pkgt.CertName)
	this.Assert().Contains(exchange.SignatureHeaderValue, "cert-sha256=*"+base64.StdEncoding.EncodeToString(certHash[:])+"*")
	var payloadPrefix bytes.Buffer
	binary.Write(&payloadPrefix, binary.BigEndian, uint64(miRecordSize))
	this.Assert().Equal(append(payloadPrefix.Bytes(), transformedBody...), exchange.Payload)
}

func (this *SignerSuite) TestEscapeQueryParamsInFetchAndSign() {
	urlSets := []util.URLSet{{
		Sign:  &util.URLPattern{[]string{"https"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(".*"), false, 2000, nil},
		Fetch: &util.URLPattern{[]string{"http"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(".*"), false, 2000, boolPtr(true)},
	}}
	resp := this.get(this.T(), this.new(urlSets),
		"/priv/doc?fetch="+url.QueryEscape(this.httpURL()+fakePath+"?<hi>")+
			"&sign="+url.QueryEscape(this.httpSignURL()+fakePath+"?<hi>"))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	this.Assert().Equal(fakePath+"?%3Chi%3E", this.lastRequest.URL.String())

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(this.httpSignURL()+fakePath+"?%3Chi%3E", exchange.RequestURI)
}

func (this *SignerSuite) TestDisallowInvalidCharsSign() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := this.get(this.T(), this.new(urlSets),
		"/priv/doc?&sign="+url.QueryEscape(this.httpSignURL()+fakePath+"<hi>"))
	this.Assert().Equal(http.StatusBadRequest, resp.StatusCode, "incorrect status: %#v", resp)
}

func (this *SignerSuite) TestNoFetchParam() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakePath, this.lastRequest.URL.String())
	this.Assert().Equal(this.httpsURL()+fakePath, exchange.RequestURI)
}

func (this *SignerSuite) TestSignAsPathParam() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := this.get(this.T(), this.new(urlSets), `/priv/doc/`+this.httpsURL()+fakePath)
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakePath, this.lastRequest.URL.String())
	this.Assert().Equal(this.httpsURL()+fakePath, exchange.RequestURI)
}

func (this *SignerSuite) TestSignAsPathParamWithQuery() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(".*"), false, 2000, nil},
	}}
	resp := this.get(this.T(), this.new(urlSets), `/priv/doc/`+this.httpsURL()+fakePath+"?amp=1")
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakePath+"?amp=1", this.lastRequest.URL.String())
	this.Assert().Equal(this.httpsURL()+fakePath+"?amp=1", exchange.RequestURI)
}

// Ensure that the server doesn't attempt to percent-decode the sign URL.
func (this *SignerSuite) TestSignAsPathParamWithUnusualPctEncoding() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := this.get(this.T(), this.new(urlSets), `/priv/doc/`+this.httpsURL()+fakePath+`%2A`)
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakePath+"%2A", this.lastRequest.URL.String())
	this.Assert().Equal(this.httpsURL()+fakePath+"%2A", exchange.RequestURI)
}

func (this *SignerSuite) TestPreservesContentType() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html;charset=utf-8;v=5")
		resp.Write(fakeBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal("text/html;charset=utf-8;v=5", exchange.ResponseHeaders.Get("Content-Type"))
}

func (this *SignerSuite) TestRemovesLinkHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Link", "rel=preload;<http://1.2.3.4/>")
		resp.Write(fakeBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Link"))
}

func (this *SignerSuite) TestRemovesStatefulHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Set-Cookie", "yum yum yum")
		resp.Write(fakeBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Set-Cookie"))
}

func (this *SignerSuite) TestMutatesCspHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{
			[]string{"https"},
			"",
			this.httpsHost(),
			stringPtr("/amp/.*"),
			[]string{},
			stringPtr(""),
			false,
			2000,
			nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Expect base-uri and block-all-mixed-content to remain unmodified.
		// Expect require-sri-for to be stripped.
		// Expect script-src to be overwritten.
		resp.Header().Set(
			"Content-Security-Policy",
			"base-uri http://*.example.com; "+
				"block-all-mixed-content; "+
				"require-sri-for script; "+
				"script-src https://notallowed.org/")
		resp.Write(fakeBody)
	}

	resp := this.get(
		this.T(),
		this.new(urlSets),
		"/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(
		http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(
		"base-uri http://*.example.com;"+
			"block-all-mixed-content;"+
			"default-src * blob: data:;"+
			"report-uri https://csp-collector.appspot.com/csp/amp;"+
			"script-src blob: https://cdn.ampproject.org/rtv/ https://cdn.ampproject.org/v0.js https://cdn.ampproject.org/v0/ https://cdn.ampproject.org/viewer/;"+
			"style-src 'unsafe-inline' https://cdn.materialdesignicons.com https://cloud.typography.com https://fast.fonts.net https://fonts.googleapis.com https://maxcdn.bootstrapcdn.com https://p.typekit.net https://pro.fontawesome.com https://use.fontawesome.com https://use.typekit.net;"+
			"object-src 'none'",
		exchange.ResponseHeaders.Get("Content-Security-Policy"))
}

func (this *SignerSuite) TestAddsLinkHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Write([]byte("<html amp><head><link rel=stylesheet href=foo><script src=bar>"))
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal("<foo>;rel=preload;as=style,<bar>;rel=preload;as=script", exchange.ResponseHeaders.Get("Link"))
}

func (this *SignerSuite) TestEscapesLinkHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		// This shouldn't happen for valid AMP, and AMP Caches should
		// verify the Link header so that it wouldn't be ingested.
		// However, it would be nice to limit the impact that could be
		// caused by transformation of an invalid AMP, e.g. on a
		// same-origin impression.
		resp.Write([]byte(`<html amp><head><script src="https://foo.com/a,b>c?d>e|f">`))
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal("<https://foo.com/a,b%3Ec?d%3Ee%7Cf>;rel=preload;as=script", exchange.ResponseHeaders.Get("Link"))
}

func (this *SignerSuite) TestRemovesHopByHopHeaders() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Connection", "PROXY-AUTHENTICATE, Server")
		resp.Header().Set("Proxy-Authenticate", "Basic")
		resp.Header().Set("Server", "thing")
		resp.Header().Set("Transfer-Encoding", "chunked") // Also removed, per RFC 2616.
		resp.Write([]byte("<html amp><head><link rel=stylesheet href=foo><script src=bar>"))
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	this.Assert().Contains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Content-Type"))
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Connection"))
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Proxy-Authenticate"))
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Server"))
	this.Assert().NotContains(exchange.ResponseHeaders, http.CanonicalHeaderKey("Transfer-Encoding"))
}

func (this *SignerSuite) TestLimitsDuration() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Write([]byte("<html amp><body><amp-script script max-age=4000>"))
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	signatures, err := structuredheader.ParseParameterisedList(exchange.SignatureHeaderValue)
	this.Require().NoError(err)
	this.Require().NotEmpty(signatures)
	date, ok := signatures[0].Params["date"].(int64)
	this.Require().True(ok)
	expires, ok := signatures[0].Params["expires"].(int64)
	this.Require().True(ok)
	this.Assert().Equal(int64(4000), expires-date)
}

func (this *SignerSuite) TestDoesNotExtendDuration() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Write([]byte("<html amp><body><amp-script script max-age=700000>"))
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	exchange, err := signedexchange.ReadExchange(resp.Body)
	this.Require().NoError(err)
	signatures, err := structuredheader.ParseParameterisedList(exchange.SignatureHeaderValue)
	this.Require().NoError(err)
	this.Require().NotEmpty(signatures)
	date, ok := signatures[0].Params["date"].(int64)
	this.Require().True(ok)
	expires, ok := signatures[0].Params["expires"].(int64)
	this.Require().True(ok)
	this.Assert().Equal(int64(604800), expires-date)
}

func (this *SignerSuite) TestErrorNoCache() {
	urlSets := []util.URLSet{{
		Fetch: &util.URLPattern{[]string{"http"}, "", this.httpHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, boolPtr(true)},
	}}
	// Missing sign param generates an error.
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?fetch="+url.QueryEscape(this.httpURL()+fakePath))
	this.Assert().Equal(http.StatusBadRequest, resp.StatusCode, "incorrect status: %#v", resp)
	this.Assert().Equal("no-store", resp.Header.Get("Cache-Control"))
}

func (this *SignerSuite) TestProxyUnsignedIfRedirect() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Set-Cookie", "yum yum yum")
		resp.Header().Set("Location", "/login")
		resp.WriteHeader(301)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(301, resp.StatusCode)
	this.Assert().Equal("yum yum yum", resp.Header.Get("set-cookie"))
	this.Assert().Equal("/login", resp.Header.Get("location"))
}

func (this *SignerSuite) TestProxyUnsignedIfNotModified() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Cache-control", "private")
		resp.Header().Set("Cookie", "yum yum yum")
		resp.Header().Set("ETag", "superrad")
		resp.WriteHeader(304)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(304, resp.StatusCode)
	this.Assert().Equal("private", resp.Header.Get("cache-control"))
	this.Assert().Equal("", resp.Header.Get("cookie"))
	this.Assert().Equal("superrad", resp.Header.Get("etag"))
}

func (this *SignerSuite) TestProxyUnsignedIfShouldntPackage() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	this.shouldPackage = false
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakeBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyUnsignedIfMissingAMPCacheTransformHeader() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := pkgt.GetH(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath), http.Header{
		"Accept": {"application/signed-exchange;v=" + accept.AcceptedSxgVersion}})
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakeBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyUnsignedIfInvalidAMPCacheTransformHeader() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := pkgt.GetH(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath), http.Header{
		"Accept":              {"application/signed-exchange;v=" + accept.AcceptedSxgVersion},
		"AMP-Cache-Transform": {"donotmatch"},
	})
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakeBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyUnsignedIfMissingAcceptHeader() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	resp := pkgt.GetH(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath), http.Header{
		"AMP-Cache-Transform": {"google"}})
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakeBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyUnsignedNonCachable() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html")
		resp.Header().Set("Cache-Control", "no-store")
		resp.WriteHeader(200)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("no-store", resp.Header.Get("Cache-Control"))
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))
}

func (this *SignerSuite) TestProxyUnsignedBadContentEncoding() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html")
		resp.Header().Set("Content-Encoding", "br")
		resp.WriteHeader(200)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("br", resp.Header.Get("Content-Encoding"))
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))
}

func (this *SignerSuite) TestProxyUnsignedErrOnStatefulHeader() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), true, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Set-Cookie", "chocolate chip")
		resp.Header().Set("Content-Type", "text/html")
		resp.WriteHeader(200)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("chocolate chip", resp.Header.Get("Set-Cookie"))
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))
}

func (this *SignerSuite) TestProxyUnsignedOnVariants() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), true, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Variants", "foo")
		resp.Header().Set("Content-Type", "text/html")
		resp.WriteHeader(200)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("foo", resp.Header.Get("Variants"))
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))
}

func (this *SignerSuite) TestProxyUnsignedOnVariants04() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), true, 2000, nil},
	}}
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		resp.Header().Set("Variants-04", "foo")
		resp.Header().Set("Content-Type", "text/html")
		resp.WriteHeader(200)
	}

	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("foo", resp.Header.Get("Variants-04"))
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))
}

func (this *SignerSuite) TestProxyUnsignedIfNotAMP() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	nonAMPBody := []byte("<html><body>They like to OPINE. Get it? (Is he fir real? Yew gotta be kidding me.)")
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html")
		resp.Write(nonAMPBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(nonAMPBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyUnsignedIfWrongAMP() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil}}}
	wrongAMPBody := []byte("<html amp4email><body>They like to OPINE. Get it? (Is he fir real? Yew gotta be kidding me.)")
	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html")
		resp.Write(wrongAMPBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(http.StatusOK, resp.StatusCode, "incorrect status: %#v", resp)

	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(wrongAMPBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyTransformError() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}

	// Generate a request for non-existent transformer that will fail
	getTransformerRequest = func(r *rtv.RTVCache, s, u string) *rpb.Request {
		return &rpb.Request{Html: string(s), DocumentUrl: u, Config: rpb.Request_CUSTOM,
			AllowedFormats: []rpb.Request_HtmlFormat{rpb.Request_AMP},
			Transformers:   []string{"bogus"}}
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)
	this.Assert().Equal("text/html", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	this.Require().NoError(err)
	this.Assert().Equal(fakeBody, body, "incorrect body: %#v", resp)
}

func (this *SignerSuite) TestProxyHeadersUnaltered() {
	urlSets := []util.URLSet{{
		Sign: &util.URLPattern{[]string{"https"}, "", this.httpsHost(), stringPtr("/amp/.*"), []string{}, stringPtr(""), false, 2000, nil},
	}}

	// "Perform local transformations" is close to the last opportunity that a
	// response could be proxied instead of signed. Intentionally cause an error
	// to occur so that we can verify the proxy response has not been altered.
	getTransformerRequest = func(r *rtv.RTVCache, s, u string) *rpb.Request {
		return &rpb.Request{Html: string(s), DocumentUrl: u, Config: rpb.Request_CUSTOM,
			AllowedFormats: []rpb.Request_HtmlFormat{rpb.Request_AMP},
			Transformers:   []string{"bogus"}}
	}

	originalHeaders := map[string]string{
		"Content-Type":   "text/html",
		"Set-Cookie":     "chocolate chip",
		"Cache-Control":  "max-age=31536000",
		"Content-Length": fmt.Sprintf("%d", len(fakeBody)),
	}

	this.fakeHandler = func(resp http.ResponseWriter, req *http.Request) {
		for key, value := range originalHeaders {
			resp.Header().Set(key, value)
		}
		resp.Write(fakeBody)
	}
	resp := this.get(this.T(), this.new(urlSets), "/priv/doc?sign="+url.QueryEscape(this.httpsURL()+fakePath))
	this.Assert().Equal(200, resp.StatusCode)

	// Compare the final headers to the originals, removing each one after
	// checking, so that we can finally verify that no additional headers were
	// appended.
	for key, value := range originalHeaders {
		this.Assert().Equal([]string{value}, resp.Header[key])
		resp.Header.Del(key)
	}
	this.Assert().Equal([]string{"Accept, AMP-Cache-Transform"}, resp.Header["Vary"])
	resp.Header.Del("Vary")
	resp.Header.Del("Date") // Date header is not tested; it may be updated.
	this.Assert().Empty(resp.Header)
}

func TestSignerSuite(t *testing.T) {
	suite.Run(t, new(SignerSuite))
}
