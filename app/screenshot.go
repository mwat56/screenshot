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
		boolean, bC, bE, bM, jS bool
		iH, iQ, iW              int
		iS                      float64
		uA, iD, jpf, s, jw      string
	)

	flag.StringVar(&rURL, "a", rURL,
		"(*required*) the address/URL to lookup for a screenshot")

	boolean = screenshot.Cookies()
	s = fmt.Sprintf("allow the browser to handle web cookies (default %v)", boolean)
	flag.BoolVar(&bC, "bc", boolean, s)

	boolean = screenshot.CertErrors()
	s = fmt.Sprintf("skip sites with Certificate errors (default %v)", boolean)
	flag.BoolVar(&bE, "be", boolean, s)

	boolean = screenshot.Mobile()
	s = fmt.Sprintf("let browser emulate a mobile device (default %v)", boolean)
	flag.BoolVar(&bM, "bm", boolean, s)

	flag.Float64Var(&iS, "bs", screenshot.ImageScale(),
		"the browser's scale factor for the screenshot")

	flag.StringVar(&iD, "d", "./",
		"directory for storing the screenshot")

	flag.IntVar(&iH, "ih", screenshot.ImageHeight(),
		"max. height of the screenshot")

	flag.IntVar(&iQ, "iq", screenshot.ImageQuality(),
		"quality of the screenshot")

	flag.IntVar(&iW, "iw", screenshot.ImageWidth(),
		"max. width of the screenshot")

	boolean = screenshot.JavaScript()
	s = fmt.Sprintf("allow use of JavaScript (default %v)", boolean)
	flag.BoolVar(&jS, "j", boolean, s)

	flag.StringVar(&jpf, "jp", screenshot.Platform(),
		"Identifier the JavaScript `navigator.platform` should use")

	flag.StringVar(&uA, "ju", screenshot.UserAgent(),
		"description of the UserAgent to use")

	flag.StringVar(&jw, "jw", screenshot.WhiteJS(),
		"name of the JavaScript whitelist")

	s = fmt.Sprintf("verbose (default %v)", rVerbose)
	flag.BoolVar(&rVerbose, "v", rVerbose, s)

	flag.Usage = showHelp
	flag.Parse()

	screenshot.Setup(&screenshot.ScreenshotParams{
		Cookies:      bC,
		CertErrors:   bE,
		ImageAge:     screenshot.ImageAge(), // use default value
		ImageDir:     iD,
		ImageHeight:  iH,
		ImageQuality: iQ,
		ImageScale:   iS,
		ImageWidth:   iW,
		JavaScript:   jS,
		Mobile:       bM,
		Platform:     jpf,
		Scrollbars:   screenshot.Scrollbars(), // use default value
		UserAgent:    uA,
		WhiteJS:      jw,
	})

	return
} // parseCommandLine()

// showHelp lists the commandline options to `Stderr`.
func showHelp() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
} // showHelp()

// Main function running this program.
func main() {
	var (
		fName, url string
		err        error
		verbose    bool
	)

	if url, verbose = parseCommandLine(); 0 == len(url) {
		fmt.Println("\nmissing URL - terminating ...")
		showHelp()
		os.Exit(1)
	}

	if fName, err = screenshot.CreateImage(url); nil != err {
		fmt.Fprint(os.Stderr, "\n", "error:", err)
		showHelp()
		os.Exit(1)
	}

	if verbose {
		fmt.Println(screenshot.String())
	}

	fmt.Println("generated URL screenshot:", fName, "\n\t")
	os.Exit(0)
} // main()

/* _EoF_ */
