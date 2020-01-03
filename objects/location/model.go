// Copyright 2015-2020 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file in the root of the source tree.

package location

import (
	"github.com/freetaxii/libstix2/objects"
	"github.com/freetaxii/libstix2/objects/properties"
)

// ----------------------------------------------------------------------
// Define Object Model
// ----------------------------------------------------------------------

/* Location - This type implements the STIX 2 Location SDO and
defines all of the properties and methods needed to create and work with this
object. All of the methods not defined local to this type are inherited from the
individual properties. */
type Location struct {
	objects.CommonObjectProperties
	properties.NameProperty
	properties.DescriptionProperty
	Latitude           float64 `json:"latitude,omitempty"`
	Longitude          float64 `json:"longitude,omitempty"`
	Precision          float64 `json:"precision,omitempty"`
	Region             string  `json:"region,omitempty"`
	Country            string  `json:"country,omitempty"`
	AdministrativeArea string  `json:"administrative_area,omitempty"`
	City               string  `json:"city,omitempty"`
	StreetAddress      string  `json:"street_address,omitempty"`
	PostalCode         string  `json:"postal_code,omitempty"`
}

// ----------------------------------------------------------------------
// Initialization Functions
// ----------------------------------------------------------------------

/* New - This function will create a new STIX Location object and return
it as a pointer. It will also initialize the object by setting all of the basic
properties. */
func New() *Location {
	var obj Location
	obj.InitSDO("location")
	return &obj
}
