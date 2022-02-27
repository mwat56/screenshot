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

// `processOptions()` handles the commandline arguments.
//
// While all properties of the `screenshot` library are exposed
// only the `-a {URL}` argument is required all other options use
// reasonable default values.
func processOptions() (rURL string, rVerbose bool) {
	var s string

	opts := screenshot.Options() // get the library's default values

	// --- setup handling of the program's commandline options

	// --- browser related settings:

	s = `allow the browser to handle web cookies`
	if !opts.Cookies {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.Cookies, `bc`, opts.Cookies, s)

	s = `skip sites with Certificate errors`
	if !opts.CertErrors {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.CertErrors, `be`, opts.CertErrors, s)

	s = `let browser emulate a mobile device`
	if !opts.Mobile {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.Mobile, `bm`, opts.Mobile, s)

	s = `let browser show scrollbars if available`
	if !opts.Scrollbars {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.Scrollbars, `bs`, opts.Scrollbars, s)

	flag.Int64Var(&opts.MaxProcessTime, `bt`, opts.MaxProcessTime,
		"max. time (seconds) allowed to process a single web page")

	// --- image related settings:

	s = `accept the respective other image format`
	if !opts.AcceptOther {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.AcceptOther, `ia`, opts.AcceptOther, s)

	flag.StringVar(&opts.ImageDir, `id`, opts.ImageDir,
		"directory for storing the screenshot image")

	flag.IntVar(&opts.ImageHeight, `ih`, opts.ImageHeight,
		"max. height of the screenshot image")

	s = `overwrite an existing image`
	if !opts.ImageOverwrite {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.ImageOverwrite, `io`, opts.ImageOverwrite, s)

	flag.IntVar(&opts.ImageQuality, `iq`, opts.ImageQuality,
		"quality of the screenshot image")

	s = "the browser's scale factor for the screenshot image"
	if 0 >= opts.ImageScale {
		s += ` (default 0.00)`
	}
	flag.Float64Var(&opts.ImageScale, `is`, opts.ImageScale, s)

	flag.IntVar(&opts.ImageWidth, `iw`, opts.ImageWidth,
		"max. width of the screenshot image")

	// --- JavaScript related settings:

	flag.StringVar(&opts.HostsAvoidJSfile, `ja`, opts.HostsAvoidJSfile,
		"name of text-file that contains sites better avoiding JavaScript\n")

	flag.StringVar(&opts.HostsNeedJSfile, `jn`, opts.HostsNeedJSfile,
		"name of text-file that contains sites needing JavaScript\n")

	flag.StringVar(&opts.Platform, `jp`, opts.Platform,
		"Identifier the JavaScript `navigator.platform` should use")

	s = `allow browser's use of JavaScript`
	if !opts.JavaScript {
		s += ` (default false)`
	}
	flag.BoolVar(&opts.JavaScript, `js`, opts.JavaScript, s)

	flag.StringVar(&opts.UserAgent, `ju`, opts.UserAgent,
		"description of the UserAgent the browser should report\n")

	// --- general options:

	flag.StringVar(&rURL, `u`, rURL,
		"(*required*) the URL for the browser's screenshot")

	s = fmt.Sprintf(`verbose (default %v)`, rVerbose)
	flag.BoolVar(&rVerbose, `v`, rVerbose, s)

	// --- process the commandline:

	flag.Usage = showHelp
	flag.Parse()

	// --- setup the `screenshot` library:

	screenshot.Setup(opts)

	return
} // processOptions()

// `showHelp()` lists the commandline options to `Stderr`.
func showHelp() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Println()
} // showHelp()

// `exit()` does _not_ return but terminates the program.
//
//	`aText` Info message to present to the user.
//	`aHelp` Flag whether to show the commandline options.
//	`isVerbose` Flag whether to show the configured screenshot options.
//	`aCode` The program's exit code.
func exit(aText string, aHelp, isVerbose bool, aCode int) {
	fmt.Print("\n", aText, "\n\n")
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
