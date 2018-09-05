// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package index

import (
	"bytes"
	re "regexp"

	vregex "github.com/couchbase/vellum/regexp"
)

// CompileRegex compiles the provided regexp.
func CompileRegex(r []byte) (CompiledRegex, error) {
	var simpleRESBytes, fstREBytes []byte

	// NB(prateek): the fst segment (Vellum) treats all Regexps as anchored, so
	// we make sure the mem segment does the same to ensure the behaviour is consistent.
	if bytes.HasPrefix(r, []byte("^")) {
		simpleRESBytes = r
		fstREBytes = fstREBytes[1:]
	} else {
		simpleRESBytes = append([]byte("^"), r...)
		fstREBytes = r
	}

	var (
		simpleREString = string(simpleRESBytes)
		fstREString    = string(fstREBytes)
		compiledRegex  = CompiledRegex{}
	)
	simpleRE, err := re.Compile(simpleREString)
	if err != nil {
		return compiledRegex, err
	}
	compiledRegex.Simple = simpleRE

	fstRE, err := vregex.New(fstREString)
	if err != nil {
		return compiledRegex, err
	}
	compiledRegex.FST = fstRE

	return compiledRegex, nil
}
