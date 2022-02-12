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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"golang.org/x/image/draw" // go get -u golang.org/x/image/draw
)

const (
	// Identifier used in error messages
	libName = "ScreenShot"

	// Default name of the list of hosts/domains
	defJsWhiteName = `jswhite.lst`
)

// ScreenshotParams bundles all available configuration Options and
// pass them to the `Setup()` function in a single call.
type ScreenshotParams struct {

	// Flag whether certificate errors should be ignored.
	CertErrors bool

	// Allow use of web cookies
	Cookies bool

	// Max. age of cached page screenshot images (in seconds).
	ImageAge time.Duration

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

	// Path/filename of a list of web hosts/domains where JavaScript
	// is required to work.
	WhiteJS string
}

var (
	// R/O RegEx to extract a filename's extension.
	ssExtRE = regexp.MustCompile(`(\.\w+)([\?\#].*)?$`)

	// Internal lookup table for image type and filename extension.
	ssImageTypes = map[bool]string{
		false: `png`,
		true:  `jpeg`,
	}

	// The actually used screenshot options:
	ssOptions *ScreenshotParams = &ScreenshotParams{
		Cookies:  false,
		ImageAge: 0,
		ImageDir: os.TempDir(),
		/*
			func() string {
				dir, _ := filepath.Abs("./")
				return dir
			}(),
		*/
		ImageHeight:  768,
		ImageQuality: 100,
		ImageScale:   0,
		ImageWidth:   896,
		JavaScript:   false,
		Mobile:       false,
		Platform:     "Linux x86_64",
		CertErrors:   false,
		Scrollbars:   false, //FIXME EXPERIMENTAL
		UserAgent:    "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
		WhiteJS: func() string {
			dir, _ := filepath.Abs("./" + defJsWhiteName)
			return dir
		}(),
	}

	// R/O RegEx to find all non alpha/digits in URLs.
	ssReplaceNonAlphasRE = regexp.MustCompile(`\W+`)

	// Timeout duration (seconds) for page processing
	ssTimeoutDuration = time.Second << 5 // i.e. 32 seconds
)

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
// want to change with their already existing default values by calling
// the respective Getter function of the option in question, like:
//
//	myOptions := &ScreenshotParams{
//		// set fields …
//		ImageHeight: myHeightValue,
//		ImageQuality:  myQualityValue,
//		// …
//		// say, you don't want to change the width option
//		ImageWidth:  screenshot.ImageWidth(),
//	}
//	screenshot.Setup(myOptions)
//	// continue with your program …
//
//	`aOptions` The screenshot options tu use.
func Setup(aOptions *ScreenshotParams) {
	if *aOptions == *ssOptions {
		return // nothing to change
	}

	ssOptions.Cookies = aOptions.Cookies
	SetImageAge(aOptions.ImageAge)
	SetImageDir(aOptions.ImageDir)
	SetImageHeight(aOptions.ImageHeight)
	SetImageQuality(aOptions.ImageQuality)
	SetImageWidth(aOptions.ImageWidth)
	ssOptions.JavaScript = aOptions.JavaScript
	ssOptions.Mobile = aOptions.Mobile
	ssOptions.Platform = aOptions.Platform
	SetImageScale(aOptions.ImageScale)
	// ssOptions.Security = aOptions.Security
	ssOptions.Scrollbars = aOptions.Scrollbars
	SetUserAgent(aOptions.UserAgent)
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
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageAge", ssOptions.ImageAge))
	sb.WriteString(fmt.Sprintf(fmtStr, "ImageDir", ssOptions.ImageDir))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageHeight", ssOptions.ImageHeight))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageQuality", ssOptions.ImageQuality))
	sb.WriteString(fmt.Sprintf(fmtFlt, "ImageScale", ssOptions.ImageScale))
	sb.WriteString(fmt.Sprintf(fmtInt, "ImageWidth", ssOptions.ImageWidth))
	sb.WriteString(fmt.Sprintf(fmtBoo, "JavaScript", ssOptions.JavaScript))
	sb.WriteString(fmt.Sprintf(fmtBoo, "Mobile", ssOptions.Mobile))
	sb.WriteString(fmt.Sprintf(fmtStr, "Platform", ssOptions.Platform))
	sb.WriteString(fmt.Sprintf(fmtBoo, "Scrollbars", ssOptions.Scrollbars))
	sb.WriteString(fmt.Sprintf(fmtStr, "UserAgent", ssOptions.UserAgent))

	return sb.String()
} // String()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

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
// viewport the size of which is determined by `ImageWidth()`/`ImageHeight()`.
//
//	`aURL` The address of the web page to process.
//	`aResult` The data structure to receive the generated screenshot image.
func configure(aURL string, aResult *[]byte) chromedp.Tasks {
	// Note: `chromedp.FullScreenshot()` overrides the device's
	// emulation settings.
	// Use `device.Reset` to reset the emulation and viewport settings.
	return chromedp.Tasks{
		// ensure basic setup:
		chromedp.Emulate(device.Reset),
		emulation.ClearDeviceMetricsOverride(),
		emulation.ClearGeolocationOverride(),

		// values of '0' will disable the override:
		emulation.SetDeviceMetricsOverride(int64(ssOptions.ImageWidth), int64(ssOptions.ImageHeight), ssOptions.ImageScale, ssOptions.Mobile),

		// set some browser options:
		emulation.SetDocumentCookieDisabled(!ssOptions.Cookies),
		emulation.SetScrollbarsHidden(!ssOptions.Scrollbars),
		security.SetIgnoreCertificateErrors(!ssOptions.CertErrors),
		// security.CertificateErrorAction("continue"),

		emulation.SetScriptExecutionDisabled(!ssOptions.JavaScript),
		// configure the UserAgent to pose as:
		emulation.SetUserAgentOverride(ssOptions.UserAgent).
			// WithAcceptLanguage("en").	//FIXME get proper value format
			WithPlatform(ssOptions.Platform),

		// perform the actual scraping action:
		chromedp.Navigate(aURL),
		chromedp.Sleep(time.Second), // << 1 time to receive&render the page
		chromedp.FullScreenshot(aResult, ssOptions.ImageQuality),
	}
} // configure()

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
			// if xIsBigger {
			// 	// Set the configured size:
			// 	result := image.NewRGBA(image.Rect(0, 0,
			// 		ssOptions.ImageWidth, size.Y))

			// 	// Do the shrinking:
			// 	draw.BiLinear.Scale(result, result.Rect,
			// 		aImgData, bounds, draw.Over, nil)

			// 	return result.SubImage(image.Rect(0, 0, size.X, size.Y))

			// 	// return result
			// } // else the screenshot is too long
			// We just cut off the part outside (below)
			// our wanted/configured height.
			return aImgData.(interface {
				SubImage(aRect image.Rectangle) image.Image
			}).
				SubImage(image.Rect(0, 0, size.X, size.Y))
		} // else xIsBigger

		// Set the configured size:
		result := image.NewRGBA(image.Rect(0, 0, ssOptions.ImageWidth, size.Y))

		// Do the actual shrinking:
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

// `exists()` returns whether there is an usable file cached.
//
// This function uses the `MaxAge()` value to determine whether
// an already existing local file is considered to be too old.
func exists(aFilename string) bool {
	if 0 == ssOptions.ImageAge {
		// shortcut: no checks at all …
		return false
	}

	fi, err := os.Stat(aFilename)
	if (nil != err) || fi.IsDir() {
		return false
	}

	if 10240 > fi.Size() {
		// Empty and small (i.e. `<10KB`) files are ignored.
		// File sizes smaller than ~10KB indicate some kind of
		// error during retrieval of the web page or rendering it.
		// Valid preview images take approximately between 10 to
		// 100 KB depending on the respective web page (e.g. number
		// and size of embedded images).
		return false
	}

	if 0 < ssOptions.ImageAge {
		maxTime := fi.ModTime().Add(ssOptions.ImageAge * time.Second)
		// files too old are ignored
		return time.Now().Before(maxTime)
	}

	return true
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
	// defer cancel()

	defer func() {
		// `chromedp.FullScreenshot()` might panic :-((
		if r := recover(); nil != r {
			if nil == rErr {
				rErr = errors.New(libName +
					": error reading '" + aURL + "'")
			}
			log.Println(libName, rErr)
		}
		cancel()
	}()

	// Capture the entire browser viewport
	if rErr = chromedp.Run(ctx, configure(aURL, &rawData)); nil != rawData {
		if nil != rErr {
			log.Println(libName, ":", aURL, ImageType(), ssOptions.ImageQuality, rErr)
		}
		if rImage = cleanupOutput(rawData); 4096 < len(rImage) {
			rErr = nil
		}
	}

	return
} // generateImage()

// `sanitise()` returns `aURL` with all non alpha/digits removed.
// The resulting string can then be used as the screenshot's file name.
//
//	`aURL` The URL to sanitise.
func sanitise(aURL string) string {
	return ssReplaceNonAlphasRE.ReplaceAllLiteralString(aURL, ``)
} // sanitise()

// 'writeFile()' stores the given image data to a file, returning an error
// in case of problems.
//
//	`aName` The path/file name to use for storing the image.
//	`aData` The image data to store.
//	`aResponse` A (possibly NIL) response data from downloading an image file.
func writeFile(aName string, aData []byte, aResponse *http.Response) (rErr error) {
	if 0 == len(aName) {
		return errors.New(libName + ": empty file name argument")
	}

	var file *os.File

	if file, rErr = os.OpenFile(aName,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640); /* #nosec G302 */ nil != rErr {
		return
	}
	defer file.Close()

	if 0 < len(aData) {
		_, rErr = file.Write(aData)
	} else if (nil != aResponse) && (0 < aResponse.ContentLength) {
		_, rErr = io.Copy(file, aResponse.Body)
	} else {
		_ = os.Remove(aName)
		rErr = errors.New(libName + ": no image data to write '" + aName + "'")
	}

	if nil != rErr {
		// In case of errors during write we delete the file
		// ignoring possible errors here and return the error.
		_ = os.Remove(aName)
	}

	return
} // writeFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// CertErrors returns whether to skip sites with Certificate errors;
// defaults to `false` for historic reasons.
func CertErrors() bool {
	return ssOptions.CertErrors
} // CertErrors()

// SetCertErrors determines whether to skip sites with certificate errors.
//
//	`doAllow` If `false` (i.e. the default) all web-sites will be processed
// regardless of certificate errors.
func SetCertErrors(doAllow bool) {
	ssOptions.CertErrors = doAllow
} // SetCertErrors()

// Cookies returns whether to allow web cookies during page retrieval;
// defaults to `false` for safety and speed reasons.
func Cookies() bool {
	return ssOptions.Cookies
} // Cookies()

// SetCookies determines whether to allow web cookies during page
// retrieval or not.
//
//	`doAllow` If `false` (i.e. the default) no cookies will be available
// during page retrieval, otherwise (i.e. `true`) they will be used.
func SetCookies(doAllow bool) {
	ssOptions.Cookies = doAllow
} // SetCookies()

// CreateImage generates an image of `aURL` and stores it in
// `ImageDirectory()`, returning the file name of the saved image
// or an error in case of problems.
//
//	`aURL` The address of the web page to process.
func CreateImage(aURL string) (string, error) {
	//TODO add 'context' argument

	result := sanitise(aURL) + `.` + ssImageTypes[100 > ssOptions.ImageQuality]
	fName := filepath.Join(ssOptions.ImageDir, result)
	// Check whether we've already got an image file
	// so we might avoid additional network traffic:
	if exists(fName) {
		return result, nil
	}

	var (
		// Declare variables here so we can use them
		// in different contexts/closures below.
		err       error
		imageData []byte
		response  *http.Response
	)

	//TODO turn `Background()` into calltime argument:
	ctx, cancel := context.WithTimeout(context.Background(), ssTimeoutDuration)
	defer func() {
		if r := recover(); nil != r {
			// Timing problems or invalid site data might indirectly
			// cause the image generation to panic.
			log.Println(libName, err)
		}
		cancel()
	}()

	// Exclude certain filetypes from preview generation:
	ext := strings.ToLower(fileExt(aURL))
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
		return "", errors.New(libName +
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
		return "", errors.New(libName +
			": no data received for '" + fName + "'")
	}

	if err = writeFile(fName, imageData, response); nil != err {
		// some problem during attempt to save image to disk
		return "", err
	}

	// Everything went well it seems …
	return result, nil
} // CreateImage()

// ImageAge returns the maximum age of locally stored screenshot images.
func ImageAge() time.Duration {
	return ssOptions.ImageAge
} // ImageAge()

// SetImageAge sets the maximum age of locally stored screenshot images
// before they might get updated by a new call to `CreateImage(…)`.
//
// Usually you'll want this property at its default value (`0`, zero)
// which disables an age check because usually you want an image of the
// page at the time you linked to it.
//
//	`aMaxAge` is the age a page image can have before
// requesting it again.
func SetImageAge(aMaxAge time.Duration) {
	if 0 < aMaxAge {
		ssOptions.ImageAge = aMaxAge
	} else {
		ssOptions.ImageAge = 0
	}
} // SetImageAge()

// ImageDir returns the directory used to store the generated images.
func ImageDir() string {
	return ssOptions.ImageDir
} // ImageDir()

// SetImageDir sets the directory to use for storing the generated
// images, returning an error if `aDirectory` can't be used.
//
//	`aDirectory` The directory to store the generated images.
func SetImageDir(aDirectory string) {
	if aDirectory = strings.TrimSpace(aDirectory); 0 == len(aDirectory) {
		aDirectory = "./"
	}
	dir, err := filepath.Abs(aDirectory)
	if (nil != err) || (0 == len(dir)) {
		dir, _ = filepath.Abs("./")
	}

	ssOptions.ImageDir = dir
} // SetImageDirectory()

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
// Values supported between `1` and `100`; default is `100`.
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
//	`doAllow` If `false` (i.e. the default) no JavaScript will be available
// during page retrieval, otherwise (i.e. `true`) it will be activated.
func SetJavaScript(doAllow bool) {
	ssOptions.JavaScript = doAllow
} // SetJavaScript()

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

// Scrollbars returns whether the virtual browser will show scrollbars.
//
// `@EXPERIMENTAL`
func Scrollbars() bool {
	return ssOptions.Scrollbars
} // Scrollbars()

// SetScrollbars sets whether the virtual browser will show scrollbars.
//
// `@EXPERIMENTAL`
//
//	`aFactor` the new scale factor; `0` disables scaling.
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

// WhiteJS returns the name of the path/file containing hosts/domains
// requiring JavaScript to be active/working.
//
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
func WhiteJS() string {
	return ssOptions.WhiteJS
} // WhiteJS()

// SetWhiteJS configures the name of the file containing hosts/domains
// requiring JavaScript to be active/working.
//
// NOTE: This value is used only if the `JavaScript()` option is set `true`.
//
//	`aName` The path/filename of sites with required JavaScript.
func SetWhiteJS(aName string) {
	aName = strings.TrimSpace(aName)
	if 0 == len(aName) {
		aName = "./" + defJsWhiteName
	}

	if pathname, err := filepath.Abs(aName); nil == err {
		ssOptions.WhiteJS = pathname
	}
	// else: leave the setting unchanged
} // SetWhiteJS()

/* _EoF_ */
