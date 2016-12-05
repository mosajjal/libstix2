// Copyright 2016 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package report

import (
	"github.com/freetaxii/libstix2/objects/common"
)

// ----------------------------------------------------------------------
// Define Message Type
// ----------------------------------------------------------------------

type ReportType struct {
	common.CommonPropertiesType
	common.DescriptivePropertiesType
	Published   string   `json:"published,omitempty"`
	Object_refs []string `json:"object_refs,omitempty"`
}

// ----------------------------------------------------------------------
// Public Create Functions
// ----------------------------------------------------------------------

func New() ReportType {
	var obj ReportType
	obj.MessageType = "report"
	obj.Id = obj.NewId("report")
	obj.Created = obj.GetCurrentTime()
	obj.Modified = obj.Created
	obj.Version = 1
	return obj
}

// ----------------------------------------------------------------------
// Public Methods - ReportType
// ----------------------------------------------------------------------

// SetPublished takes in two parameters and returns and error if there is one
// param: t a timestamp in either time.Time or string format
func (this *ReportType) SetPublished(t interface{}) error {

	ts, err := this.VerifyTimestamp(t)
	if err != nil {
		return err
	}
	this.Published = ts

	return nil
}

func (this *ReportType) AddObject(value string) {
	if this.Object_refs == nil {
		a := make([]string, 0)
		this.Object_refs = a
	}
	this.Object_refs = append(this.Object_refs, value)
}