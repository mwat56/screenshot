/*
   Copyright © 2022 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/

package screenshot

import (
	"context"
	"net/http"
	"os"
	"testing"
)

const (
	testImgDirectory = "/tmp/"
	agentTest        = "Screenshot debugging aid v0.0.1"
	agentLynx        = "Lynx/2.8.9dev.16 libwww-FM/2.14 SSL-MM/1.4.1 GNUTLS/3.5.17"
	agentFirefox     = `Mozilla/5.0 (X11; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0`
)

func Test_fileExt(t *testing.T) {
	type args struct {
		aURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{" 1", args{""}, ""},
		{" 2", args{"image.gif"}, ".gif"},
		{" 3", args{"document.txt"}, ".txt"},
		{" 4", args{"document.txt.doc"}, ".doc"},
		{" 5", args{"http://example.com/page.html?view=print"}, ".html"},
		{" 6", args{"http://example.com/sometopic?show=all&lang=en"}, ""},
		{" 5", args{"http://example.com/page.md?view=print#top"}, ".md"},
		{" 6", args{"https://github.com/mwat56/Nele/blob/master/README.md#nele-blog"}, ".md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileExt(tt.args.aURL); got != tt.want {
				t.Errorf("extension() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_fileExt()

func Test_generateImage(t *testing.T) {
	SetImageDir(testImgDirectory)

	type args struct {
		aContext context.Context
		aURL     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{context.TODO(), "http://www.mwat.de/Antik/"}, false},
		{" 2", args{context.TODO(), "ftp://ftp.gibbet.nich/"}, true},
		{" 3", args{context.TODO(), "gopher://localhost/index"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generateImage(tt.args.aContext, tt.args.aURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
} // Test_generateImage()

func Test_sanitise(t *testing.T) {
	tests := []struct {
		name string
		aURL string
		want string
	}{
		// TODO: Add test cases.
		{" 1", "http://dev.mwat.de/#main", "httpdevmwatdemain"},
		{" 2", "http://www.gibbet.nich/~matthias/index.html", "httpwwwgibbetnichmatthiasindexhtml"},
		{" 3", "gopher://localhost/a/b/c", "gopherlocalhostabc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitise(tt.aURL); got != tt.want {
				t.Errorf("sanitise() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_sanitise()

func Test_writeFile(t *testing.T) {
	SetImageDir(testImgDirectory)

	n1 := PathFile("http://www.mwat.de/")
	d1 := []byte("\nScreenShot_1\n")
	n2 := PathFile("http://dev.mwat.de/")
	d2 := []byte("\nScreenShot_2\n")
	n3 := PathFile("http://bla.mwat.de")
	d3 := []byte("\nScreenShot_3\n")
	n4 := "/tmp/ScreenShot_4"
	var d4 []byte
	var n5 string
	d5 := []byte("\nScreenShot_5\n")

	defer func() {
		_ = os.Remove(n1)
		_ = os.Remove(n2)
		_ = os.Remove(n3)
		_ = os.Remove(n4)
		_ = os.Remove(n5)
	}()

	type args struct {
		aName     string
		aData     []byte
		aResponse *http.Response
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{n1, d1, nil}, false},
		{" 2", args{n2, d2, nil}, false},
		{" 3", args{n3, d3, nil}, false},
		{" 4", args{n4, d4, nil}, true},
		{" 5", args{n5, d5, nil}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeFile(tt.args.aName, tt.args.aData, tt.args.aResponse); (err != nil) != tt.wantErr {
				t.Errorf("writeFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} // Test_writeFile()

func TestCreateImage(t *testing.T) {
	SetImageDir(testImgDirectory)
	// SetImageHeight(0)
	SetImageQuality(75)
	// SetImageWidth(1024)
	SetJavaScript(false)
	SetImageScale(0.999)
	// SetUserAgent(agentFirefox)
	// SetUserAgent(agentLynx)
	// SetUserAgent(agentTest)
	fileExt := ImageType()

	u1 := "https://www.buzzfeednews.com/article/alexkantrowitz/how-the-retweet-ruined-the-internet"
	w1 := sanitise(u1) + "." + fileExt
	u2 := "https://github.com/mwat56/pageview"
	w2 := sanitise(u2) + "." + fileExt
	u3 := "https://www.eff.org/"
	w3 := sanitise(u3) + "." + fileExt
	u4 := "https://www.ohchr.org/Documents/Issues/Opinion/Legislation/OL-DEU-1-2017.pdf"
	u5 := "https://www.startpage.com/do/mypage.pl?prfe=11742c69614c5050a145b9bdfad5f008146ec505c5c35adf413b6739b701f24f929c4c53c5685edf58d38e108cf7c95f86bb0436a1b232f8dbddd145af3b1808fba2c6679ec1fdea5462bd4c31b364e98f"
	w5 := sanitise(u5) + "." + fileExt
	u6 := "https://uncommongroundmedia.com/thats-not-what-i-meant-brain-damage-communication-part-i-%EF%BB%BF-dr-em/"
	w6 := sanitise(u6) + "." + fileExt
	u7 := "https://thehackernews.com/2017/10/kaspersky-nsa-russian-hackers.html#articlebody"
	w7 := sanitise(u7) + "." + fileExt
	u8 := "http://www.mwat.de/CSS/mwBack-526x841.gif"
	w8 := sanitise(u8) + ".gif"
	u9 := "https://www.youtube.com/watch?v=tvcwMcGWf-w"
	w9 := sanitise(u9) + "." + fileExt
	u10 := "https://twitter.com/seerutkchawla/status/1337231261430132738"
	w10 := sanitise(u10) + "." + fileExt
	u11 := "https://www.facebook.com/robertjsawyer/posts/10155509139641013#contentArea"
	w11 := sanitise(u11) + "." + fileExt
	u12 := "https://diekolumnisten.de/2018/12/28/deutschland-verrecke/#post-19938"
	w12 := sanitise(u12) + "." + fileExt

	prep := func() {
		_ = os.Remove(PathFile(u1))
		_ = os.Remove(PathFile(u2))
		_ = os.Remove(PathFile(u3))
		_ = os.Remove(PathFile(u4))
		_ = os.Remove(PathFile(u5))
		_ = os.Remove(PathFile(u6))
		_ = os.Remove(PathFile(u7))
		_ = os.Remove(PathFile(u8))
		_ = os.Remove(PathFile(u9))
		_ = os.Remove(PathFile(u10))
		_ = os.Remove(PathFile(u11))
		_ = os.Remove(PathFile(u12))
	}
	prep()
	// defer func() {
	// 	prep()
	// }()

	tests := []struct {
		name    string
		aURL    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", u1, w1, false},
		{" 2", u2, w2, false},
		{" 3", u3, w3, false},
		{" 4", u4, "", true},
		{" 5", u5, w5, false},
		{" 6", u6, w6, false},
		{" 7", u7, w7, false},
		{" 8", u8, w8, false},
		{" 9", u9, w9, false},
		{"10", u10, w10, false},
		{"11", u11, w11, false},
		{"12", u12, w12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateImage(tt.aURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateImage() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateImage() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestCreateImage()

func TestSetup(t *testing.T) {
	o1 := &ScreenshotParams{}
	o2 := &ScreenshotParams{
		ImageAge: 0, // matches default value
	}
	o3 := &ScreenshotParams{
		ImageDir: "",
	}

	tests := []struct {
		name     string
		aOptions *ScreenshotParams
	}{
		// TODO: Add test cases.
		{" 1", o1},
		{" 2", o2},
		{" 3", o3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(tt.aOptions)
		})
	}
} // TestSetup()

func TestString(t *testing.T) {
	w1 := `CertErrors:	false
Cookies:	false
ImageAge:	0
ImageDir:	'/tmp'
ImageHeight:	768
ImageQuality:	100
ImageScale:	0.00
ImageWidth:	896
JavaScript:	false
Mobile:	false
Platform:	'Linux x86_64'
Scrollbars:	false
UserAgent:	'Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0'
`
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{" 1", w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(); got != tt.want {
				t.Errorf("String() = »%v«, want »%v«", got, tt.want)
			}
		})
	}
} // TestString()
