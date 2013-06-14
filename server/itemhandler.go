// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"github.com/andreaskoch/allmark/path"
	"net/http"
	"os"
	"strings"
)

var itemHandler = func(w http.ResponseWriter, r *http.Request) {
	requestedPath := r.URL.Path
	requestedPath = strings.TrimLeft(requestedPath, "/")

	fmt.Printf("Requested route %q\n", requestedPath)

	// make sure the request body is closed
	defer r.Body.Close()

	filePath, ok := routes[requestedPath]
	if !ok {

		// check for fallbacks before returning a 404
		if fallbackRoute, fallbackRouteFound := getFallbackRoute(requestedPath); fallbackRouteFound {
			redirect(w, r, fallbackRoute)
			return
		}

		error404Handler(w, r)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		error404Handler(w, r)
		return
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		error404Handler(w, r)
		return
	}

	http.ServeContent(w, r, filePath, fileInfo.ModTime(), file)
}

func getFallbackRoute(requestedPath string) (fallbackRoute string, found bool) {

	requestedPath = path.StripLeadingUrlDirectorySeperator(requestedPath)

	if len(requestedPath) == 0 {
		fmt.Printf("Fallback for %q is %q\n", requestedPath, path.WebServerDefaultFilename)
		return path.WebServerDefaultFilename, true
	}

	if strings.HasSuffix(requestedPath, path.WebServerDefaultFilename) {
		fmt.Printf("No fallback found for %q\n", requestedPath)
		return "", false
	}

	route := path.CombineUrlComponents(requestedPath, path.WebServerDefaultFilename)
	if _, ok := routes[route]; ok {
		fmt.Printf("Fallback for %q is %q\n", requestedPath, route)
		return route, true
	}

	fmt.Printf("No fallback found for %q\n", requestedPath)
	return "", false
}
