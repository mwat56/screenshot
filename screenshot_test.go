/*
   Copyright © 2022 M.Watermann, 10247 Berlin, Germany
                 All rights reserved
               EMail : <support@mwat.de>
*/
package screenshot

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

const (
	testImgDirectory = "/tmp/"
	agentTest        = "Screenshot debugging aid v0.0.1"
	agentLynx        = "Lynx/2.8.9dev.16 libwww-FM/2.14 SSL-MM/1.4.1 GNUTLS/3.5.17"
	agentFirefox     = `Mozilla/5.0 (X11; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0`
)

func setupScreenshot() {
	SetAcceptOther(true)
	SetCookies(false)
	SetImageAge(0)
	SetImageDir(testImgDirectory)
	SetImageHeight(768)
	SetImageQuality(75)
	SetImageScale(0.99)
	SetImageWidth(896)
	SetJavaScript(false)
	SetMaxProcessTime(24)
	SetScrollbars(true)
	SetUserAgent(agentFirefox)
	// SetUserAgent(agentLynx)
	// SetUserAgent(agentTest)
} // setupScreenshot()

func Test_chk4(t *testing.T) {
	type args struct {
		aURL       string
		aHostsFile string
	}
	//
	h1 := ""
	u1 := ""
	w1 := false
	//
	h2 := HostsAvoidJS
	u2 := ""
	w2 := false
	//
	h3 := HostsNeedJS
	u3 := "Twitter.Com"
	w3 := true
	//
	h4 := HostsNeedJS
	u4 := "https://you.tube.me/"
	w4 := false
	//
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{" 1", args{u1, h1}, w1},
		{" 2", args{u2, h2}, w2},
		{" 3", args{u3, h3}, w3},
		{" 4", args{u4, h4}, w4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chk4(tt.args.aURL, tt.args.aHostsFile); got != tt.want {
				t.Errorf("chk4() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_chk4()

func Test_containsHost(t *testing.T) {
	type args struct {
		aNeedle   string
		aHaystack *sort.StringSlice
	}
	//
	n1 := ""
	h1 := &sort.StringSlice{}
	//
	n2 := "www.example.org"
	h2 := &sort.StringSlice{
		"",
		"# test list",
		"",
		"gibbet.nich",
		"example.com",
		"",
		"# _EoF_",
	} // h2
	//
	n3 := "www.example.com"
	h3 := h2
	//
	n4 := "example.com"
	h4 := &sort.StringSlice{
		"",
		"# test list",
		".example.com",
		"",
		"# _EoF_",
	} // h4
	//

	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{" 1", args{n1, h1}, false},
		{" 2", args{n2, h2}, false},
		{" 3", args{n3, h3}, true},
		{" 4", args{n4, h4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsHost(tt.args.aNeedle, tt.args.aHaystack); got != tt.want {
				t.Errorf("listContains() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_containsHost()

func Test_exists(t *testing.T) {
	f1 := `/ is not there`   // root dir is not writeable
	f2 := ``                 // invalid (empty) filename
	f3 := `/home`            // is directory, i.e. irregular
	f4 := `/etc/cron.d/`     // (dito)
	f5 := `/etc/crontab`     // exists but too small
	f6 := `/etc/ld.so.cache` // exists and big enough
	tests := []struct {
		name      string
		aFilename string
		want      bool
	}{
		// TODO: Add test cases.
		{" 1", f1, false},
		{" 2", f2, false},
		{" 3", f3, true},
		{" 4", f4, true},
		{" 5", f5, false},
		{" 6", f6, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exists(tt.aFilename); got != tt.want {
				t.Errorf("exists() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_exists()

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
	setupScreenshot()
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

func Test_readListFile(t *testing.T) {
	const fName = "./Crash_Test_Dummies.lst"
	list := `
# Sites requiring JavaScript to work: 'jsneedsites.lst'

Twitter.com

Facebook.com

YouTube.com

Youtu.be

# _EoF_
`
	writeFile(fName, []byte(list), nil)
	defer func() {
		_ = os.Remove(fName)
	}()

	n1 := ""
	var w1 sort.StringSlice

	n2 := "/dev/not/there"
	w2 := w1

	n3 := fName
	w3 := sort.StringSlice{"twitter.com", "facebook.com", "youtube.com", "youtu.be"}

	tests := []struct {
		name      string
		aFilename string
		wantRList sort.StringSlice
	}{
		// TODO: Add test cases.
		{" 1", n1, w1},
		{" 2", n2, w2},
		{" 3", n3, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRList := readListFile(tt.aFilename); !reflect.DeepEqual(gotRList, tt.wantRList) {
				t.Errorf("readListFile() = %v,\nwant %v", gotRList, tt.wantRList)
			}
		})
	}
} // Test_readListFile()

func Test_removeIndex(t *testing.T) {
	type args struct {
		aList  sort.StringSlice
		aIndex int
	}

	list := sort.StringSlice{
		"0", "1", "2", "3", "4", "5",
	}
	//
	l1 := list
	i1 := 0
	w1 := sort.StringSlice{
		"1", "2", "3", "4", "5",
	}
	//
	l2 := list
	i2 := 1
	w2 := sort.StringSlice{
		"0", "2", "3", "4", "5",
	}
	//
	l3 := list
	i3 := 3
	w3 := sort.StringSlice{
		"0", "1", "2", "4", "5",
	}
	//
	l4 := list
	i4 := 5
	w4 := sort.StringSlice{
		"0", "1", "2", "3", "4",
	}
	//
	l5 := list
	i5 := 7
	w5 := list
	//
	l6 := sort.StringSlice{}
	i6 := 1
	w6 := sort.StringSlice{}
	//
	l7 := sort.StringSlice{}
	i7 := -1
	w7 := sort.StringSlice{}
	//

	tests := []struct {
		name string
		args args
		want sort.StringSlice
	}{
		// TODO: Add test cases.
		{" 1", args{l1, i1}, w1},
		{" 2", args{l2, i2}, w2},
		{" 3", args{l3, i3}, w3},
		{" 4", args{l4, i4}, w4},
		{" 5", args{l5, i5}, w5},
		{" 6", args{l6, i6}, w6},
		{" 7", args{l7, i7}, w7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeIndex(tt.args.aList, tt.args.aIndex); (0 < len(got)) && (0 < len(tt.want)) && (!reflect.DeepEqual(got, tt.want)) {
				t.Errorf("removeIndex() = »%v«,\nwant »%v«", got, tt.want)
			}
		})
	}
} // Test_removeIndex()

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

func Test_stat(t *testing.T) {
	const homeDir = "/home/matthias/devel/Go/src/github.com/mwat56/screenshot"
	n1 := "./file is missing"
	w1 := ""
	o1 := false
	//
	n2 := "./LICENSE"
	w2 := filepath.Join(homeDir, n2)
	o2 := true
	//
	d3 := []byte("# dummy")
	n3 := "./CrashTestDummy.txt" // >       -rw-r-----
	if err := os.WriteFile(n3, d3, fs.FileMode(0640)); nil != err {
		log.Println("Test_stat() error:", err)
	}
	defer func() {
		os.Remove(n3)
	}()
	w3 := filepath.Join(homeDir, n3)
	o3 := true

	n4 := "/dev/null"
	w4 := ""
	o4 := false
	//
	n5 := "/Delete.Me"
	w5 := ""
	o5 := false
	//

	tests := []struct {
		name      string
		aFilename string
		wantName  string
		wantOK    bool
	}{
		// TODO: Add test cases.
		{" 1", n1, w1, o1},
		{" 2", n2, w2, o2},
		{" 3", n3, w3, o3},
		{" 4", n4, w4, o4},
		{" 5", n5, w5, o5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := stat(tt.aFilename)
			if got != tt.wantName {
				t.Errorf("stat() got = »%v«,\nwant »%v«", got, tt.wantName)
			}
			if got1 != tt.wantOK {
				t.Errorf("stat() got1 = %v, want %v", got1, tt.wantOK)
			}
		})
	}
} // Test_stat()

func Test_writeFile(t *testing.T) {
	setupScreenshot()

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
	setupScreenshot()
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

func testSetHosts4JS(t *testing.T, aNameConstant string) {
	//
	a1 := ""
	w1 := "/home/matthias/devel/Go/src/github.com/mwat56/screenshot/" + aNameConstant
	//
	a2 := filepath.Join(os.TempDir(), aNameConstant)
	w2 := "" // file does not exist yet
	//
	a3 := "/dev/bla/bla"
	w3 := ""
	//
	a4 := "/tmp"
	w4 := ""
	//
	a5 := "dummy"
	w5 := ""
	//
	a6 := aNameConstant
	w6 := "/home/matthias/devel/Go/src/github.com/mwat56/screenshot/" + aNameConstant
	//
	tests := []struct {
		name      string
		aPathname string
		want      string
	}{
		// TODO: Add test cases.
		{" 1", a1, w1},
		{" 2", a2, w2},
		{" 3", a3, w3},
		{" 4", a4, w4},
		{" 5", a5, w5},
		{" 6", a6, w6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setHosts4JS(tt.aPathname, aNameConstant); got != tt.want {
				t.Errorf("setHosts4JS() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_setHosts4JS()

func TestSetAvoidJS(t *testing.T) {
	testSetHosts4JS(t, HostsAvoidJS)
} // TestSetHostsAvoidJS()

func TestSetHostsNeedJS(t *testing.T) {
	testSetHosts4JS(t, HostsNeedJS)
} // TestSetHostsNeedJS()

func TestSetup(t *testing.T) {
	setupScreenshot()

	o1 := Options() // nothing changes

	o2 := Options()
	o2.ImageAge = 0 // matches default value

	o3 := Options()
	o3.ImageDir = "/tmp" // matches default value

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
			if got := Setup(tt.aOptions); !reflect.DeepEqual(got, tt.aOptions) {
				t.Errorf("Setup() = »%v«,\nwant »%v«", got, tt.aOptions)
			}
		})
	}
} // TestSetup()

func TestString(t *testing.T) {
	setupScreenshot()
	w1 := `AcceptOther:	true
CertErrors:	false
Cookies:	false
HostsAvoidJS:	'/home/matthias/devel/Go/src/github.com/mwat56/screenshot/hostsavoidjs.list'
HostsNeedJS:	'/home/matthias/devel/Go/src/github.com/mwat56/screenshot/hostsneedjs.list'
ImageAge:	0
ImageDir:	'/tmp'
ImageHeight:	768
ImageOverwrite:	false
ImageQuality:	75
ImageScale:	0.99
ImageWidth:	896
JavaScript:	false
MaxProcessTime:	24
Mobile:	false
Platform:	'Linux x86_64'
Scrollbars:	true
UserAgent:	'Mozilla/5.0 (X11; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0'
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
