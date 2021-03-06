// torequtil has utility functions used by toreq and toreqnew
// which don't require the Traffic Ops client, and thus can be shared.
package torequtil

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"errors"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/apache/trafficcontrol/lib/go-log"
)

// GetRetry attempts to get the given object, retrying with exponential backoff up to cfg.NumRetries.
// The objName is not used in actual fetching or logic, but only for logging. It can be any printable string, but should be unique and reflect the object being fetched.
func GetRetry(numRetries int, objName string, obj interface{}, getter func(obj interface{}) error) error {
	start := time.Now()
	currentRetry := 0
	for {
		err := getter(obj)
		if err == nil {
			break
		}
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			// if the server returned a 404, retrying won't help
			return errors.New("getting uncached: " + err.Error())
		}
		if currentRetry == numRetries {
			return errors.New("getting uncached: " + err.Error())
		}

		sleepSeconds := RetryBackoffSeconds(currentRetry)
		log.Warnf("getting '%v', sleeping for %v seconds: %v\n", objName, sleepSeconds, err)
		currentRetry++
		time.Sleep(time.Second * time.Duration(sleepSeconds)) // TODO make backoff configurable?
	}

	log.Infof("GetRetry %v retries %v took %v\n", objName, currentRetry, time.Since(start).Round(time.Millisecond))
	return nil
}

func RetryBackoffSeconds(currentRetry int) int {
	// TODO make configurable?
	return int(math.Pow(2.0, float64(currentRetry)))
}

// MaybeIPStr returns the addr string if it isn't nil, or the empty string if it is.
// This is intended for logging, to allow logging with one line, whether addr is nil or not.
func MaybeIPStr(addr net.Addr) string {
	if addr != nil {
		return addr.String()
	}
	return ""
}

func StringToCookies(cookiesStr string) []*http.Cookie {
	hdr := http.Header{}
	hdr.Add("Cookie", cookiesStr)
	req := http.Request{Header: hdr}
	return req.Cookies()
}

func CookiesToString(cookies []*http.Cookie) string {
	strs := []string{}
	for _, cookie := range cookies {
		strs = append(strs, cookie.String())
	}
	return strings.Join(strs, "; ")
}
