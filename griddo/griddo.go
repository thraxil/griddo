package griddo

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type Grid struct {
	Owner    *datastore.Key
	Title    string
	Created  time.Time
	Modified time.Time
}

type Dimension struct {
	Grid         *datastore.Key
	Label        string
	DisplayOrder int
}

type Slice struct {
	Dimension    *datastore.Key
	Label        string
	DisplayOrder int
}

type Cell struct {
	Value    int
	Created  time.Time
	Modified time.Time
}

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}
