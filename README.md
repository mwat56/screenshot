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
	- [Licence](#licence)

----

## Purpose

Sometimes yoo don't want just standard web links in your web presentation but a preview image showing the page you're linking to.
That is where this package comes in.
It generates – by way of calling the external `Chrome` browser – an image of the web page a given URL addresses.
Those image files are stored locally and may be used as often as you want without additional external network traffic.

## Installation

You can use `Go` to install this package for you:

	go get github.com/mwat56/screenshot

After that you can `import` it the usual Go way to use the library.

## Usage

There are only two functions you have to worry about:

	// SetImageDirectory sets the directory to use for storing the
	// generated images, returning an error if `aDirectory` can't be used.
	//
	//	`aDirectory` The directory to store the generated images.
	func SetImageDirectory(aDirectory string) error { … }

This function should be called before any other one to make sure the generated images end up where you want them to be.
The default is the current directory from where the program was run.

To actually create a screenshot image you'd call:

	// CreateImage generates an image of `aURL` and stores it in
	// `ImageDirectory()`, returning the file name of the saved image
	// or an error in case of problems.
	//
	//	`aURL` The address of the web page to process.
	func CreateImage(aURL string) (string, error) { … }

The returned string is the name of the generated image file (without its path).
If you combine it with the directory returned by `ImageDirectory()` you get the complete path/filename to locally access the image.

Generating a preview image usually takes between one and five seconds, depending on the actual web-page in question, however, it can take considerably longer.
To avoid hanging the program the `CreateImage()` function uses an internal timeout of half minute.

And, finally, not all web-pages can be rendered properly and turned into an image.
In case of errors (like network-errors or problem while storing the image file) `CreateImage()` returns an empty filename and an error.

There are a few more functions which you will probably barely need; for details refer to the [source code documentation](https://godoc.org/github.com/mwat56/screenshot).

## Libraries

The Go library controlling a headless instance of the `Chrome` browser

* [ChromeDP](https://github.com/chromedp/chromedp)

is  **_required_**  for this package to work.
Under Linux this browser is usually part of your distribution (as `chromium-browser`).

To resize the screenshot if required by the `ImageHeight()`/`ImageWidth()` values the

* [Go `draw` library](http://golang.org/x/image/draw/)

must be part of your Go installation (if not run: `go get -u golang.org/x/image/draw`).

## Example

In the source code directory's sub-directory `app/` there's a demo program (`screenshot.go`) allowing you to generate a screenshot image of an URL given on the commandline.
To run it call

	cd app
	go build screenshot.go
	./screenshot

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
