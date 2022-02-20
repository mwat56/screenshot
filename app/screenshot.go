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
	"time"

	"github.com/mwat56/screenshot"
)

// `processOptions()` handles the commandline arguments.
//
// While all properties of the `screenshot` library are exposed
// only the `-a {URL}` argument is required all other options use
// reasonable default values.
func processOptions() (rURL string, rVerbose bool) {
	// get the library's default values:
	opts := screenshot.Options()

	// --- setup handling of the program's commandline options

	flag.StringVar(&rURL, "a", rURL,
		"(*required*) the address/URL for the browser's screenshot")

	// --- browser related settings:

	s := fmt.Sprintf("allow the browser to handle web cookies (default %v)", opts.Cookies)
	flag.BoolVar(&opts.Cookies, "bc", opts.Cookies, s)

	s = fmt.Sprintf("skip sites with Certificate errors (default %v)", opts.CertErrors)
	flag.BoolVar(&opts.CertErrors, "be", opts.CertErrors, s)

	s = fmt.Sprintf("let browser emulate a mobile device (default %v)", opts.Mobile)
	flag.BoolVar(&opts.Mobile, "bm", opts.Mobile, s)

	s = fmt.Sprintf("let browser show scrollbars if available (default %v)", opts.Scrollbars)
	flag.BoolVar(&opts.Scrollbars, "bs", opts.Scrollbars, s)

	maxProcessTime := int64(opts.MaxProcessTime / time.Second)
	flag.Int64Var(&maxProcessTime, "bt", maxProcessTime,
		"max. time (seconds) allowed to process a single web page")

	// --- image related settings:

	flag.StringVar(&opts.ImageDir, "id", opts.ImageDir,
		"directory for storing the screenshot image")

	flag.IntVar(&opts.ImageHeight, "ih", opts.ImageHeight,
		"max. height of the screenshot image")

	flag.IntVar(&opts.ImageQuality, "iq", opts.ImageQuality,
		"quality of the screenshot image")

	flag.Float64Var(&opts.ImageScale, "is", opts.ImageScale,
		"the browser's scale factor for the screenshot image")

	flag.IntVar(&opts.ImageWidth, "iw", opts.ImageWidth,
		"max. width of the screenshot image")

	// --- JavaScript related settings:

	flag.StringVar(&opts.HostsAvoidJS, "ja", opts.HostsAvoidJS,
		"name of sites that should avoid JavaScript running")

	flag.StringVar(&opts.HostsNeedJS, "jn", opts.HostsNeedJS,
		"name of sites that need JavaScript available")

	flag.StringVar(&opts.Platform, "jp", opts.Platform,
		"Identifier the JavaScript `navigator.platform` should use")

	s = fmt.Sprintf("allow browser's use of JavaScript (default %v)", opts.JavaScript)
	flag.BoolVar(&opts.JavaScript, "js", opts.JavaScript, s)

	flag.StringVar(&opts.UserAgent, "ju", opts.UserAgent,
		"description of the browser's UserAgent to use")

	// --- general settings:

	s = fmt.Sprintf("verbose (default %v)", rVerbose)
	flag.BoolVar(&rVerbose, "v", rVerbose, s)

	// --- process the commandline:

	flag.Usage = showHelp
	flag.Parse()

	// --- setup the program:

	opts.MaxProcessTime = time.Duration(maxProcessTime)
	screenshot.Setup(opts)

	return
} // processOptions()

// `showHelp()` lists the commandline options to `Stderr`.
func showHelp() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
} // showHelp()

// `exit()` does _not_ return but terminates the program.
//
//	`aText` Info message to present to the user.
//	`aHelp` Flag whether to show the commandline options.
//	`isVerbose` Flag whether to show the configured screenshot options.
//	`aCode` The program's exit code.
func exit(aText string, aHelp, isVerbose bool, aCode int) {
	fmt.Print("\n", aText, "\n")
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

	if url, verbose = processOptions(); 0 == len(url) {
		exit("missing URL - terminating ...", true, verbose, 1)
	}

	if fName, err = screenshot.CreateImage(url); nil != err {
		exit(fmt.Sprintln("error:", err), true, verbose, 1)
	}

	exit(fmt.Sprintln("generated URL screenshot:", fName, "\n\t"),
		false, verbose, 0)
} // main()

/* _EoF_ */
