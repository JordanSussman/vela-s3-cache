// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"testing"
)

func TestS3Cache_Flush_Validate(t *testing.T) {
	// setup types
	f := &Flush{
		Root: "bucket",
	}

	err := f.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Flush_Validate_NoRoot(t *testing.T) {
	// setup types
	f := &Flush{}

	err := f.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}
