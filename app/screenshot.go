/*
   Copyright © 2022 M.Watermann, 10247 Berlin, Germany
               All rights reserved
           EMail : <support@mwat.de>
*/

package main

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"flag"
	"fmt"
	"os"

	"github.com/mwat56/screenshot"
)

// `parseCommandLine()` handles the commandline arguments.
// While all properties of the `screenshot` library are exposed
// only the `-a {URL}` argument is required everything else uses
// reasonable default values.
func parseCommandLine() (rURL string, rVerbose bool) {
	var (
		boolean, bC, bE, bM, bS, jS bool
		iS                          float64
		iH, iQ, iW                  int
		uA, iD, jpf, s, jw          string
	)

	flag.StringVar(&rURL, "ba", rURL,
		"(*required*) the address/URL for the browser's screenshot")

	boolean = screenshot.Cookies()
	s = fmt.Sprintf("allow the browser to handle web cookies (default %v)", boolean)
	flag.BoolVar(&bC, "bc", boolean, s)

	boolean = screenshot.CertErrors()
	s = fmt.Sprintf("skip sites with Certificate errors (default %v)", boolean)
	flag.BoolVar(&bE, "be", boolean, s)

	boolean = screenshot.Mobile()
	s = fmt.Sprintf("let browser emulate a mobile device (default %v)", boolean)
	flag.BoolVar(&bM, "bm", boolean, s)

	boolean = screenshot.Scrollbars() //FIXME EXPERIMENTAL
	s = fmt.Sprintf("let browser show scrollbars if available (default %v)", boolean)
	flag.BoolVar(&bS, "bs", boolean, s)

	flag.StringVar(&iD, "id", screenshot.ImageDir(),
		"directory for storing the screenshot image")

	flag.IntVar(&iH, "ih", screenshot.ImageHeight(),
		"max. height of the screenshot image")

	flag.IntVar(&iQ, "iq", screenshot.ImageQuality(),
		"quality of the screenshot image")

	flag.Float64Var(&iS, "is", screenshot.ImageScale(),
		"the browser's scale factor for the screenshot image")

	flag.IntVar(&iW, "iw", screenshot.ImageWidth(),
		"max. width of the screenshot image")

	boolean = screenshot.JavaScript()
	s = fmt.Sprintf("allow browser's use of JavaScript (default %v)", boolean)
	flag.BoolVar(&jS, "js", boolean, s)

	flag.StringVar(&jpf, "jp", screenshot.Platform(),
		"Identifier the JavaScript `navigator.platform` should use")

	flag.StringVar(&uA, "ju", screenshot.UserAgent(),
		"description of the browser's UserAgent to use")

	flag.StringVar(&jw, "jw", screenshot.WhiteJS(),
		"name of the JavaScript whitelist")

	s = fmt.Sprintf("verbose (default %v)", rVerbose)
	flag.BoolVar(&rVerbose, "v", rVerbose, s)

	flag.Usage = showHelp
	flag.Parse()

	screenshot.Setup(&screenshot.ScreenshotParams{
		Cookies:      bC,
		CertErrors:   bE,
		ImageAge:     0, // i.e. default value
		ImageDir:     iD,
		ImageHeight:  iH,
		ImageQuality: iQ,
		ImageScale:   iS,
		ImageWidth:   iW,
		JavaScript:   jS,
		Mobile:       bM,
		Platform:     jpf,
		Scrollbars:   bS, //FIXME EXPERIMENTAL
		UserAgent:    uA,
		WhiteJS:      jw,
	})

	return
} // parseCommandLine()

// showHelp lists the commandline options to `Stderr`.
func showHelp() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	// fmt.Fprint(os.Stderr, "\n")
} // showHelp()

// `exit()` does _not_ return but terminates the program.
//
//	`aText`
//	`aHelp` Flag whether to show the commandline options.
//	`isVerbose` Flag whether to show the configured screenshot options.
//	`aCode` The program's exit code.
func exit(aText string, aHelp, isVerbose bool, aCode int) {
	fmt.Println(aText)
	if isVerbose {
		fmt.Println(screenshot.String())
		showHelp()
	} else if aHelp {
		showHelp()
	}
	os.Exit(aCode)
} // exit()

// Main function running this program.
func main() {
	var (
		fName, url string
		err        error
		verbose    bool
	)

	if url, verbose = parseCommandLine(); 0 == len(url) {
		exit("\nmissing URL - terminating ...", true, verbose, 1)
	}

	if fName, err = screenshot.CreateImage(url); nil != err {
		exit(fmt.Sprintln("\nerror:", err), true, verbose, 1)
	}

	exit(fmt.Sprintln("generated URL screenshot:", fName, "\n\t"),
		false, verbose, 0)
} // main()

/* _EoF_ */
