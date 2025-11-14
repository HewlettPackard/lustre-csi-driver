/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRoundUpBytes(t *testing.T) {
	var sizeInBytes int64 = 1024
	actual := RoundUpBytes(sizeInBytes)
	if actual != 1*GiB {
		t.Fatalf("Wrong result for RoundUpBytes. Got: %d", actual)
	}
}

func TestRoundUpGiB(t *testing.T) {
	var sizeInBytes int64 = 1
	actual := RoundUpGiB(sizeInBytes)
	if actual != 1 {
		t.Fatalf("Wrong result for RoundUpGiB. Got: %d", actual)
	}
}

func TestSimpleLockEntry(t *testing.T) {
	testLockMap := NewLockMap()

	callbackChan1 := make(chan any)
	go testLockMap.lockAndCallback(t, "entry1", callbackChan1)
	ensureCallbackHappens(t, callbackChan1)
}

func TestSimpleLockUnlockEntry(t *testing.T) {
	testLockMap := NewLockMap()

	callbackChan1 := make(chan any)
	go testLockMap.lockAndCallback(t, "entry1", callbackChan1)
	ensureCallbackHappens(t, callbackChan1)
	testLockMap.UnlockEntry("entry1")
}

func TestConcurrentLockEntry(t *testing.T) {
	testLockMap := NewLockMap()

	callbackChan1 := make(chan any)
	callbackChan2 := make(chan any)

	go testLockMap.lockAndCallback(t, "entry1", callbackChan1)
	ensureCallbackHappens(t, callbackChan1)

	go testLockMap.lockAndCallback(t, "entry1", callbackChan2)
	ensureNoCallback(t, callbackChan2)

	testLockMap.UnlockEntry("entry1")
	ensureCallbackHappens(t, callbackChan2)
	testLockMap.UnlockEntry("entry1")
}

func (lm *LockMap) lockAndCallback(_ *testing.T, entry string, callbackChan chan<- any) {
	lm.LockEntry(entry)
	callbackChan <- true
}

var callbackTimeout = 2 * time.Second

func ensureCallbackHappens(t *testing.T, callbackChan <-chan any) bool {
	t.Helper()
	select {
	case <-callbackChan:
		return true
	case <-time.After(callbackTimeout):
		t.Fatalf("timed out waiting for callback")
		return false
	}
}

func ensureNoCallback(t *testing.T, callbackChan <-chan any) bool {
	t.Helper()
	select {
	case <-callbackChan:
		t.Fatalf("unexpected callback")
		return false
	case <-time.After(callbackTimeout):
		return true
	}
}

func TestUnlockEntryNotExists(t *testing.T) {
	testLockMap := NewLockMap()

	callbackChan1 := make(chan any)
	go testLockMap.lockAndCallback(t, "entry1", callbackChan1)
	ensureCallbackHappens(t, callbackChan1)
	// entry2 does not exist
	testLockMap.UnlockEntry("entry2")
	testLockMap.UnlockEntry("entry1")
}

func TestBytesToGiB(t *testing.T) {
	var sizeInBytes int64 = 5 * GiB

	actual := BytesToGiB(sizeInBytes)
	if actual != 5 {
		t.Fatalf("Wrong result for BytesToGiB. Got: %d", actual)
	}
}

func TestGiBToBytes(t *testing.T) {
	var sizeInGiB int64 = 3

	actual := GiBToBytes(sizeInGiB)
	if actual != 3*GiB {
		t.Fatalf("Wrong result for GiBToBytes. Got: %d", actual)
	}
}

func TestGetMountOptions(t *testing.T) {
	tests := []struct {
		options  []string
		expected string
	}{
		{
			options:  []string{"-o allow_other", "-o ro", "--use-https=true"},
			expected: "-o allow_other -o ro --use-https=true",
		},
		{
			options:  []string{"-o allow_other"},
			expected: "-o allow_other",
		},
		{
			options:  []string{""},
			expected: "",
		},
		{
			options:  []string{},
			expected: "",
		},
	}

	for _, test := range tests {
		result := GetMountOptions(test.options)
		if result != test.expected {
			t.Errorf("getMountOptions(%v) result: %s, expected: %s", test.options, result, test.expected)
		}
	}
}

func TestConvertTagsToMap(t *testing.T) {
	tests := []struct {
		desc          string
		tags          string
		expected      map[string]string
		expectedError error
	}{
		{
			desc:          "Invalid tag",
			tags:          "invalid,test,tag",
			expected:      nil,
			expectedError: errors.New("tags 'invalid,test,tag' are invalid, the format should be: 'key1=value1,key2=value2'"),
		},
		{
			desc:          "Invalid key",
			tags:          "=test",
			expected:      nil,
			expectedError: errors.New("tags '=test' are invalid, the format should be: 'key1=value1,key2=value2'"),
		},
		{
			desc:          "Valid tags",
			tags:          "testTag=testValue",
			expected:      map[string]string{"testTag": "testValue"},
			expectedError: nil,
		},
		{
			desc:          "Multiple tags",
			tags:          "key1=value1,key2=value2",
			expected:      map[string]string{"key1": "value1", "key2": "value2"},
			expectedError: nil,
		},
		{
			desc:          "Handles spaces",
			tags:          " key1 = value1 , key2 = value2 ",
			expected:      map[string]string{"key1": "value1", "key2": "value2"},
			expectedError: nil,
		},
		{
			desc:          "Empty tags",
			tags:          "",
			expected:      map[string]string{},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		_, err := ConvertTagsToMap(test.tags)
		if !reflect.DeepEqual(err, test.expectedError) {
			t.Errorf("test[%s]: unexpected error: %v, expected error: %v", test.desc, err, test.expectedError)
		}
	}
}

func TestSetKeyValueInMap(t *testing.T) {
	tests := []struct {
		desc     string
		m        map[string]string
		key      string
		value    string
		expected map[string]string
	}{
		{
			desc:  "nil map",
			key:   "key",
			value: "value",
		},
		{
			desc:     "empty map",
			m:        map[string]string{},
			key:      "key",
			value:    "value",
			expected: map[string]string{"key": "value"},
		},
		{
			desc:  "non-empty map",
			m:     map[string]string{"k": "v"},
			key:   "key",
			value: "value",
			expected: map[string]string{
				"k":   "v",
				"key": "value",
			},
		},
		{
			desc:     "same key already exists",
			m:        map[string]string{"subDir": "value2"},
			key:      "subDir",
			value:    "value",
			expected: map[string]string{"subDir": "value"},
		},
		{
			desc:     "case insensitive key already exists",
			m:        map[string]string{"subDir": "value2"},
			key:      "subdir",
			value:    "value",
			expected: map[string]string{"subDir": "value"},
		},
	}

	for _, test := range tests {
		SetKeyValueInMap(test.m, test.key, test.value)
		if !reflect.DeepEqual(test.m, test.expected) {
			t.Errorf("test[%s]: unexpected output: %v, expected result: %v", test.desc, test.m, test.expected)
		}
	}
}

func TestGetValueInMap(t *testing.T) {
	tests := []struct {
		desc     string
		m        map[string]string
		key      string
		expected string
	}{
		{
			desc:     "nil map",
			key:      "key",
			expected: "",
		},
		{
			desc:     "empty map",
			m:        map[string]string{},
			key:      "key",
			expected: "",
		},
		{
			desc:     "non-empty map",
			m:        map[string]string{"k": "v"},
			key:      "key",
			expected: "",
		},
		{
			desc:     "same key already exists",
			m:        map[string]string{"subDir": "value2"},
			key:      "subDir",
			expected: "value2",
		},
		{
			desc:     "case insensitive key already exists",
			m:        map[string]string{"subDir": "value2"},
			key:      "subdir",
			expected: "value2",
		},
	}

	for _, test := range tests {
		result := GetValueInMap(test.m, test.key)
		if result != test.expected {
			t.Errorf("test[%s]: unexpected output: %v, expected result: %v", test.desc, result, test.expected)
		}
	}
}

func TestReplaceWithMap(t *testing.T) {
	pvcNameMetadata := "${pvc.metadata.name}"
	pvcNamespaceMetadata := "${pvc.metadata.namespace}"
	pvNameMetadata := "${pv.metadata.name}"

	tests := []struct {
		desc     string
		str      string
		m        map[string]string
		expected string
	}{
		{
			desc:     "empty string",
			str:      "",
			expected: "",
		},
		{
			desc:     "empty map",
			str:      "",
			m:        map[string]string{},
			expected: "",
		},
		{
			desc:     "empty key",
			str:      "prefix-" + pvNameMetadata,
			m:        map[string]string{"": "pv"},
			expected: "prefix-" + pvNameMetadata,
		},
		{
			desc:     "empty value",
			str:      "prefix-" + pvNameMetadata,
			m:        map[string]string{pvNameMetadata: ""},
			expected: "prefix-",
		},
		{
			desc:     "one replacement",
			str:      "prefix-" + pvNameMetadata,
			m:        map[string]string{pvNameMetadata: "pv"},
			expected: "prefix-pv",
		},
		{
			desc:     "multiple replacements",
			str:      pvcNamespaceMetadata + pvcNameMetadata,
			m:        map[string]string{pvcNamespaceMetadata: "namespace", pvcNameMetadata: "pvcname"},
			expected: "namespacepvcname",
		},
	}

	for _, test := range tests {
		result := ReplaceWithMap(test.str, test.m)
		if result != test.expected {
			t.Errorf("test[%s]: unexpected output: %v, expected result: %v", test.desc, result, test.expected)
		}
	}
}

func TestMakeDir(t *testing.T) {
	// Successfully create directory
	targetTest := "./target_test"
	err := MakeDir(targetTest)
	require.NoError(t, err)

	// Check if directory exists
	_, err = os.Stat(targetTest)
	require.False(t, os.IsNotExist(err))

	// Remove the directory created
	err = os.RemoveAll(targetTest)
	require.NoError(t, err)
}

func TestMakeDirAlreadyExists(t *testing.T) {
	// Create directory
	targetTest := "./target_test"
	err := MakeDir(targetTest)
	require.NoError(t, err)

	// Try to create the same directory again
	err = MakeDir(targetTest)
	require.NoError(t, err)

	// Remove the directory created
	err = os.RemoveAll(targetTest)
	require.NoError(t, err)
}

func TestMakeDirInvalidPath(t *testing.T) {
	// Try to create a directory with an invalid path
	invalidPath := string([]byte{0})
	err := MakeDir(invalidPath)
	require.Error(t, err)
}
