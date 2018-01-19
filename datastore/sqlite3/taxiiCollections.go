// Copyright 2017 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file in the root of the source tree.

package sqlite3

import (
	"errors"
	"fmt"
	"github.com/freetaxii/libstix2/datastore"
	"github.com/freetaxii/libstix2/defs"
	"github.com/freetaxii/libstix2/objects"
	"github.com/freetaxii/libstix2/resources"
	"strings"
	"time"
)

// ----------------------------------------------------------------------
//
// Public Methods
//
// ----------------------------------------------------------------------

/*
GetAllCollections - This method will return all collections, even those that
are disabled and hidden. This is primarily used for administration tools that
need to see all collections.
*/
func (ds *Sqlite3DatastoreType) GetAllCollections() (*resources.CollectionsType, error) {
	return ds.getCollections("all")
}

/*
GetAllEnabledCollections - This method will return only enabled collections,
even those that are hidden. This is used for setup up the HTTP MUX routers.
*/
func (ds *Sqlite3DatastoreType) GetAllEnabledCollections() (*resources.CollectionsType, error) {
	return ds.getCollections("allEnabled")
}

/*
GetCollections - This method will return just those collections that are both
enabled and visible. This is primarily used for client that pull a collections
resource.
*/
func (ds *Sqlite3DatastoreType) GetCollections() (*resources.CollectionsType, error) {
	return ds.getCollections("enabledVisible")
}

/*
GetBundle - This method will take in a query struct with range
parameters for a collection and will return a STIX Bundle that contains all
of the STIX objects that are in that collection that meet those query or range
parameters.
*/
func (ds *Sqlite3DatastoreType) GetBundle(query datastore.QueryType) (*objects.BundleType, *datastore.QueryReturnDataType, error) {

	stixBundle := objects.InitBundle()

	rangeCollectionRawData, metaData, err := ds.GetObjectList(query)
	if err != nil {
		return nil, nil, err
	}

	for _, v := range *rangeCollectionRawData {
		// Only get the objects that are part of the response
		obj, err := ds.GetObject(v.STIXID, v.STIXVersion)

		if err != nil {
			return nil, nil, err
		}
		stixBundle.AddObject(obj)
	}

	return stixBundle, metaData, nil
}

/*
GetObjectList - This method will take in a query struct with range
parameters for a collection and will return a datastore collection raw data type
that contains all of the STIX IDs and their associated meta data that are in
that collection that meet those query or range parameters.
*/
func (ds *Sqlite3DatastoreType) GetObjectList(query datastore.QueryType) (*[]datastore.CollectionRawDataType, *datastore.QueryReturnDataType, error) {
	var metaData datastore.QueryReturnDataType
	var collectionRawData []datastore.CollectionRawDataType
	var rangeCollectionRawData []datastore.CollectionRawDataType

	sqlStmt, err := sqlGetObjectList(query)

	// If an error is found, that means a query parameter was passed incorrectly
	// and we should return an error versus just skipping the option.
	if err != nil {
		return nil, nil, err
	}

	// Query database for all the collection entries
	rows, err := ds.DB.Query(sqlStmt)
	if err != nil {
		return nil, nil, fmt.Errorf("database execution error querying collection content: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dateAdded, stixid, modified, specVersion string
		if err := rows.Scan(&dateAdded, &stixid, &modified, &specVersion); err != nil {
			return nil, nil, fmt.Errorf("database scan error: ", err)
		}
		var rawData datastore.CollectionRawDataType
		rawData.STIXID = stixid
		rawData.DateAdded = dateAdded
		rawData.STIXVersion = modified
		rawData.SpecVersion = specVersion

		collectionRawData = append(collectionRawData, rawData)
	}

	metaData.Size = len(collectionRawData)

	// If no records are returned, then return an error before processing anything else.
	if metaData.Size == 0 {
		return nil, nil, errors.New("no records returned")
	}

	first, last, errRange := ds.processRangeValues(query.RangeBegin, query.RangeEnd, query.RangeMax, metaData.Size)

	if errRange != nil {
		return nil, nil, errRange
	}

	// Get a new slice based on the range of records
	rangeCollectionRawData = collectionRawData[first:last]
	metaData.DateAddedFirst = rangeCollectionRawData[0].DateAdded
	metaData.DateAddedLast = rangeCollectionRawData[len(rangeCollectionRawData)-1].DateAdded
	metaData.RangeBegin = first
	metaData.RangeEnd = last - 1

	// metaData is already a pointer
	return &rangeCollectionRawData, &metaData, nil
}

/*
GetManifestData - This method will take in query struct with range
parameters for a collection and will return a TAXII manifest that contains all
of the records that match the query and range parameters.
*/
func (ds *Sqlite3DatastoreType) GetManifestData(query datastore.QueryType) (*resources.ManifestType, *datastore.QueryReturnDataType, error) {
	manifest := resources.InitManifest()
	rangeManifest := resources.InitManifest()
	var metaData datastore.QueryReturnDataType
	var first, last int
	var errRange error

	sqlStmt, err := sqlGetManifestData(query)

	// If an error is found, that means a query parameter was passed incorrectly
	// and we should return an error versus just skipping the option.
	if err != nil {
		return nil, nil, err
	}

	// Query database for all the collection entries
	rows, err := ds.DB.Query(sqlStmt)
	if err != nil {
		return nil, nil, fmt.Errorf("database execution error querying collection content: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dateAdded, stixid, modified, specVersion string
		if err := rows.Scan(&dateAdded, &stixid, &modified, &specVersion); err != nil {
			return nil, nil, fmt.Errorf("database scan error: ", err)
		}
		manifest.CreateManifestEntry(stixid, dateAdded, modified, specVersion)
	}

	metaData.Size = len(manifest.Objects)

	first, last, errRange = ds.processRangeValues(query.RangeBegin, query.RangeEnd, query.RangeMax, metaData.Size)

	if errRange != nil {
		return nil, nil, errRange
	}

	// Get a new slice based on the range of records
	rangeManifest.Objects = manifest.Objects[first:last]
	metaData.DateAddedFirst = rangeManifest.Objects[0].DateAdded
	metaData.DateAddedLast = rangeManifest.Objects[len(rangeManifest.Objects)-1].DateAdded

	return rangeManifest, &metaData, nil
}

// ----------------------------------------------------------------------
//
// Private Methods
//
// ----------------------------------------------------------------------

/*
addCollection - This method will add a collection to the t_collections table in
the database.
*/
func (ds *Sqlite3DatastoreType) addCollection(obj *resources.CollectionType) error {
	dateAdded := time.Now().UTC().Format(defs.TIME_RFC_3339_MICRO)

	stmt1, _ := sqlAddCollection()

	_, err1 := ds.DB.Exec(stmt1,
		dateAdded,
		obj.ID,
		obj.Title,
		obj.Description,
		obj.CanRead,
		obj.CanWrite)

	if err1 != nil {
		return fmt.Errorf("database execution error inserting collection", err1)
	}

	if obj.MediaTypes != nil {
		for _, media := range obj.MediaTypes {
			stmt2, _ := sqlAddCollectionMediaType()

			// TODO look up in cache
			mediavalue := 0
			if media == "application/vnd.oasis.stix+json" {
				mediavalue = 1
			}
			_, err2 := ds.DB.Exec(stmt2, obj.ID, mediavalue)

			if err2 != nil {
				return fmt.Errorf("database execution error inserting collection media type", err2)
			}
		}
	}
	return nil
}

/*
addObjectToColleciton - This method will add an object to a collection by adding
an entry in the taxii_collection_data table. In this table we use the STIX ID
not the Object ID because we need to make sure we include all versions of an
object. So we need to store just the STIX ID.
*/
func (ds *Sqlite3DatastoreType) addObjectToCollection(obj *resources.CollectionRecordType) error {
	dateAdded := time.Now().UTC().Format(defs.TIME_RFC_3339_MICRO)

	stmt, _ := sqlAddObjectToCollection()
	_, err := ds.DB.Exec(stmt, dateAdded, obj.CollectionID, obj.STIXID)

	if err != nil {
		return fmt.Errorf("database execution error inserting collection data", err)
	}
	return nil
}

/*
getCollections - This method is called from either GetAllCollections(),
GetAllEnabledCollections(), or GetCollections() and will return all of the
collections that are asked for based on the method that called it.  The options
that can be passed in are: "all", "allEnabled", and "enabledVisible". The "all"
option returns every collection, even those that are hidden or disabled.
"allEnabled" will return all enabled collections, even those that are hidden.
"enabledVisible" will return all collections that are both enabled and not
hidden (aka those that are visible). Administration tools using the database
will probably want to see all collections. The HTTP Router MUX needs to know
about all enabled collections, even those that are hidden, so that it can start
an HTTP router for it. The enabled and visible list is what would be displayed
to a client that is pulling a collections resource.
*/
func (ds *Sqlite3DatastoreType) getCollections(whichCollections string) (*resources.CollectionsType, error) {

	allCollections := resources.InitCollections()

	getAllCollectionsStmt, _ := sqlGetAllCollections(whichCollections)

	// Query database for all the collections
	rows, err := ds.DB.Query(getAllCollectionsStmt)
	if err != nil {
		return nil, fmt.Errorf("database execution error querying collection: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var enabled, hidden, iCanRead, iCanWrite int
		var dateAdded, id, title, description, mediaType string
		if err := rows.Scan(&dateAdded, &enabled, &hidden, &id, &title, &description, &iCanRead, &iCanWrite, &mediaType); err != nil {
			return nil, fmt.Errorf("database scan error querying collection: ", err)
		}

		// Add collection information to Collection object
		c, _ := allCollections.GetNewCollection()
		c.DateAdded = dateAdded
		if enabled == 1 {
			c.SetEnabled()
		} else {
			c.SetDisabled()
		}

		if hidden == 1 {
			c.SetHidden()
		} else {
			c.SetVisible()
		}

		c.SetID(id)
		c.SetTitle(title)
		c.SetDescription(description)
		if iCanRead == 1 {
			c.SetCanRead()
		}
		if iCanWrite == 1 {
			c.SetCanWrite()
		}

		mediatypes := strings.Split(mediaType, ",")
		for i, mt := range mediatypes {

			// If the media types are all the same, due to the way the SQL query
			// returns results, then only record one entry.
			if i > 0 && mt == mediatypes[i-1] {
				continue
			}
			c.AddMediaType(mt)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database row error querying collection: ", err)
	}

	return allCollections, nil
}

/*
processRangeValues - This method will take in the various range parameters and size
of the dataset and will return the correct first and last index values to be used.
*/
func (ds *Sqlite3DatastoreType) processRangeValues(first, last, max, size int) (int, int, error) {

	if first < 0 {
		return 0, 0, errors.New("the starting value can not be negative")
	}

	if first > last {
		return 0, 0, errors.New("the starting range value is larger than the ending range value")
	}

	if first >= size {
		return 0, 0, errors.New("the starting range value is out of scope")
	}

	// If no range is requested and the server is not forcing it, do nothing.
	if last == 0 && first == 0 && max != 0 {
		last = first + max
	} else {
		// We need to be inclusive of the last value that was provided
		last++
	}

	// If the last record requested is bigger than the total size of the data
	// set the last size to be the size of the data
	if last > size {
		last = size
	}

	// If the request is for more records than the max size will allow, then
	// compute where the new last record should be, but only if the server is
	// forcing a max size.
	if max != 0 && (last-first) > max {
		last = first + max
	}

	return first, last, nil
}