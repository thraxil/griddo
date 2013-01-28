package griddo

import (
	"appengine"
	"appengine/datastore"
	_ "appengine/user"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Grid struct {
	//	Owner string
	Title string
	//	Created  time.Time
	//	Modified time.Time
}

type Row struct {
	Grid         *datastore.Key
	Label        string
	DisplayOrder int
}

type Col struct {
	Grid         *datastore.Key
	Label        string
	DisplayOrder int
}

type Cell struct {
	Value int
	Row   *datastore.Key
	Col   *datastore.Key
	//	Created  time.Time
	//	Modified time.Time
}

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

func newKey() string {
	var N = 10
	r := make([]byte, N)
	var i = 0
	for i = 0; i < N; i++ {
		r[i] = chars[rand.Intn(len(chars))]
	}
	return string(r)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", index)
	http.HandleFunc("/new/", newGrid)
	http.HandleFunc("/grid/", showGrid)
	http.HandleFunc("/cellupdate/", cellUpdate)
}

func cellUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 5 {
		http.Error(w, "bad request", 404)
		return
	}
	//	gridkey := parts[2]
	//	k := datastore.NewKey(ctx, "Grid", gridkey, 0, nil)
	//	g := new(Grid)

	ridx, _ := strconv.Atoi(parts[3])
	cidx, _ := strconv.Atoi(parts[4])

	v, _ := strconv.Atoi(r.FormValue("v"))
	ctx.Errorf("value: %v", r.FormValue("v"))
	ctx.Errorf("setting cell (%d,%d) to %d", ridx, cidx, v)
	fmt.Fprint(w, "ok")
}

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var indexTemplate = template.Must(template.New("index").Parse(indexTmpl))

	err := indexTemplate.Execute(w, map[string]string{})
	if err != nil {
		c.Errorf("indexTemplate: %v", err)
	}
}

func newGrid(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	title := r.FormValue("title")
	rows := strings.Split(r.FormValue("rows"), "\n")
	cols := strings.Split(r.FormValue("cols"), "\n")
	key := newKey()
	k := datastore.NewKey(c, "Grid", key, 0, nil)
	g := new(Grid)
	g.Title = title
	_, err := datastore.Put(c, k, g)
	if err != nil {
		c.Errorf("error adding grid: %v", err)
	}
	for i, r := range rows {
		rkey := datastore.NewKey(c, "Row", newKey(), 0, nil)
		row := new(Row)
		row.Grid = k
		row.Label = r
		row.DisplayOrder = i
		_, err := datastore.Put(c, rkey, row)
		c.Errorf("added row %v", err)
	}
	for i, co := range cols {
		ckey := datastore.NewKey(c, "Col", newKey(), 0, nil)
		col := new(Col)
		col.Grid = k
		col.Label = co
		col.DisplayOrder = i
		_, err := datastore.Put(c, ckey, col)
		c.Errorf("added col %v", err)
	}

	http.Redirect(w, r, "/grid/"+key, http.StatusFound)
}

type gridPage struct {
	Grid    *Grid
	GridKey string
	Rows    []Row
	Cols    []Col
}

func showGrid(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	gridkey := parts[2]
	k := datastore.NewKey(c, "Grid", gridkey, 0, nil)
	g := new(Grid)
	err := datastore.Get(c, k, g)
	if err != nil {
		http.Error(w, "Couldn't load Grid", http.StatusInternalServerError)
		c.Errorf("setting up: %v", err)
		c.Errorf("key: %v", parts)
		return
	}

	rq := datastore.NewQuery("Row").Filter("Grid=", k)
	rows := make([]Row, 0, 100)
	if _, err := rq.GetAll(c, &rows); err != nil {
		// handle the error
		c.Errorf("rows fetch: %v", err)
	}

	cq := datastore.NewQuery("Col").Filter("Grid=", k)
	cols := make([]Col, 0, 100)
	if _, err := cq.GetAll(c, &cols); err != nil {
		// handle the error
		c.Errorf("cols fetch: %v", err)
	}

	var gridTemplate = template.Must(template.New("grid").Parse(gridTmpl))

	err = gridTemplate.Execute(w, gridPage{g, gridkey, rows, cols})
	if err != nil {
		c.Errorf("gridTemplate: %v", err)
	}
}

var gridTmpl = `
<html>
<head>
<title>d3 testing</title>
<style>

.background {
  fill: #eee;
}

line {
  stroke: #fff;
}

text.active {
  fill: red;
}

line.active {
  stroke: #ccc;
}

</style>
<script src="http://d3js.org/d3.v3.min.js"></script>

<script>
var rows = [];
{{range .Rows}}
rows.push("{{.Label}}");
{{end}}
var columns = [];
{{range .Cols}}
columns.push("{{.Label}}");
{{end}}

var gridKey = "{{.GridKey}}";
</script>
</head>

<body>
<h2>{{.Grid.Title}}</h2>
<script src="/media/js/griddo.js"></script>
</body>
</html>

`

var indexTmpl = `<html>
	<head>
		<title>Griddo</title>
	</head>
	<body>
		<form action="/new/" method="post">
			<input type="text" name="title" placeholder="title" /><br />
			<textarea name="rows" rows="5" cols="50" placeholder="row labels. one per line"></textarea><br />
			<textarea name="cols" rows="5" cols="50" placeholder="columns labels. one per line"></textarea><br />
			<input type="submit" value="new grid" />
		</form>
	</body>
</html>
`
