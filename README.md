# ScreenShot

[![golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/mwat56/screenshot?status.svg)](https://godoc.org/github.com/mwat56/screenshot)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/screenshot)](https://goreportcard.com/report/github.com/mwat56/screenshot)
[![Issues](https://img.shields.io/github/issues/mwat56/screenshot.svg)](https://github.com/mwat56/screenshot/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/screenshot.svg)](https://github.com/mwat56/screenshot/)
[![Tag](https://img.shields.io/github/tag/mwat56/screenshot.svg)](https://github.com/mwat56/screenshot/tags)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/screenshot/blob/main/app/screenshot.go)
[![License](https://img.shields.io/github/mwat56/screenshot.svg)](https://github.com/mwat56/screenshot/blob/main/LICENSE)

- [ScreenShot](#screenshot)
	- [Purpose](#purpose)
	- [Installation](#installation)
	- [Usage](#usage)
	- [Libraries](#libraries)
	- [Example](#example)
	- [History](#history)
	- [Licence](#licence)

----

## Purpose

Sometimes you don't want just standard web links in your web presentation but a preview image showing the page you're linking to. That is where this package comes in. It generates – by way of calling the external `Chrome` browser – an image of the web page a given URL addresses. Those image files are stored locally and may be used as often as you want without additional external network traffic.

## Installation

You can use `Go` to install this package for you:

	go get github.com/mwat56/screenshot

After that you can `import` it the usual Go way to use the library.

## Usage

There are only two functions you have to worry about:

	// SetImageDir sets the directory to use for storing the generated
	// screenshot images.
	//
	// If `aDirectory` is empty or invalid the system's temp directory is used.
	//
	// `aDirectory` The directory to store the generated images.
	func SetImageDir(aDirectory string) { … }

This function should be called before any other one to make sure the generated screenshots end up where you want them to be. The default is the system's `temp` directory (e.g. `/tmp` under GNU/Linux).

To actually create the screenshot image you'd call:

	// CreateImage generates an image of `aURL` and stores it in `ImageDir()`,
	// returning the file name of the saved image or an error in case of problems.
	//
	//	`aURL` The address of the web page to process.
	func CreateImage(aURL string) (string, error) { … }

The returned string is the name of the generated image file (without its path). If you combine it with the directory returned by `ImageDir()` you get the complete path/filename to locally access the image.

Generating a screenshot image usually takes between one and five seconds, depending on the actual web-page in question; however, it can take considerably longer. To avoid hanging the program the `CreateImage()` function uses a timeout of half a minute.

And, finally, not all web-pages can be rendered properly and turned into an image. In case of errors (like network-errors or problem while storing the image file) `CreateImage()` returns an empty filename and an error.

There are a couple more functions (mostly property GETters and SETters) which you will probably barely need; for details refer to the [source code documentation](https://godoc.org/github.com/mwat56/screenshot).

## Libraries

The Go library controlling a headless instance of the `Chrome` browser

* [ChromeDP](https://github.com/chromedp/chromedp)

is  _required_  for this package to work.
Under Linux this browser is usually part of your distribution (as `chromium-browser`).

To resize the screenshot if required by the `ImageHeight()`/`ImageWidth()` values the

* [Go 'draw' library](https://golang.org/x/image/draw/)

must be part of your Go installation (if not, run: `go get -u golang.org/x/image/draw`).

## Example

In the source code's sub-directory `app/` there's a demo program (`screenshot.go`) allowing you to generate a screenshot image of an URL given on the commandline.

To run it call e.g.

	#> cd app
	#> go build screenshot.go
	#> ./screenshot

It will show you all available commandline options e.g.:

	Usage: ./screenshot [OPTIONS]

	-bc
		allow the browser to handle web cookies (default false)
	-be
		skip sites with Certificate errors (default false)
	-bm
		let browser emulate a mobile device (default false)
	-bs
		let browser show scrollbars if available (default false)
	-bt int
		max. time (seconds) allowed to process a single web page (default 32)
	-ia
		accept the respective other image format (default true)
	-id string
		directory for storing the screenshot image (default "/tmp")
	-ih int
		max. height of the screenshot image (default 768)
	-io
		overwrite an existing image (default false)
	-iq int
		quality of the screenshot image (default 75)
	-is float
		the browser's scale factor for the screenshot image (default 0.00)
	-iw int
		max. width of the screenshot image (default 896)
	-ja string
		name of text-file that contains sites better avoiding JavaScript
		(default "/home/matthias/devel/Go/src/github.com/mwat56/screenshot/app/hostsavoidjs.list")
	-jn string
		name of text-file that contains sites needing JavaScript
		(default "/home/matthias/devel/Go/src/github.com/mwat56/screenshot/app/hostsneedjs.list")
	-jp navigator.platform
		Identifier the JavaScript navigator.platform should use (default "Linux x86_64")
	-js
		allow browser's use of JavaScript (default false)
	-ju string
		description of the UserAgent the browser should report
		(default "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	-u string
		(*required*) the URL for the browser's screenshot
	-v	verbose (default false)

As noted before you'll only need the `-u string` option, obviously.

You can use this program to generate screenshot images "by hand" and fiddle with the various commandline options to see what difference it makes if you change them.

## History

Prior to this a few years back I wrote the [pageview](https://github.com/mwat56/pageview/) package which used the external [wkhtmltoimage](https://wkhtmltopdf.org/downloads.html) program; and in most cases it worked just fine. However, once in a while `wkhtmltoimage` produced a `segmentation fault (core dumped)` – reproducible. For a while I thought I could live with it, but over time it happened more often (i.e. with additional URLs). Fiddling around with various commandline options provided no improvement.
In the end I started to look around, searching for alternative approaches – short of writing my own URL retrieval and rendering system. That's when I found [ChromeDP](https://github.com/chromedp/chromedp) and hence this [package](https://godoc.org/github.com/mwat56/screenshot) came into existence.

## Licence

        Copyright © 2022 M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.

----
[![GFDL](https://www.gnu.org/graphics/gfdl-logo-tiny.png)](http://www.gnu.org/copyleft/fdl.html)
