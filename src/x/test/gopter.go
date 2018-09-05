// Copyright (c) 2018 Uber Technologies, Inc.
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

package test

// FOLLOWUP(prateek): This file needs to be removed once https://github.com/leanovate/gopter/pull/41 lands

import (
	"reflect"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// SliceOf generates an arbitrary slice of generated elements
// genParams.MaxSize sets an (exclusive) upper limit on the size of the slice
// genParams.MinSize sets an (inclusive) lower limit on the size of the slice
func SliceOf(elementGen gopter.Gen, typeOverrides ...reflect.Type) gopter.Gen {
	var typeOverride reflect.Type
	if len(typeOverrides) > 1 {
		panic("too many type overrides specified, at most 1 may be provided.")
	} else if len(typeOverrides) == 1 {
		typeOverride = typeOverrides[0]
	}
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		len := 0
		if genParams.MaxSize > 0 || genParams.MinSize > 0 {
			if genParams.MinSize > genParams.MaxSize {
				panic("GenParameters.MinSize must be <= GenParameters.MaxSize")
			}

			if genParams.MaxSize == genParams.MinSize {
				len = genParams.MaxSize
			} else {
				len = genParams.Rng.Intn(genParams.MaxSize-genParams.MinSize) + genParams.MinSize
			}
		}
		result, elementSieve, elementShrinker := genSlice(elementGen, genParams, len, typeOverride)

		genResult := gopter.NewGenResult(result.Interface(), gen.SliceShrinker(elementShrinker))
		if elementSieve != nil {
			genResult.Sieve = forAllSieve(elementSieve)
		}
		return genResult
	}
}

// SliceOfN generates a slice of generated elements with definied length
func SliceOfN(l int, elementGen gopter.Gen, typeOverrides ...reflect.Type) gopter.Gen {
	var typeOverride reflect.Type
	if len(typeOverrides) > 1 {
		panic("too many type overrides specified, at most 1 may be provided.")
	} else if len(typeOverrides) == 1 {
		typeOverride = typeOverrides[0]
	}
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		result, elementSieve, elementShrinker := genSlice(elementGen, genParams, l, typeOverride)

		genResult := gopter.NewGenResult(result.Interface(), gen.SliceShrinkerOne(elementShrinker))
		if elementSieve != nil {
			genResult.Sieve = func(v interface{}) bool {
				rv := reflect.ValueOf(v)
				return rv.Len() == l && forAllSieve(elementSieve)(v)
			}
		} else {
			genResult.Sieve = func(v interface{}) bool {
				return reflect.ValueOf(v).Len() == l
			}
		}
		return genResult
	}
}

func genSlice(elementGen gopter.Gen, genParams *gopter.GenParameters, len int, typeOverride reflect.Type) (reflect.Value, func(interface{}) bool, gopter.Shrinker) {
	element := elementGen(genParams)
	elementSieve := element.Sieve
	elementShrinker := element.Shrinker

	sliceType := typeOverride
	if sliceType == nil {
		sliceType = element.ResultType
	}

	result := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, len)

	for i := 0; i < len; i++ {
		value, ok := element.Retrieve()

		if ok {
			if value == nil {
				result = reflect.Append(result, reflect.Zero(sliceType))
			} else {
				result = reflect.Append(result, reflect.ValueOf(value))
			}
		}
		element = elementGen(genParams)
	}

	return result, elementSieve, elementShrinker
}

func forAllSieve(elementSieve func(interface{}) bool) func(interface{}) bool {
	return func(v interface{}) bool {
		rv := reflect.ValueOf(v)
		for i := rv.Len() - 1; i >= 0; i-- {
			if !elementSieve(rv.Index(i).Interface()) {
				return false
			}
		}
		return true
	}
}
