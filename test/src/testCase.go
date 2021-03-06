// Copyright 2016 Etix Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Test represents a test launched with Cameradar
type Test struct {
	expected []Result      // Contains the expected results
	result   []Result      // Contains the results that have been validated
	time     time.Duration // Contains the runtime duration
}

func removeResult(expected []Result, index int) []Result {
	if len(expected) > 1 {
		return append(expected[:index], expected[index+1:]...)
	}
	return []Result{}
}

// Invoke the test
// Wrap results in a TestResult object
func (t *Tester) invokeTestCase(testCase *Test, wg *sync.WaitGroup) {
	startTime := time.Now()
	t.runTestCase(testCase)
	testCase.time = time.Since(startTime)
	fmt.Printf("Test OK in %.6fs\n", testCase.time.Seconds())
	wg.Done()
}

// Checks all valid results that are supposed to match
// Adds them to the valid results and leave the failed
// ones in the expected slice
//
// Then, if the result did not match the expected but it was supposed to fail
// Add it to the valid results and remove it from the expected slice
func (t *Tester) runTestCase(test *Test) {
	startService(&t.Cameradar, t.ServiceConf)
	for t.Cameradar.Active {
		time.Sleep(25 * time.Millisecond)
	}

	var validResults []Result
	var invalidResults []Result
	if getResult(&test.result, "/tmp/shared/result.json") {
		for _, r := range test.result {
			r.Valid = true

			for index, e := range test.expected {
				fmt.Println("Result : ", r)
				fmt.Println("Expected test : ", e)

				if e.Address == r.Address && isValid(&e, r) {
					fmt.Println("The result of ", r.Address, " is valid.")
					validResults = Extend(validResults, r)
					test.expected = removeResult(test.expected, index)
					break
				}
			}
		}

		for _, e := range test.expected {
			if !e.Valid {
				fmt.Println("The result of", e.Address, "successfully failed.")
				validResults = Extend(validResults, e)
			} else {
				if e.err == nil {
					e.err = errors.New("The camera with the address " + e.Address + " was not found by cameradar")
				}
				invalidResults = Extend(invalidResults, e)
				fmt.Println("Should have been valid but was not found : ", e.Address)
			}
		}
		test.result = validResults
		test.expected = invalidResults
	} else {
		test.expected = nil
		test.result = nil
	}
}
