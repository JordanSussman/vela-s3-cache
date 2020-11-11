// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mholt/archiver"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

const restoreAction = "restore"

// Restore represents the plugin configuration for Restore information.
type Restore struct {
	// sets the name of the bucket
	Bucket string
	// sets the path for where to retrieve the object from
	Path string
	// sets the path for where to retrieve the object from
	Prefix string
	// sets the name of the cache object
	Filename string
	// sets the timeout on the call to s3
	Timeout time.Duration
	// will hold our final namespace for the path to the objects
	Namespace string
}

// Exec formats and runs the actions for restoring a cache in s3.
func (r *Restore) Exec(mc *minio.Client) error {
	logrus.Trace("running restore with provided configuration")

	logrus.Debugf("getting object info on bucket %s from path: %s", r.Bucket, r.Namespace)

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	// collect metadata on the object
	ok, err := mc.StatObject(ctx, r.Bucket, r.Namespace, minio.StatObjectOptions{})
	if ok.Key == "" {
		logrus.Error(err)
		return nil
	}

	logrus.Debugf("getting object in bucket %s from path: %s", r.Bucket, r.Namespace)

	// retrieve the object in specified path of the bucket
	reader, err := mc.GetObject(ctx, r.Bucket, r.Namespace, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	logrus.Debugf("creating file %s on system", r.Filename)

	// create a a file to save the object to
	localFile, err := os.Create(r.Filename)
	if err != nil {
		return err
	}
	defer localFile.Close()

	logrus.Debugf("get object of file %s", r.Filename)

	// get information on the object pulled down from the s3 store
	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	logrus.Debugf("copy object data to local file %s", r.Filename)

	// copy the object contents to the file created on system
	if _, err := io.CopyN(localFile, reader, stat.Size); err != nil {
		return err
	}

	logrus.Infof("copied %s to local filesystem", r.Filename)

	logrus.Debug("get current working directory")

	// grab the current working directory for unpacking the object
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Debugf("unarchiving file %s into directory %s", r.Filename, pwd)

	// expand the object back onto the filesystem
	err = archiver.Unarchive(r.Filename, pwd)
	if err != nil {
		return err
	}

	logrus.Infof("successfully unpacked archive %s", r.Filename)

	// delete the temporary archive file
	err = os.Remove(r.Filename)
	if err != nil {
		logrus.Infof("delete of archive file %s unsuccessful", r.Filename)
	} else {
		logrus.Infof("cache archive %s successfully deleted", r.Filename)
	}

	logrus.Infof("cache restore action completed")

	return nil
}

// Configure prepares the restore fields for the action to be taken.
func (r *Restore) Configure(repo *Repo) error {
	logrus.Trace("configuring restore action")

	// construct the object path
	path := buildNamespace(repo, r.Prefix, r.Path, r.Filename)

	logrus.Debugf("created bucket path %s", path)

	// store it in the namespace
	r.Namespace = path

	return nil
}

// Validate verifies the Restore is properly configured.
func (r *Restore) Validate() error {
	logrus.Trace("validating restore action configuration")

	// verify bucket is provided
	if len(r.Bucket) == 0 {
		return fmt.Errorf("no bucket provided")
	}

	// verify filename is provided
	if len(r.Filename) == 0 {
		return fmt.Errorf("no filename provided")
	}

	// verify timeout is provided
	if strings.EqualFold(r.Timeout.String(), "0s") {
		return fmt.Errorf("timeout must be greater than 0")
	}

	return nil
}
