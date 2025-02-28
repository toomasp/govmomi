/*
Copyright (c) 2018 VMware, Inc. All Rights Reserved.

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

package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// VAPI REST Paths
const (
	Path                           = "/rest"
	SessionPath                    = "/com/vmware/cis/session"
	CategoryPath                   = "/com/vmware/cis/tagging/category"
	TagPath                        = "/com/vmware/cis/tagging/tag"
	AssociationPath                = "/com/vmware/cis/tagging/tag-association"
	LibraryPath                    = "/com/vmware/content/library"
	LibraryItemFileData            = "/com/vmware/cis/data"
	LibraryItemPath                = "/com/vmware/content/library/item"
	LibraryItemFilePath            = "/com/vmware/content/library/item/file"
	LibraryItemUpdateSession       = "/com/vmware/content/library/item/update-session"
	LibraryItemUpdateSessionFile   = "/com/vmware/content/library/item/updatesession/file"
	LibraryItemDownloadSession     = "/com/vmware/content/library/item/download-session"
	LibraryItemDownloadSessionFile = "/com/vmware/content/library/item/downloadsession/file"
	LocalLibraryPath               = "/com/vmware/content/local-library"
	SubscribedLibraryPath          = "/com/vmware/content/subscribed-library"
	VCenterOVFLibraryItem          = "/com/vmware/vcenter/ovf/library-item"
	SessionCookieName              = "vmware-api-session-id"
)

// AssociatedObject is the same structure as types.ManagedObjectReference,
// just with a different field name (ID instead of Value).
// In the API we use mo.Reference, this type is only used for wire transfer.
type AssociatedObject struct {
	Type  string `json:"type"`
	Value string `json:"id"`
}

// Reference implements mo.Reference
func (o AssociatedObject) Reference() types.ManagedObjectReference {
	return types.ManagedObjectReference(o)
}

// Association for tag-association requests.
type Association struct {
	ObjectID *AssociatedObject `json:"object_id,omitempty"`
}

// NewAssociation returns an Association, converting ref to an AssociatedObject.
func NewAssociation(ref mo.Reference) Association {
	obj := AssociatedObject(ref.Reference())
	return Association{
		ObjectID: &obj,
	}
}

// CloneURL defines an interface for cloned urls
type CloneURL interface {
	URL() *url.URL
}

// Resource wraps url.URL with helpers
type Resource struct {
	u *url.URL
}

// URL creates a URL resource
func URL(c CloneURL, path string) *Resource {
	r := &Resource{u: c.URL()}
	r.u.Path = Path + path
	return r
}

func (r *Resource) String() string {
	return r.u.String()
}

// WithID appends id to the URL.Path
func (r *Resource) WithID(id string) *Resource {
	r.u.Path += "/id:" + id
	return r
}

// WithAction sets adds action to the URL.RawQuery
func (r *Resource) WithAction(action string) *Resource {
	r.u.RawQuery = url.Values{
		"~action": []string{action},
	}.Encode()
	return r
}

// WithParameter sets adds a parameter to the URL.RawQuery
func (r *Resource) WithParameter(name string, value string) *Resource {
	parameter := url.Values{}
	parameter.Set(name, value)
	r.u.RawQuery = parameter.Encode()
	return r
}

// Request returns a new http.Request for the given method.
// An optional body can be provided for POST and PATCH methods.
func (r *Resource) Request(method string, body ...interface{}) *http.Request {
	rdr := io.MultiReader() // empty body by default
	if len(body) != 0 {
		rdr = encode(body[0])
	}
	req, err := http.NewRequest(method, r.u.String(), rdr)
	if err != nil {
		panic(err)
	}
	return req
}

type errorReader struct {
	e error
}

func (e errorReader) Read([]byte) (int, error) {
	return -1, e.e
}

// encode body as JSON, deferring any errors until io.Reader is used.
func encode(body interface{}) io.Reader {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(body)
	if err != nil {
		return errorReader{err}
	}
	return &b
}
