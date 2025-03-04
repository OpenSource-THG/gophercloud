package testing

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	accountTesting "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/accounts/testing"
	containerTesting "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers/testing"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	fake "github.com/gophercloud/gophercloud/testhelper/client"
)

func TestDownloadReader(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleDownloadObjectSuccessfully(t)

	response := objects.Download(fake.ServiceClient(), "testContainer", "testObject", nil)
	defer response.Body.Close()

	// Check reader
	buf := bytes.NewBuffer(make([]byte, 0))
	io.CopyN(buf, response.Body, 10)
	th.CheckEquals(t, "Successful", string(buf.Bytes()))
}

func TestDownloadExtraction(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleDownloadObjectSuccessfully(t)

	response := objects.Download(fake.ServiceClient(), "testContainer", "testObject", nil)

	// Check []byte extraction
	bytes, err := response.ExtractContent()
	th.AssertNoErr(t, err)
	th.CheckEquals(t, "Successful download with Gophercloud", string(bytes))

	expected := &objects.DownloadHeader{
		ContentLength:     36,
		ContentType:       "text/plain; charset=utf-8",
		Date:              time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		StaticLargeObject: true,
		LastModified:      time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	}
	actual, err := response.Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, expected, actual)
}

func TestDownloadWithLastModified(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleDownloadObjectSuccessfully(t)

	options1 := &objects.DownloadOpts{
		IfUnmodifiedSince: time.Date(2009, time.November, 10, 22, 59, 59, 0, time.UTC),
	}
	response1 := objects.Download(fake.ServiceClient(), "testContainer", "testObject", options1)
	_, err1 := response1.Extract()
	th.AssertErr(t, err1)

	options2 := &objects.DownloadOpts{
		IfModifiedSince: time.Date(2009, time.November, 10, 23, 0, 1, 0, time.UTC),
	}
	response2 := objects.Download(fake.ServiceClient(), "testContainer", "testObject", options2)
	content, err2 := response2.ExtractContent()
	th.AssertNoErr(t, err2)
	th.AssertEquals(t, 0, len(content))
}

func TestListObjectInfo(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListObjectsInfoSuccessfully(t)

	count := 0
	options := &objects.ListOpts{Full: true}
	err := objects.List(fake.ServiceClient(), "testContainer", options).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := objects.ExtractInfo(page)
		th.AssertNoErr(t, err)

		th.CheckDeepEquals(t, ExpectedListInfo, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 1, count)
}

func TestListObjectSubdir(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListSubdirSuccessfully(t)

	count := 0
	options := &objects.ListOpts{Full: true, Prefix: "", Delimiter: "/"}
	err := objects.List(fake.ServiceClient(), "testContainer", options).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := objects.ExtractInfo(page)
		th.AssertNoErr(t, err)

		th.CheckDeepEquals(t, ExpectedListSubdir, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 1, count)
}

func TestListObjectNames(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListObjectNamesSuccessfully(t)

	// Check without delimiter.
	count := 0
	options := &objects.ListOpts{Full: false}
	err := objects.List(fake.ServiceClient(), "testContainer", options).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := objects.ExtractNames(page)
		if err != nil {
			t.Errorf("Failed to extract container names: %v", err)
			return false, err
		}

		th.CheckDeepEquals(t, ExpectedListNames, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 1, count)

	// Check with delimiter.
	count = 0
	options = &objects.ListOpts{Full: false, Delimiter: "/"}
	err = objects.List(fake.ServiceClient(), "testContainer", options).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := objects.ExtractNames(page)
		if err != nil {
			t.Errorf("Failed to extract container names: %v", err)
			return false, err
		}

		th.CheckDeepEquals(t, ExpectedListNames, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 1, count)
}

func TestListZeroObjectNames204(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListZeroObjectNames204(t)

	count := 0
	options := &objects.ListOpts{Full: false}
	err := objects.List(fake.ServiceClient(), "testContainer", options).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := objects.ExtractNames(page)
		if err != nil {
			t.Errorf("Failed to extract container names: %v", err)
			return false, err
		}

		th.CheckDeepEquals(t, []string{}, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, 0, count)
}

func TestCreateObject(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	content := "Did gyre and gimble in the wabe"

	HandleCreateTextObjectSuccessfully(t, content)

	options := &objects.CreateOpts{ContentType: "text/plain", Content: strings.NewReader(content)}
	res := objects.Create(fake.ServiceClient(), "testContainer", "testObject", options)
	th.AssertNoErr(t, res.Err)
}

func TestCreateObjectWithCacheControl(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	content := "All mimsy were the borogoves"

	HandleCreateTextWithCacheControlSuccessfully(t, content)

	options := &objects.CreateOpts{
		CacheControl: `max-age="3600", public`,
		Content:      strings.NewReader(content),
	}
	res := objects.Create(fake.ServiceClient(), "testContainer", "testObject", options)
	th.AssertNoErr(t, res.Err)
}

func TestCreateObjectWithoutContentType(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	content := "The sky was the color of television, tuned to a dead channel."

	HandleCreateTypelessObjectSuccessfully(t, content)

	res := objects.Create(fake.ServiceClient(), "testContainer", "testObject", &objects.CreateOpts{Content: strings.NewReader(content)})
	th.AssertNoErr(t, res.Err)
}

/*
func TestErrorIsRaisedForChecksumMismatch(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/testContainer/testObject", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "acbd18db4cc2f85cedef654fccc4a4d8")
		w.WriteHeader(http.StatusCreated)
	})

	content := strings.NewReader("The sky was the color of television, tuned to a dead channel.")
	res := Create(fake.ServiceClient(), "testContainer", "testObject", &CreateOpts{Content: content})

	err := fmt.Errorf("Local checksum does not match API ETag header")
	th.AssertDeepEquals(t, err, res.Err)
}
*/

func TestCopyObject(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleCopyObjectSuccessfully(t)

	options := &objects.CopyOpts{Destination: "/newTestContainer/newTestObject"}
	res := objects.Copy(fake.ServiceClient(), "testContainer", "testObject", options)
	th.AssertNoErr(t, res.Err)
}

func TestDeleteObject(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleDeleteObjectSuccessfully(t)

	res := objects.Delete(fake.ServiceClient(), "testContainer", "testObject", nil)
	th.AssertNoErr(t, res.Err)
}

func TestBulkDelete(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleBulkDeleteSuccessfully(t)

	expected := objects.BulkDeleteResponse{
		ResponseStatus: "foo",
		ResponseBody:   "bar",
		NumberDeleted:  2,
		Errors:         [][]string{},
	}

	resp, err := objects.BulkDelete(fake.ServiceClient(), "testContainer", []string{"testObject1", "testObject2"}).Extract()
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, expected, *resp)
}

func TestUpateObjectMetadata(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleUpdateObjectSuccessfully(t)

	s := new(string)
	i := new(int64)
	options := &objects.UpdateOpts{
		Metadata:           map[string]string{"Gophercloud-Test": "objects"},
		RemoveMetadata:     []string{"Gophercloud-Test-Remove"},
		ContentDisposition: s,
		ContentEncoding:    s,
		ContentType:        s,
		DeleteAt:           i,
		DetectContentType:  new(bool),
	}
	res := objects.Update(fake.ServiceClient(), "testContainer", "testObject", options)
	th.AssertNoErr(t, res.Err)
}

func TestGetObject(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleGetObjectSuccessfully(t)

	expected := map[string]string{"Gophercloud-Test": "objects"}
	actual, err := objects.Get(fake.ServiceClient(), "testContainer", "testObject", nil).ExtractMetadata()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, expected, actual)

	getOpts := objects.GetOpts{
		Newest: true,
	}
	actualHeaders, err := objects.Get(fake.ServiceClient(), "testContainer", "testObject", getOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, true, actualHeaders.StaticLargeObject)
}

func TestETag(t *testing.T) {
	content := "some example object"
	createOpts := objects.CreateOpts{
		Content: strings.NewReader(content),
		NoETag:  true,
	}

	_, headers, _, err := createOpts.ToObjectCreateParams()
	th.AssertNoErr(t, err)
	_, ok := headers["ETag"]
	th.AssertEquals(t, false, ok)

	hash := md5.New()
	io.WriteString(hash, content)
	localChecksum := fmt.Sprintf("%x", hash.Sum(nil))

	createOpts = objects.CreateOpts{
		Content: strings.NewReader(content),
		ETag:    localChecksum,
	}

	_, headers, _, err = createOpts.ToObjectCreateParams()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, localChecksum, headers["ETag"])
}

func TestObjectCreateParamsWithoutSeek(t *testing.T) {
	content := "I do not implement Seek()"
	buf := bytes.NewBuffer([]byte(content))

	createOpts := objects.CreateOpts{Content: buf}
	reader, headers, _, err := createOpts.ToObjectCreateParams()

	th.AssertNoErr(t, err)

	_, ok := reader.(io.ReadSeeker)
	th.AssertEquals(t, true, ok)

	c, err := ioutil.ReadAll(reader)
	th.AssertNoErr(t, err)

	th.AssertEquals(t, content, string(c))

	_, ok = headers["ETag"]
	th.AssertEquals(t, true, ok)
}

func TestObjectCreateParamsWithSeek(t *testing.T) {
	content := "I implement Seek()"
	createOpts := objects.CreateOpts{Content: strings.NewReader(content)}
	reader, headers, _, err := createOpts.ToObjectCreateParams()

	th.AssertNoErr(t, err)

	_, ok := reader.(io.ReadSeeker)
	th.AssertEquals(t, ok, true)

	c, err := ioutil.ReadAll(reader)
	th.AssertNoErr(t, err)

	th.AssertEquals(t, content, string(c))

	_, ok = headers["ETag"]
	th.AssertEquals(t, true, ok)
}

func TestCreateTempURL(t *testing.T) {
	port := 33200
	th.SetupHTTP()
	th.SetupPersistentPortHTTP(t, port)
	defer th.TeardownHTTP()

	// Handle fetching of secret key inside of CreateTempURL
	containerTesting.HandleGetContainerSuccessfully(t)
	accountTesting.HandleGetAccountSuccessfully(t)
	client := fake.ServiceClient()

	// Append v1/ to client endpoint URL to be compliant with tempURL generator
	client.Endpoint = client.Endpoint + "v1/"
	tempURL, err := objects.CreateTempURL(client, "testContainer", "testObject/testFile.txt", objects.CreateTempURLOpts{
		Method:    http.MethodGet,
		TTL:       60,
		Timestamp: time.Date(2020, 07, 01, 01, 12, 00, 00, time.UTC),
	})

	sig := "89be454a9c7e2e9f3f50a8441815e0b5801cba5b"
	expiry := "1593565980"
	expectedURL := fmt.Sprintf("http://127.0.0.1:%v/v1/testContainer/testObject/testFile.txt?temp_url_sig=%v&temp_url_expires=%v", port, sig, expiry)

	th.AssertNoErr(t, err)
	th.AssertEquals(t, expectedURL, tempURL)
}
