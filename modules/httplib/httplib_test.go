// Copyright 2013 The Beego Authors. All rights reserved.
// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httplib

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	Convey("When making Get request", t, func() {
		req := Get("http://httpbin.org/get")
		Convey("Response() should return Response structure", func() {
			resp, err := req.Response()
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
		})

		Convey("Bytes() as string should equal to String()", func() {
			b, errB := req.Bytes()
			s, errS := req.String()
			So(errB, ShouldBeNil)
			So(errS, ShouldBeNil)
			So(string(b), ShouldEqual, s)
		})
	})
}

func TestSimplePost(t *testing.T) {
	Convey("Should make simple POST request", t, func() {
		v := "smallfish"
		req := Post("http://httpbin.org/post")
		req.Param("username", v)

		str, err := req.String()
		So(err, ShouldBeNil)
		n := strings.Index(str, v)
		So(n, ShouldNotEqual, -1)
	})
}

// func TestPostFile(t *testing.T) {
// 	v := "smallfish"
// 	req := Post("http://httpbin.org/post")
// 	req.Param("username", v)
// 	req.PostFile("uploadfile", "httplib_test.go")

// 	str, err := req.String()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(str)

// 	n := strings.Index(str, v)
// 	if n == -1 {
// 		t.Fatal(v + " not found in post")
// 	}
// }

func TestSimplePut(t *testing.T) {
	Convey("Should make simple put without errors", t, func() {
		_, err := Put("http://httpbin.org/put").String()
		So(err, ShouldBeNil)
	})
}

func TestSimpleDelete(t *testing.T) {
	Convey("Should make simple delete without errors", t, func() {
		_, err := Delete("http://httpbin.org/delete").String()
		So(err, ShouldBeNil)
	})
}

func TestWithCookie(t *testing.T) {
	Convey("Should have cookies support", t, func() {
		v := "smallfish"
		_, err := Get("http://httpbin.org/cookies/set?k1=" + v).SetEnableCookie(true).String()
		So(err, ShouldBeNil)
		str, err := Get("http://httpbin.org/cookies").SetEnableCookie(true).String()
		So(err, ShouldBeNil)
		n := strings.Index(str, v)
		So(n, ShouldNotEqual, -1)
	})
}

func TestWithBasicAuth(t *testing.T) {
	Convey("Should have support for Basic AUTH", t, func() {
		str, err := Get("http://httpbin.org/basic-auth/user/passwd").SetBasicAuth("user", "passwd").String()
		So(err, ShouldBeNil)
		n := strings.Index(str, "authenticated")
		So(n, ShouldNotEqual, -1)
	})
}

func TestWithUserAgent(t *testing.T) {
	Convey("Should have support for user-agent header", t, func() {
		v := "beego"
		str, err := Get("http://httpbin.org/headers").SetUserAgent(v).String()
		So(err, ShouldBeNil)
		n := strings.Index(str, v)
		So(n, ShouldNotEqual, -1)
	})
}

func TestWithSetting(t *testing.T) {
	Convey("Should have possibility to set default settings", t, func() {
		v := "beego"
		var setting BeegoHttpSettings
		setting.EnableCookie = true
		setting.UserAgent = v
		setting.Transport = nil
		SetDefaultSetting(setting)

		str, err := Get("http://httpbin.org/get").String()
		So(err, ShouldBeNil)
		n := strings.Index(str, v)
		So(n, ShouldNotEqual, -1)
	})
}

func TestToJson(t *testing.T) {
	Convey("Should export to JSON", t, func() {
		req := Get("http://httpbin.org/ip")
		_, err := req.Response()
		So(err, ShouldBeNil)
		// httpbin will return http remote addr
		type Ip struct {
			Origin string `json:"origin"`
		}
		var ip Ip
		err = req.ToJson(&ip)
		So(err, ShouldBeNil)
		n := strings.Count(ip.Origin, ".")
		So(n, ShouldEqual, 3)
	})
}

func TestToFile(t *testing.T) {
	Convey("Should export to file", t, func() {
		f := "beego_testfile"
		req := Get("http://httpbin.org/ip")
		err := req.ToFile(f)
		So(err, ShouldBeNil)
		defer os.Remove(f)
		b, err := ioutil.ReadFile(f)
		n := strings.Index(string(b), "origin")
		So(n, ShouldNotEqual, -1)
	})
}

func TestHeader(t *testing.T) {
	Convey("Should have support for headers", t, func() {
		req := Get("http://httpbin.org/headers")
		req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.57 Safari/537.36")
		_, err := req.String()
		So(err, ShouldBeNil)
	})
}
