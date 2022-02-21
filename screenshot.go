/*
   Copyright © 2022 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/
package screenshot

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"golang.org/x/image/draw" // go get -u golang.org/x/image/draw
)

const (
	// Identifier used in error messages
	ssLibName = "ScreenShot"

	// Filename of the list of hosts/domains where JS should be avoided
	ssHostsAvoidJS = "hostsavoidjs.list"

	// Filename of the list of hosts/domains where JS is needed
	ssHostsNeedJS = "hostsneedjs.list"
)

// ScreenshotParams bundles all available configuration Options and
// pass them to the `Setup()` function in a single call.
type (
	ScreenshotParams struct {
		// Flag whether certificate errors should be ignored.
		CertErrors bool

		// Allow use of web cookies
		Cookies bool

		// Path/filename of a list of web hosts/domains where JavaScript
		// running should be avoided (defaults to a file in user's homedir).
		HostsAvoidJS string

		// Path/filename of a list of web hosts/domains where JavaScript
		// is required to work (defaults to a file in user's homedir).
		HostsNeedJS string

		// Max. age of cached page screenshot images (in hours).
		ImageAge int64

		// Directory to store the generated screenshot images.
		ImageDir string

		// Max. height of the screenshot image to generate.
		ImageHeight int

		// Quality (in percent) of the screenshot image to generate.
		ImageQuality int

		// The virtual browser's scale factor value.
		// 0 disables the override.
		ImageScale float64

		// Max. width of the screenshot image to generate.
		ImageWidth int

		// Flag whether to allow JavaScript in retrieved pages.
		JavaScript bool

		// Timeout (in seconds) for page processing.
		MaxProcessTime int64

		// Flag whether to emulate a mobile device or not.
		// This includes viewport meta tag, overlay scrollbars, text
		// autosizing and more.
		Mobile bool

		// The identifier the JavaScript `navigator.platform` should return.
		Platform string

		// Flag whether to show the scraped web-page's scrollbars.
		Scrollbars bool

		// User Agent to use when queuing external sites.
		UserAgent string
	}

	tAvoidNeedFile struct {
		// Time of next reading a Avoid/Need hosts file:
		nextTime time.Time

		// List of hosts to test against:
		list sort.StringSlice
	}
)

var (
	// List with sites to avoid JavaScript:
	ssAvoidJSsites tAvoidNeedFile = tAvoidNeedFile{
		nextTime: time.Now(),
	}

	// R/O RegEx to extract a filename's extension:
	ssExtRE = regexp.MustCompile(`(\.\w+)([\?\#].*)?$`)

	// Internal lookup table for image type and filename extension.
	ssImageTypes = map[bool]string{
		false: `png`,
		true:  `jpeg`,
	}

	// List with sites to use JavaScript:
	ssNeedJSsites tAvoidNeedFile = tAvoidNeedFile{
		nextTime: time.Now(),
	}

	// The initially used screenshot options:
	ssOptions *ScreenshotParams = &ScreenshotParams{
		CertErrors:     false,
		Cookies:        false,
		HostsAvoidJS:   setHosts4JS("./", ssHostsAvoidJS),
		HostsNeedJS:    setHosts4JS("./", ssHostsNeedJS),
		ImageAge:       0,
		ImageDir:       os.TempDir(),
		ImageHeight:    768,
		ImageQuality:   75,
		ImageScale:     0,
		ImageWidth:     896,
		JavaScript:     false,
		MaxProcessTime: 32,
		Mobile:         false,
		Platform:       "Linux x86_64",
		Scrollbars:     false,
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	}

	// Number of minutes to wait before re-reading Avoid/Need hosts files:
	ssReadWaitTime int64 = 1

	// R/O RegEx to find all non alpha/digits in URLs.
	ssReplaceNonAlphasRE = regexp.MustCompile(`\W+`)
)

// Options returns the currently configured screenshot options.
func Options() (rOptions *ScreenshotParams) {
	rOptions = &ScreenshotParams{
		CertErrors:     ssOptions.CertErrors,
		Cookies:        ssOptions.Cookies,
		HostsAvoidJS:   ssOptions.HostsAvoidJS,
		HostsNeedJS:    ssOptions.HostsNeedJS,
		ImageAge:       ssOptions.ImageAge,
		ImageDir:       ssOptions.ImageDir,
		ImageHeight:    ssOptions.ImageHeight,
		ImageQuality:   ssOptions.ImageQuality,
		ImageScale:     ssOptions.ImageScale,
		ImageWidth:     ssOptions.ImageWidth,
		JavaScript:     ssOptions.JavaScript,
		MaxProcessTime: ssOptions.MaxProcessTime,
		Mobile:         ssOptions.Mobile,
		Platform:       ssOptions.Platform,
		Scrollbars:     ssOptions.Scrollbars,
		UserAgent:      ssOptions.UserAgent,
	}

	return
} // Options()

// Setup uses `aOptions` to configure the runtime options for
// taking screenshots.
//
// NOTE: While it is perfectly legal (from Go's point of view) to omit
// those fields you don't care about please be aware that those missing
// fields will nevertheless be set (by `Go`): with the respective data
// type's default value.
// And since there's no way to distinguish the automatically set default
// value of a missing field from a user provided value you have to handle
// such a situation carefully.
// Depending on the number of options you want to set you might want to
// prefer calling the various `SetXxxx()` functions (if there are less
// than half of the available options to set). Or – if you want to set
// the majority of the options – you'd provide the options you do not
// want to change with their already existing values by calling the
// respective GETter function of the option in question, like:
//
//	myOptions := &ScreenshotParams{
//		// set fields …
//		ImageHeight:  myHeightValue,
//		ImageQuality: myQualityValue,
//		// …
//		// say, you don't want to change the width option
//		ImageWidth:   screenshot.ImageWidth(),
//	}
//	screenshot.Setup(myOptions)
//	// continue with your program …
//
//	`aOptions` The screenshot options to use.
func Setup(aOptions *ScreenshotParams) (rOptions *ScreenshotParams) {
	if *aOptions == *ssOptions {
		return Options() // nothing to change
	}

	ssOptions.CertErrors = aOptions.CertErrors
	ssOptions.Cookies = aOptions.Cookies
	SetHostsAvoidJS(aOptions.HostsAvoidJS)
	SetHostsNeedJS(aOptions.HostsNeedJS)
	SetImageAge(aOptions.ImageAge)
	SetImageDir(aOptions.ImageDir)
	SetImageHeight(aOptions.ImageHeight)
	SetImageQuality(aOptions.ImageQuality)
	SetImageScale(aOptions.ImageScale)
	SetImageWidth(aOptions.ImageWidth)
	ssOptions.JavaScript = aOptions.JavaScript
	SetMaxProcessTime(aOptions.MaxProcessTime)
	ssOptions.Mobile = aOptions.Mobile
	ssOptions.Platform = aOptions.Platform
	ssOptions.Scrollbars = aOptions.Scrollbars
	SetUserAgent(aOptions.UserAgent)

	return Options()
} // Setup()

// String returns a string of lines showing the currently configured
// screenshot options.
func String() string {
	const (
		fmtBoo = "%s:\t%t\n"
		fmtFlt = "%s:\t%.2f\n"
		fmtInt = "%s:\t%d\n"
		fmtStr = "%s:\t'%s'\n"
	)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(fmtBoo, "CertErrors", ssOptions.CertErrors))
	sb.WriteString(fmt.Sprintf(fmtBoo, "Cookies", ssOptions.Cookies))
	sb.WriteString(fmt.Sprintf(fmtStr, "HostsAvoidJS", ssOptions.HostsAvoidJS))
	sb.WriteString(fmt.Sprintf(fmtStr, "HostsNeedJS", ssOptions.HostsNeedJS))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageAge", ssOptions.ImageAge))
	sb.WriteString(fmt.Sprintf(fmtStr, "ImageDir", ssOptions.ImageDir))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageHeight", ssOptions.ImageHeight))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageQuality", ssOptions.ImageQuality))
	sb.WriteString(fmt.Sprintf(fmtFlt, "ImageScale", ssOptions.ImageScale))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageWidth", ssOptions.ImageWidth))
	sb.WriteString(fmt.Sprintf(fmtBoo, "JavaScript", ssOptions.JavaScript))
	sb.WriteString(fmt.Sprintf(fmtInt, "MaxProcessTime", ssOptions.MaxProcessTime))
	sb.WriteString(fmt.Sprintf(fmtBoo, "Mobile", ssOptions.Mobile))
	sb.WriteString(fmt.Sprintf(fmtStr, "Platform", ssOptions.Platform))
	sb.WriteString(fmt.Sprintf(fmtBoo, "Scrollbars", ssOptions.Scrollbars))
	sb.WriteString(fmt.Sprintf(fmtStr, "UserAgent", ssOptions.UserAgent))

	return sb.String()
} // String()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `chk4()` checks for a match of `aURL` in hosts list `aHostsFile`
//
// NOTE: To determine which hosts file to use the `aHostsFile` argument
// is tested for the (fixed) filename/extension.
//
//	`aURL` The URL to check for matching an entry in `aHostsFile`.
//	`aHostsFilename` The Avoid/Need list path/file to read from disk.
//	`aList` The internal string list to hold the lines of `aHostsFile`.
func chk4(aURL, aHostsFilename string) bool {
	var (
		err    error
		hosts  *tAvoidNeedFile
		needle string
		URL    *url.URL
	)

	// We can't use `switch` here since the order of tests is
	// significant (which is guaranteed with `switch`).
	if (0 == len(aHostsFilename)) || (0 == len(aURL)) {
		return false
	}
	if strings.HasSuffix(aHostsFilename, ssHostsAvoidJS) {
		hosts = &ssAvoidJSsites
	} else if strings.HasSuffix(aHostsFilename, ssHostsNeedJS) {
		hosts = &ssNeedJSsites
	} else {
		return false // unrecognised filename
	}

	if URL, err = url.Parse(aURL); nil != err {
		return false
	}
	if needle = URL.Hostname(); 0 == len(needle) {
		// The given `aURL` is obviously not a full/correct URL
		// but probably just a host name.
		if needle = URL.Path; 0 == len(needle) {
			return false
		}
	}

	if (0 == hosts.list.Len()) || time.Now().After(hosts.nextTime) {
		hosts.nextTime = time.Now().Add(time.Duration(ssReadWaitTime) * time.Minute)
		if hosts.list = readListFile(aHostsFilename); 0 == hosts.list.Len() {
			return false
		}
	}

	return containsHost(strings.ToLower(needle), &hosts.list)
} // chk4()

// `cleanupOutput()` removes unneeded leading data from `aRawData`
// and returns the properly encoded image.
//
//	`aRawData` The raw image data to cleanup.
func cleanupOutput(aRawData []byte) []byte {
	if 0 == len(aRawData) {
		return aRawData
	}
	var (
		buffer  bytes.Buffer
		decoded image.Image
		err     error
	)

	if 100 == ssOptions.ImageQuality { // 'png' format
		decoded, err = png.Decode(bytes.NewReader(aRawData))
		for nil != err {
			if aRawData = aRawData[1:]; 0 == len(aRawData) {
				return aRawData // i.e. empty array
			}
			decoded, err = png.Decode(bytes.NewReader(aRawData))
		}
		decoded = crop(decoded) // adjust the image's size
		_ = png.Encode(&buffer, decoded)
	} else { // 'jpeg' format
		decoded, err = jpeg.Decode(bytes.NewReader(aRawData))
		for nil != err {
			if aRawData = aRawData[1:]; 0 == len(aRawData) {
				return aRawData // i.e. empty array
			}
			decoded, err = jpeg.Decode(bytes.NewReader(aRawData))
		}
		decoded = crop(decoded) // adjust the image's size
		opts := jpeg.Options{Quality: ssOptions.ImageQuality}
		_ = jpeg.Encode(&buffer, decoded, &opts)
	}

	if 4096 < buffer.Len() {
		return buffer.Bytes()
	}

	return aRawData // i.e. original data
} // cleanupOutput()

// `configure()` sets up how to take a screenshot of the entire browser
// viewport the size of which is determined by `ImageWidth`/`ImageHeight`.
//
//	`aURL` The address of the web page to process.
//	`aResult` Data structure to receive the generated screenshot image.
func configure(aURL string, aResult *[]byte) chromedp.Tasks {
	enableJS := ssOptions.JavaScript
	if enableJS {
		// If the domain is found in the 'avoid' list then we
		// do NOT want to activate JS here:
		enableJS = !chk4(aURL, ssOptions.HostsAvoidJS)
	} else {
		// If the domain is found in the 'need' list then we DO
		// want to activate JS here:
		enableJS = chk4(aURL, ssOptions.HostsNeedJS)
	}
	waitDuration := time.Second << 1 // two seconds
	if enableJS {
		waitDuration <<= 1 // four seconds
	}

	// Note: `chromedp.FullScreenshot()` overrides the device's
	// emulation settings.
	// Use `device.Reset` to reset the emulation and viewport settings.
	return chromedp.Tasks{
		// ensure basic setup:
		chromedp.Emulate(device.Reset),
		emulation.ClearDeviceMetricsOverride(),
		emulation.ClearGeolocationOverride(),
		emulation.ResetPageScaleFactor(),

		// values of '0' will disable the override:
		emulation.SetDeviceMetricsOverride(int64(ssOptions.ImageWidth), int64(ssOptions.ImageHeight), ssOptions.ImageScale, ssOptions.Mobile),

		// setup some browser options:
		emulation.SetDocumentCookieDisabled(!ssOptions.Cookies),
		emulation.SetEmitTouchEventsForMouse(false),
		emulation.SetFocusEmulationEnabled(false),
		emulation.SetScriptExecutionDisabled(!enableJS),
		emulation.SetScrollbarsHidden(!ssOptions.Scrollbars),
		// ignore certificate errors (e.g. self-signed):
		security.SetIgnoreCertificateErrors(!ssOptions.CertErrors),
		// configure the UserAgent to pose as:
		emulation.SetUserAgentOverride(ssOptions.UserAgent).
			// WithAcceptLanguage("en").	//FIXME get proper value format
			WithPlatform(ssOptions.Platform),

		// perform the actual scraping action:
		chromedp.Navigate(aURL),
		chromedp.Sleep(waitDuration), // time to receive&render the page
		chromedp.FullScreenshot(aResult, ssOptions.ImageQuality),
	}
} // configure()

// `containsHost()` returns whether `aNeedle` matches a line
// in `aHaystack`.
//
// NOTE: This function doesn't check whether `aNeedle` is literally
// found in `aHaystack` but checks whether `aNeedle` _ends_ with an
// entry in `aHaystack`.
// So both `example.com` and `mobile.example.com`(as needle) are matched
// by the line `example.com` (in haystack) while only `mobile.example.com`
// (as needle) would be matched by `.example.com` (in haystack).
//
//	`aNeedle` The string to search in the provided list
//	`aHaystack` The string list to walk through.
func containsHost(aNeedle string, aHaystack *sort.StringSlice) bool {
	for _, entry := range *aHaystack {
		if (0 == len(entry)) || `#` == entry[0:1] {
			continue // shouldn't happen: `readListFile()` removes
			// those lines, but UnitTests might send such lists.
		}
		if strings.HasSuffix(aNeedle, entry) {
			return true
		}
	}

	return false
} // containsHost()

// `crop()` Adjusts the image's size to the configured
// `ImageWidth`/`ImageHeight` values.
//
//	`aImgData` The raw image data to crop.
func crop(aImgData image.Image) image.Image {
	bounds := aImgData.Bounds()
	doCrop := false
	doMagnify := false
	size := bounds.Size()
	xIsBigger := (0 < ssOptions.ImageWidth) && (size.X > ssOptions.ImageWidth)
	yIsBigger := (0 < ssOptions.ImageHeight) && (size.Y > ssOptions.ImageHeight)

	if xIsBigger {
		size.X = ssOptions.ImageWidth
		doCrop = true
	} else if size.X < ssOptions.ImageWidth {
		doMagnify = true
	}
	if yIsBigger {
		size.Y = ssOptions.ImageHeight
		doCrop = true
	} else if size.Y < ssOptions.ImageHeight {
		doMagnify = true
	}

	if doCrop {
		// Either width or height or both are greater than
		// the wanted/configured max. dimensions and are done.
		if yIsBigger {
			// We just cut off the part outside (below)
			// our wanted/configured height.
			return aImgData.(interface {
				SubImage(aRect image.Rectangle) image.Image
			}).
				SubImage(image.Rect(0, 0, size.X, size.Y))
		}

		// Set the configured size:
		result := image.NewRGBA(image.Rect(0, 0, ssOptions.ImageWidth, size.Y))

		// Perform the actual shrinking:
		draw.BiLinear.Scale(result, result.Rect, aImgData, bounds, draw.Over, nil)

		return result
	}

	if doMagnify {
		// Set the configured size:
		result := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

		// Do the actual enlarging:
		draw.BiLinear.Scale(result, result.Rect, aImgData, bounds, draw.Over, nil)

		return result
	}

	return aImgData // unmodified image
} // crop()

// `exists()` returns whether there is an usable image file cached.
//
// This function uses the `ImageAge()` value to determine whether
// an already existing local file is considered to be too old.
//
// Files empty or smaller than 10KB are ignored.
func exists(aFilename string) bool {
	if 0 == ssOptions.ImageAge {
		// shortcut: no checks at all …
		return false
	}

	fi, err := os.Stat(aFilename)
	if (nil != err) || fi.IsDir() {
		return false
	}

	if 8192 > fi.Size() {
		// Empty and small (i.e. `<8KB`) files are ignored.
		// File sizes smaller than ~10KB indicate some kind of
		// error during retrieval of the web page or rendering it.
		// Valid preview images take approximately between 10 to
		// 100 KB depending on the respective web page (e.g. number
		// and size of embedded images).
		return false
	}

	maxTime := fi.ModTime().Add(time.Duration(ssOptions.ImageAge) * time.Hour)
	// files too old are ignored
	return time.Now().Before(maxTime)
} // exists()

// `fileExt()` returns the filename extension of `aURL`.
//
//	`aURL` The URL to process.
func fileExt(aURL string) string {
	result := ssExtRE.FindStringSubmatch(aURL)
	if 1 < len(result) {
		return result[1]
	}

	return ""
} // fileExt()

// `generateImage()` creates an image from `aURL`.
// It returns the image data and any error encountered.
//
//	`aContext` The active context to use.
//	`aURL` The remote URL to be handled.
func generateImage(aContext context.Context, aURL string) (rImage []byte, rErr error) {
	var rawData []byte

	ctx, cancel := chromedp.NewContext(aContext,
		chromedp.WithLogf(log.Printf),
		// chromedp.WithRunnerOptions(runner.Flag("ignore-certificate-errors", "1")),
	)

	defer func() {
		// `chromedp.FullScreenshot()` might panic :-((
		if r := recover(); nil != r {
			if nil == rErr {
				rErr = errors.New(ssLibName +
					": error reading '" + aURL + "'")
			}
			log.Println(ssLibName, rErr)
		}
		cancel()
	}()

	// Capture the entire browser viewport
	if rErr = chromedp.Run(ctx, configure(aURL, &rawData)); nil != rawData {
		if nil != rErr {
			log.Println(ssLibName, ":", aURL, ImageType(), ssOptions.ImageQuality, rErr)
		}
		if rImage = cleanupOutput(rawData); 4096 < len(rImage) {
			rErr = nil
		}
	}

	return
} // generateImage()

// `readListFile()` reads the named text file and returns its lines
// as a list of strings.
//
// NOTE: The resulting list may contain empty lines.
// All lines in the list are lowercased and trimmed.
//
//	`aFilename` The name of the file to read.
func readListFile(aFilename string) (rList sort.StringSlice) {
	if 0 == len(aFilename) {
		return
	}

	data, err := os.ReadFile(aFilename)
	if (err != nil) || (0 == len(data)) {
		return
	}
	rList = strings.Split(string(data), "\n")

	defer func() {
		if r := recover(); nil != r {
			log.Println("caught panic", rList, r)
		}
	}()

	syncIdx := 0
reStart:
	for idx, line := range rList {
		if idx < syncIdx {
			continue // skip: go to next line
		} else if idx > syncIdx {
			syncIdx = idx
		}

		// Make sure there are no "\t" or "\r" left and
		// everything is in lowercase letters which is
		// expected by `chk4()`/`containsHost()`.
		line = strings.ToLower(strings.TrimSpace(line))
		if (0 == len(line)) || (`#` == line[0:1]) {
			rList = removeIndex(rList, idx)
			goto reStart // restart the loop
		}
		rList[idx] = line
	}
	// rList.Sort() // not needed: list is read sequentially anyway

	return
} // readListFile()

// `removeIndex()` deletes the element at `aIndex` from `aList`
// returning a shortened list.
//
// This function does not change `aList` "in place" but returns a new list.
//
//	`aList` The string list to handle.
//	`aIndex` The list index to remove.
func removeIndex(aList sort.StringSlice, aIndex int) sort.StringSlice {
	// Working with the `append` function on slices without taking care
	// of the origin and destination of the values we are dealing with.
	// Since we do _not_ want to modify the given `aList` we create a
	// new list to return to the caller.
	// That way even if the caller later modifies this function's result
	// the original (`aList`) values remain as-is.
	var result sort.StringSlice
	lastIdx := len(aList) - 1 // index of the list's last element

	// We can't use a `switch` statement here because the order of
	// tests is significant (but not guaranteed with `switch/case`).
	if 0 > lastIdx {
		// return a new empty list
		return result
	}

	if aIndex > lastIdx {
		// copy the whole list:
		return append(result, aList[:]...)
	}

	if 0 == aIndex {
		// skip the very first list element:
		return append(result, aList[1:]...)
	}

	if aIndex == lastIdx {
		// omit the last list element:
		return append(result, aList[:lastIdx]...)
	}

	// return a list without the `aIndex` element in `aList`:
	return append(append(result, aList[:aIndex]...), aList[aIndex+1:]...)
} // removeIndex()

// `sanitise()` returns `aURL` with all non alpha/digits removed.
// The resulting string can then be used as the screenshot's file name.
//
//	`aURL` The URL to sanitise.
func sanitise(aURL string) string {
	return ssReplaceNonAlphasRE.ReplaceAllLiteralString(aURL, ``)
} // sanitise()

// `setHosts4JS()` configures the name of the file containing
// hosts/domains where JavaScript should be disabled/active.
//
// NOTE: This function avoids code duplication and is called internally by
// `SetHostsAvoidJS()` and `SetHostsNeedJS()` to check whether `aPathname`
// exists and is readable.
//
//	`aPathname` The path/filename of sites' list.
//	`aNameConstant` The list file's constant name.
func setHosts4JS(aPathname, aNameConstant string) string {
	if aPathname = strings.TrimSpace(aPathname); 0 == len(aPathname) {
		aPathname = "./" + aNameConstant
	}

	if !strings.HasSuffix(aPathname, aNameConstant) {
		aPathname = filepath.Join(aPathname, ".", aNameConstant)
	}

	if fName, ok := stat(aPathname); ok {
		// if err := syscall.Access(fName, syscall.O_RDONLY); nil == err {
		return fName
		// }
	}

	return ""
} // setHosts4JS()

// `stat()` checks whether `aFilename` points to a valid path/file
// which exists and is readable.
//
//	`aFilename` The filename to check.
func stat(aFilename string) (string, bool) {
	var ( // separate declaration for better debugging
		err   error
		fi    os.FileInfo
		fName string
	)

	if fName, err = filepath.Abs(aFilename); nil == err {
		if fi, err = os.Stat(fName); (nil == err) && (!fi.IsDir()) && (0 < fi.Size()) {
			if err = syscall.Access(fName, syscall.O_RDONLY); nil == err {
				return fName, true
			}
		}
	}

	return "", false
} // stat()

// 'writeFile()' stores the given image data to a file, returning an
// error in case of problems.
//
//	`aFilename` The path/file name to use for storing the image.
//	`aData` The image data to store.
//	`aResponse` A (possibly NIL) response data from downloading an image file.
func writeFile(aFilename string, aData []byte, aResponse *http.Response) (rErr error) {
	if 0 == len(aFilename) {
		return errors.New(ssLibName + ": empty file name argument")
	}

	var file *os.File

	if file, rErr = os.OpenFile(aFilename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.FileMode(0640)); nil != rErr { /* #-nosec G302 */
		return
	}
	defer file.Close()

	if 0 < len(aData) {
		_, rErr = file.Write(aData)
	} else if (nil != aResponse) && (0 < aResponse.ContentLength) {
		_, rErr = io.Copy(file, aResponse.Body)
	} else {
		_ = os.Remove(aFilename)
		rErr = errors.New(ssLibName + ": no image data to write '" + aFilename + "'")
	}

	if nil != rErr {
		// In case of errors during write we delete the file
		// ignoring possible errors here and return the error.
		_ = os.Remove(aFilename)
	}

	return
} // writeFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
/*                           public functions                              */

// CertErrors returns whether to skip sites with Certificate errors;
// defaults to `false` for historic reasons.
func CertErrors() bool {
	return ssOptions.CertErrors
} // CertErrors()

// SetCertErrors determines whether to skip sites with certificate errors.
//
//	`anAllow` If `false` (i.e. the default) all web-sites will be processed
// regardless of certificate errors.
func SetCertErrors(anAllow bool) {
	ssOptions.CertErrors = anAllow
} // SetCertErrors()

// Cookies returns whether to allow web cookies during page retrieval;
// defaults to `false` for safety and speed reasons.
func Cookies() bool {
	return ssOptions.Cookies
} // Cookies()

// SetCookies determines whether to allow web cookies during page
// retrieval or not.
//
//	`anAllow` If `false` (i.e. the default) no cookies will be available
// during page retrieval, otherwise (i.e. `true`) they will be used.
func SetCookies(anAllow bool) {
	ssOptions.Cookies = anAllow
} // SetCookies()

// CreateImage generates an image of `aURL` and stores it in `ImageDir()`,
// returning the file name of the saved image or an error in case of problems.
//
//	`aURL` The address of the web page to process.
func CreateImage(aURL string) (string, error) {
	//TODO add 'context' argument

	if 0 == len(ssOptions.ImageDir) {
		return "", errors.New(ssLibName + ": property 'ImageDir' is empty")
	}

	result := sanitise(aURL) + `.` + ssImageTypes[100 > ssOptions.ImageQuality]
	fName := filepath.Join(ssOptions.ImageDir, result)
	// Check whether we've already got an image file
	// so we might avoid additional network traffic:
	if exists(fName) {
		return result, nil
	}

	var (
		// Declare variables here so we can use them in different
		// contexts/closures below (and it eases debugging).
		cancel    context.CancelFunc
		ctx       context.Context
		err       error
		ext       string
		imageData []byte
		response  *http.Response
	)

	//TODO turn `Background()` into calltime argument:
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(ssOptions.MaxProcessTime)*time.Second)
	defer func() {
		if r := recover(); nil != r {
			// Timing problems or invalid site data might indirectly
			// cause the image generation to panic.
			log.Println(ssLibName, err)
		}
		cancel()
	}()

	// Exclude certain filetypes from preview generation:
	ext = strings.ToLower(fileExt(aURL))
	switch ext {
	case ".amr", ".arj", ".avi", ".azw3",
		".bak", ".bibtex", ".bz2",
		".cfg", ".com", ".conf", ".csv",
		".db", ".deb", ".doc", ".docx", ".dia",
		".epub", ".exe", ".flv", ".gz",
		".ics", ".iso", ".jar", ".json",
		".md", ".mobi", ".mp3", ".mp4", ".mpeg",
		".odf", ".odg", ".odp", ".ods", ".odt", ".otf", ".oxt",
		".pas", ".pdf", ".ppd", ".ppt", ".pptx",
		".rip", ".rpm", ".spk", ".sxg", ".sxw",
		".ttf", ".vbox", ".vmdk", ".vcs", ".wav",
		".xls", ".xpi", ".xsl", ".zip":
		return "", errors.New(ssLibName +
			": excluded filename extension '" + ext + "'")

	case ".gif", ".jpeg", ".jpg", ".png", ".svg":
		if response, err = http.Get(aURL); /* #nosec G107 */ nil != err {
			return "", err
		}
		defer response.Body.Close()
		result = sanitise(aURL) + ext
		fName = filepath.Join(ssOptions.ImageDir, result)

	default:
		if imageData, err = generateImage(ctx, aURL); nil != err {
			return "", err
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err() // Canceled? TimeOut?
		default:
			break
		}
	}

	if (0 == len(imageData)) && (nil == response) {
		return "", errors.New(ssLibName +
			": no data received for '" + fName + "'")
	}

	if err = writeFile(fName, imageData, response); nil != err {
		// some problem during attempt to save image to disk
		return "", err
	}

	// Everything went well it seems …
	return result, nil
} // CreateImage()

// HostsAvoidJS returns the name of the path/file containing
// hosts/domains where to avoid running JavaScript.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
func HostsAvoidJS() string {
	return ssOptions.HostsAvoidJS
} // HostsAvoidJS()

// SetHostsAvoidJS configures the name of the file containing
// hosts/domains where to avoid running JavaScript.
//
// NOTE: This value is used only if the `JavaScript` option is set `true`.
// An invalid filename disables the feature.
//
//	`aFilename` The path/filename of sites with JavaScript to avoid.
func SetHostsAvoidJS(aFilename string) {
	ssOptions.HostsAvoidJS = setHosts4JS(aFilename, ssHostsAvoidJS)
} // SetHostsAvoidJS()

// HostsNeedJS returns the name of the path/file containing
// hosts/domains requiring JavaScript to be active/working.
//
// NOTE: This value is used only if the `JavaScript()` option is set `false`.
func HostsNeedJS() string {
	return ssOptions.HostsNeedJS
} // HostsNeedJS()

// SetHostsNeedJS configures the name of the file containing
// hosts/domains requiring JavaScript to be active/working.
//
// NOTE: This value is used only if the `JavaScript()` option is set `false`.
// An invalid filename disables the feature.
//
//	`aFilename` The path/filename of sites with required JavaScript.
func SetHostsNeedJS(aFilename string) {
	ssOptions.HostsNeedJS = setHosts4JS(aFilename, ssHostsNeedJS)
} // SetHostsNeedJS()

// ImageAge returns the maximum age (in hours) of the locally stored
// screenshot images.
func ImageAge() int64 {
	return ssOptions.ImageAge
} // ImageAge()

// SetImageAge sets the maximum age of locally stored screenshot images
// before they may get updated by a new call to `CreateImage(…)`.
//
// Usually you'll want this property at its default value (`0`, zero)
// which disables an age check because usually you want an image of the
// page at the time you linked to it.
//
//	`aMaxAge` is the age (in hours) a page image can have before
// requesting it again.
func SetImageAge(aMaxAge int64) {
	if 0 < aMaxAge {
		ssOptions.ImageAge = aMaxAge
	} else {
		ssOptions.ImageAge = 0
	}
} // SetImageAge()

// ImageDir returns the directory to store the generated screenshot images.
func ImageDir() string {
	return ssOptions.ImageDir
} // ImageDir()

// SetImageDir sets the directory to use for storing the generated
// screenshot images.
//
// If `aDirectory` is empty or invalid the system's temp directory is used.
//
//	`aDirectory` The directory to store the generated images.
func SetImageDir(aDirectory string) {
	if aDirectory = strings.TrimSpace(aDirectory); 0 == len(aDirectory) {
		// may be not writeable for current user (like /usr/bin/…):
		// aDirectory, _ = os.Getwd()
		aDirectory = os.TempDir() // the system's temp directory
	}

	dir, err := filepath.Abs(aDirectory)
	if (nil != err) || (0 == len(dir)) {
		// dir, _ = filepath.Abs("./") // see comment above ^^^
		dir = os.TempDir() // the system's temp directory
	}

	ssOptions.ImageDir = dir
} // SetImageDir()

// ImageHeight is the max. height of the virtual screen used to render.
// The default value is `768`.
//
// NOTE: This is the max. height of the screenshot.
// Depending on the actual web-site and its rendering by the running
// 'Chrome' instance the generated image's height could be less.
//
// The value `0` (zero) renders the entire page top to bottom,
// calculating the actual height from the page content.
func ImageHeight() int {
	return ssOptions.ImageHeight
} // ImageHeight()

// SetImageHeight sets the height in pixels of the screenshot images
// to generate.
// The default value is `768`.
//
// NOTE: This is the max. height of the screenshot.
// Depending on the actual web-site and its rendering by the running
// `Chrome` instance the generated image's height could be less.
//
// Setting this value to `0` will result in an image containing the
// whole web-page (which might be quite lang); so the actual height
// of the generated screenshot would be unpredictable.
//
//	`aHeight` The new height of the images to generate.
func SetImageHeight(aHeight int) {
	if 0 < aHeight {
		ssOptions.ImageHeight = aHeight
	} else {
		ssOptions.ImageHeight = 0
	}
} // SetImageHeight()

// ImageQuality returns the desired image quality.
func ImageQuality() int {
	return ssOptions.ImageQuality
} // ImageQuality

// SetImageQuality changes the quality of the screenshot image to be
// generated.
// Values supported between `1` and `100`; default is `75`.
//
//	`aQuality` the new desired image quality.
func SetImageQuality(aQuality int) {
	if (0 < aQuality) && (101 > aQuality) {
		ssOptions.ImageQuality = aQuality // 'jpeg' format
	} else {
		ssOptions.ImageQuality = 100 // i.e. 'png' format
	}
} // SetImageQuality()

// ImageScale returns the virtual browser's scale factor for
// the generated screenshot image.
func ImageScale() float64 {
	return ssOptions.ImageScale
} // ImageScale()

// SetImageScale sets the virtual browser's scale factor for
// the generated screenshot image.
//
//	`aFactor` the new scale factor; `0` disables scaling.
func SetImageScale(aFactor float64) {
	if 0 < aFactor {
		ssOptions.ImageScale = aFactor
	} else {
		ssOptions.ImageScale = 0
	}
} // SetImageScale()

// ImageType returns the type/format of the screenshot file generated.
//
// NOTE: The image type/format depends on the given `ImageQuality()`:
// `quality == 100` results in a `png` image,
// `quality < 100` results in a `jpeg` image.
//
// If the URL to shoot points to an image file
// (".gif", ".jpeg", ".jpg", ".png", ".svg")
// the result of this function might be wrong because the actually
// generated image depends on the type of the requested image.
func ImageType() string {
	return ssImageTypes[100 > ssOptions.ImageQuality]
} // ImageType()

// ImageWidth is the width in pixels of the imaginary screen used to render.
// The default value is `1024`.
//
// NOTE: This is the max. width of the screenshot.
// Depending on the actual web-site and its rendering by the running
// 'Chrome' instance the generated image could be smaller.
func ImageWidth() int {
	return ssOptions.ImageWidth
} // ImageWidth()

// SetImageWidth sets the width of the images to generate.
// The default value is `1024`.
//
// NOTE: This is the max. width of the screenshot.
// Depending on the actual web-site and its rendering by the running
// 'Chrome' instance the generated image could be smaller.
//
//	`aWidth` The new width of the images to generate.
func SetImageWidth(aWidth int) {
	if 0 < aWidth {
		ssOptions.ImageWidth = aWidth
	} else {
		ssOptions.ImageWidth = 0
	}
} // SetImageWidth()

// JavaScript returns whether to allow JavaScript during page retrieval;
// defaults to `false` for safety and speed reasons.
func JavaScript() bool {
	return ssOptions.JavaScript
} // JavaScript()

// SetJavaScript determines whether to allow JavaScript during page
// retrieval or not.
//
//	`anAllow` If `false` (i.e. the default) no JavaScript will be available
// during page retrieval, otherwise (i.e. `true`) it will be activated.
func SetJavaScript(anAllow bool) {
	ssOptions.JavaScript = anAllow
} // SetJavaScript()

// MaxProcessTime returns the timeout (in seconds) used to
// retrieve & render a requested web page.
func MaxProcessTime() int64 {
	return ssOptions.MaxProcessTime
} // MaxProcessTime()

// SetMaxProcessTime sets the timeout used to retrieve & render
// a requested web page.
//
// NOTE: A wrong (i.e. negative) value and `0` (zero) resets the
// timeout value to its default od 32 seconds.
//
//	`aProcessTime` The new max. seconds allowed to process a web page.
func SetMaxProcessTime(aProcessTime int64) {
	if 0 < aProcessTime {
		ssOptions.MaxProcessTime = aProcessTime
	} else {
		ssOptions.MaxProcessTime = 32
	}
} // SetMaxProcessTime()

// Mobile returns whether the virtual browser should emulate a mobile
// device.
func Mobile() bool {
	return ssOptions.Mobile
} // Mobile()

// SetMobile set whether to emulate mobile device.
// This includes viewport meta tag, overlay scrollbars, text
// autosizing and more.
//
//	`aForceMobile`
func SetMobile(aForceMobile bool) {
	ssOptions.Mobile = aForceMobile
} // setMobile()

// PathFile returns the complete local path/file of `aURL`.
//
// NOTE: This function does not check whether the file for `aURL`
// actually exists in the local filesystem.
//
//	`aURL` The address of the web page to process.
func PathFile(aURL string) string {
	return filepath.Join(ssOptions.ImageDir,
		sanitise(aURL)+`.`+ssImageTypes[100 > ssOptions.ImageQuality])
} // PathFile()

// Platform returns the platform `navigator.platform` should return.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
func Platform() string {
	return ssOptions.Platform
} // Platform()

// SetPlatform sets the platform `navigator.platform` should return.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
//
//	`aPlatform` The platform identifier to use for `navigator.platform`.
func SetPlatform(aPlatform string) {
	ssOptions.Platform = aPlatform
} // SetPlatform()

// ReadWaitTime returns the number of minutes to wait before an Avoid/Need
// hosts file is re-read.
func ReadWaitTime() int64 {
	return ssReadWaitTime
} // ReadWaitTime()

// SetReadWaitTime sets the number of minutes to wait before an Avoid/Need
// hosts file is re-read.
//
// Usually you'll want this property at its default value (`1`, one)
// which seems to be a reasonable compromise between batch processing
// (i.e. looping through a list of URLs to process) and mitigation of
// disk access.
//
//	`aMinutes` is the number of minutes to wait before an Avoid/Need
// hosts file is re-read.
func SetReadWaitTime(aMinutes int64) {
	if 0 < aMinutes {
		ssReadWaitTime = aMinutes
	} else {
		ssReadWaitTime = 0
	}
} // SetReadWaitTime()

// Scrollbars returns whether the virtual browser will show scrollbars
// (if available in web-page).
func Scrollbars() bool {
	return ssOptions.Scrollbars
} // Scrollbars()

// SetScrollbars sets whether the virtual browser will show scrollbars
// (if available in web-page).
//
//	`aScrollbar` Flag whether to show scrollbars (if available).
func SetScrollbars(aScrollbar bool) {
	ssOptions.Scrollbars = aScrollbar
} // SetScrollbars()

// UserAgent returns the current `User Agent` setting.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
func UserAgent() string {
	return ssOptions.UserAgent
} // UserAgent()

// SetUserAgent changes the current `User Agent` setting to `aAgent`.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
//
//	`aAgent` The new `User Agent` setting.
func SetUserAgent(aAgent string) {
	aAgent = strings.TrimSpace(aAgent)
	if 0 < len(aAgent) {
		ssOptions.UserAgent = aAgent
	} else {
		ssOptions.UserAgent = ""
	}
} // SetUserAgent()

/* _EoF_ */
