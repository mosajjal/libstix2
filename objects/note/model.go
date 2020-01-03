// Copyright 2015-2020 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file in the root of the source tree.

package note

import (
	"github.com/freetaxii/libstix2/objects"
	"github.com/freetaxii/libstix2/objects/properties"
)

// ----------------------------------------------------------------------
// Define Object Model
// ----------------------------------------------------------------------

/* Note - This type implements the STIX 2 Note SDO and
defines all of the properties and methods needed to create and work with this
object. All of the methods not defined local to this type are inherited from the
individual properties. */
type Note struct {
	objects.CommonObjectProperties
	Abstract string `json:"abstract,omitempty"`
	Content  string `json:"content,omitempty"`
	properties.AuthorsProperty
	properties.ObjectRefsProperty
}

// ----------------------------------------------------------------------
// Initialization Functions
// ----------------------------------------------------------------------

/* New - This function will create a new STIX Note object and return
it as a pointer. It will also initialize the object by setting all of the basic
properties. */
func New() *Note {
	var obj Note
	obj.InitSDO("note")
	return &obj
}
