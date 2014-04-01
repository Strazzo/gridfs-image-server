package server

import (
	"image"
	"image/jpeg"
	"labix.org/v2/mgo"
	. "launchpad.net/gocheck"
	"net/http"
	"os"
	"time"
)

type ServerTestSuite struct{}

var _ = Suite(&ServerTestSuite{})

var testMongoFile *mgo.GridFile
var TestCon *mgo.Session

// SetUpTest creates files for further tests to use
func (s *ServerTestSuite) SetUpTest(c *C) {
	filename, _ := os.Getwd()
	imageFile, err := os.Open(filename + "/../testdata/image.jpg")
	c.Assert(err, IsNil)
	TestCon, err = mgo.Dial("localhost")
	c.Assert(err, IsNil)
	TestCon.SetMode(mgo.Monotonic, true)

	tempMongo, mongoErr := TestCon.DB("unittest").GridFS("fs").Create("test.jpg")
	c.Assert(mongoErr, IsNil)

	dIm, _, _ := image.Decode(imageFile)

	jpeg.Encode(tempMongo, dIm, &jpeg.Options{JpegMaximumQuality})
	tempMongo.Close()

	var openErr error

	testMongoFile, openErr = TestCon.DB("unittest").GridFS("fs").Open("test.jpg")

	c.Assert(openErr, IsNil)

	c.Assert(testMongoFile.MD5(), Equals, "d5b390993a34a440891a6f20407f9dde")
}

// TearDownTest removes the created test file.
func (s *ServerTestSuite) TearDownTest(c *C) {
	TestCon, _ = mgo.Dial("localhost")
	Connection.DB("unittest").DropDatabase()
}

// TestIsModifiedNoCache
func (s *ServerTestSuite) TestIsModifiedNoHeaders(c *C) {
	header := http.Header{}

	c.Assert(isModified(testMongoFile, &header), Equals, true)
}

// TestIsModifiedNoCache
func (s *ServerTestSuite) TestIsModifiedNoCache(c *C) {

	header := http.Header{}
	header.Set("Cache-Control", "no-cache")

	c.Assert(isModified(testMongoFile, &header), Equals, true)
}

// TestIsModifiedMd5Mismatch
func (s *ServerTestSuite) TestIsModifiedMd5Mismatch(c *C) {

	header := http.Header{}
	header.Set("If-None-Match", "invalid md5")

	c.Assert(isModified(testMongoFile, &header), Equals, true)
}

// TestCacheToOldHeader
func (s *ServerTestSuite) TestCacheToOldHeader(c *C) {
	modified := time.Unix(0, 0).Format(time.RFC1123)

	header := http.Header{}
	header.Set("If-None-Match", testMongoFile.MD5())
	header.Set("If-Modified-Since", modified)

	c.Assert(isModified(testMongoFile, &header), Equals, true)
}

// TestCacheHitSuccess
func (s *ServerTestSuite) TestCacheHitSuccess(c *C) {
	modified := time.Now().Format(time.RFC1123)

	header := http.Header{}
	header.Set("If-None-Match", testMongoFile.MD5())
	header.Set("If-Modified-Since", modified)

	c.Assert(isModified(testMongoFile, &header), Equals, false)
}
