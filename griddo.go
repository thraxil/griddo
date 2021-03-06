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
	Id string
	//	Created  time.Time
	//	Modified time.Time
}

type Row struct {
	Grid         *datastore.Key
	Label        string
	DisplayOrder int
}

func (r Row) Index() int {
	return r.DisplayOrder
}

type Col struct {
	Grid         *datastore.Key
	Label        string
	DisplayOrder int
}

func (c Col) Index() int {
	return c.DisplayOrder
}

type Cell struct {
	Value int
	Grid  *datastore.Key
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
	http.HandleFunc("/add_row/", addRow)
	http.HandleFunc("/add_col/", addCol)
}

func addRow(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	gridkey := parts[2]
	k := datastore.NewKey(ctx, "grid", gridkey, 0, nil)
	rq := datastore.NewQuery("row").Filter("Grid=", k).Order("DisplayOrder")
	rows := make([]Row, 0, 100)
	_, err := rq.GetAll(ctx, &rows)
	if err != nil {
		// handle the error
		ctx.Errorf("rows fetch: %v", err)
	}
	var maxOrder = 0
	for _, row := range rows {
		if row.DisplayOrder > maxOrder {
			maxOrder = row.DisplayOrder
		}
	}

	rkey := datastore.NewKey(ctx, "row", newKey(), 0, nil)
	row := new(Row)
	row.Grid = k
	row.Label = strings.TrimSpace(r.FormValue("label"))
	row.DisplayOrder = maxOrder + 1
	_, err = datastore.Put(ctx, rkey, row)
	ctx.Errorf("added row %v", err)

	http.Redirect(w, r, "/grid/"+gridkey, http.StatusFound)
}

func addCol(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	gridkey := parts[2]
	k := datastore.NewKey(ctx, "grid", gridkey, 0, nil)

	cq := datastore.NewQuery("col").Filter("Grid=", k).Order("DisplayOrder")
	cols := make([]Col, 0, 100)
	_, err := cq.GetAll(ctx, &cols)
	if err != nil {
		// handle the error
		ctx.Errorf("cols fetch: %v", err)
	}
	var maxOrder = 0
	for _, col := range cols {
		if col.DisplayOrder > maxOrder {
			maxOrder = col.DisplayOrder
		}
	}

	ckey := datastore.NewKey(ctx, "col", newKey(), 0, nil)
	col := new(Col)
	col.Grid = k
	col.Label = strings.TrimSpace(r.FormValue("label"))
	col.DisplayOrder = maxOrder + 1
	_, err = datastore.Put(ctx, ckey, col)
	ctx.Errorf("added col %v", err)

	http.Redirect(w, r, "/grid/"+gridkey, http.StatusFound)
}

func cellUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 5 {
		http.Error(w, "bad request", 404)
		return
	}
	gridkey := parts[2]
	k := datastore.NewKey(ctx, "grid", gridkey, 0, nil)
	//	g := new(Grid)

	ridx, _ := strconv.Atoi(parts[3])
	cidx, _ := strconv.Atoi(parts[4])

	v, _ := strconv.Atoi(r.FormValue("v"))
	ctx.Errorf("setting cell (%d,%d) to %d", ridx, cidx, v)

	rq := datastore.NewQuery(
		"row").Filter("Grid=",
		k).Filter("DisplayOrder=",
		ridx).Limit(1)

	rows := make([]Row, 0, 1)
	rkeys, err := rq.GetAll(ctx, &rows)
	if err != nil {
		ctx.Errorf("rows fetch: %v", err)
		fmt.Fprint(w, "not ok")
		return
	}

	cq := datastore.NewQuery(
		"col").Filter("Grid=",
		k).Filter("DisplayOrder=",
		cidx).Limit(1)

	cols := make([]Col, 0, 1)
	ckeys, err := cq.GetAll(ctx, &cols)
	if err != nil {
		// handle the error
		ctx.Errorf("cols fetch: %v", err)
		fmt.Fprint(w, "not ok")
		return
	}

	cellq := datastore.NewQuery(
		"cell").Filter("Grid=", k).Filter(
		"Row=", rkeys[0]).Filter("Col=", ckeys[0]).Limit(1)
	cells := make([]Cell, 0, 1)
	cellkeys, err := cellq.GetAll(ctx, &cells)
	if err != nil {
		ctx.Errorf("cells fetch: %v", err)
		fmt.Fprintf(w, "not ok")
		return
	}
	if len(cells) > 0 {
		if v != 0 {
			cells[0].Value = v
			_, err = datastore.Put(ctx, cellkeys[0], &cells[0])
			if err != nil {
				ctx.Errorf("error saving: %v", err)
			}
		} else {
			// value is zero, so delete it
			err = datastore.Delete(ctx, cellkeys[0])
			if err != nil {
				ctx.Errorf("error deleting: %v", err)
			}
		}
	} else {
		if v != 0 {
			ck := datastore.NewKey(ctx, "cell", newKey(), 0, nil)
			cell := new(Cell)
			cell.Grid = k
			cell.Row = rkeys[0]
			cell.Col = ckeys[0]
			cell.Value = v
			datastore.Put(ctx, ck, cell)
		}
		// value == 0, do nothing
	}

	fmt.Fprint(w, "ok")
}

var indexTemplate = template.Must(template.New("index").Parse(indexTmpl))

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

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
	k := datastore.NewKey(c, "grid", key, 0, nil)
	g := new(Grid)
	g.Id = key
	g.Title = title
	_, err := datastore.Put(c, k, g)
	if err != nil {
		c.Errorf("error adding grid: %v", err)
	}
	for i, r := range rows {
		rkey := datastore.NewKey(c, "row", newKey(), 0, nil)
		row := new(Row)
		row.Grid = k
		row.Label = strings.TrimSpace(r)
		row.DisplayOrder = i
		_, err := datastore.Put(c, rkey, row)
		c.Errorf("added row %v", err)
	}
	for i, co := range cols {
		ckey := datastore.NewKey(c, "col", newKey(), 0, nil)
		col := new(Col)
		col.Grid = k
		col.Label = strings.TrimSpace(co)
		col.DisplayOrder = i
		_, err := datastore.Put(c, ckey, col)
		c.Errorf("added col %v", err)
	}

	http.Redirect(w, r, "/grid/"+key, http.StatusFound)
}

type vcell struct {
	Cell Cell
	Row  Row
	Col  Col
	RN   int
	CN   int
}

type gridPage struct {
	Grid    *Grid
	GridKey string
	Rows    []Row
	Cols    []Col
	Cells   []vcell
}

var gridTemplate = template.Must(template.New("grid").Parse(gridTmpl))

func getGrid(ctx appengine.Context, key string) (*datastore.Key, *Grid, error) {
	gq := datastore.NewQuery("grid").Filter("Id=",key).Limit(1)
	grids := make([]Grid, 0, 1)
	gridkeys, err := gq.GetAll(ctx, &grids)
	ctx.Errorf("keys found: %v", gridkeys)
	ctx.Errorf("grids: %v", grids)
	if err == nil {
		return gridkeys[0], &grids[0], err
	}
	return nil, nil, err
}

func showGrid(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 404)
		return
	}
	gridkey := parts[2]
	k, g, err := getGrid(c, gridkey)

	if err != nil {
		http.Error(w, "Couldn't load Grid", http.StatusInternalServerError)
		c.Errorf("setting up: %v", err)
		c.Errorf("key: %v", parts)
		return
	}

	rowmap := make(map[string]Row)
	rq := datastore.NewQuery("row").Filter("Grid=", k).Order("DisplayOrder")
	rows := make([]Row, 0, 100)
	rkeys, err := rq.GetAll(c, &rows)
	if err != nil {
		// handle the error
		c.Errorf("rows fetch: %v", err)
	}
	for i, rk := range rkeys {
		rowmap[rk.String()] = rows[i]
	}

	colmap := make(map[string]Col)
	cq := datastore.NewQuery("col").Filter("Grid=", k).Order("DisplayOrder")
	cols := make([]Col, 0, 100)
	ckeys, err := cq.GetAll(c, &cols)
	if err != nil {
		// handle the error
		c.Errorf("cols fetch: %v", err)
	}
	for i, ck := range ckeys {
		colmap[ck.String()] = cols[i]
	}

	cellq := datastore.NewQuery("cell").Filter("Grid=", k).Limit(100 * 100)
	cells := make([]Cell, 0, 100*100)
	vcells := make([]vcell, 0, 100*100)

	if _, err := cellq.GetAll(c, &cells); err != nil {
		c.Errorf("cells fetch: %v", err)
	}

	for _, cell := range cells {
		var fr = rowmap[cell.Row.String()]
		var fc = colmap[cell.Col.String()]
		vcells = append(vcells, vcell{cell, fr, fc, fr.DisplayOrder, fc.DisplayOrder})
	}

	err = gridTemplate.Execute(w, gridPage{g, gridkey, rows, cols, vcells})
	if err != nil {
		c.Errorf("gridTemplate: %v", err)
	}
}

const indexTmpl = `<html>
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

const gridTmpl = `<html>
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
rows.push("{{.Label}}");{{end}}
var columns = [];
{{range .Cols}}
columns.push("{{.Label}}");{{end}}

var cells = [];
{{range .Cells}}
cells.push({"row": {{.Row.Index}}, "col": {{.Col.Index}}, "value": {{.Cell.Value}}});{{end}}

var gridKey = "{{.GridKey}}";
</script>
</head>

<body>
<h2>{{.Grid.Title}}</h2>
<script src="/media/js/griddo.js"></script>

	<form action="/add_row/{{.GridKey}}" method="post">
  <input type="text" name="label" placeholder="row label"/><input type="submit" value="add row" />
	</form>
	<form action="/add_col/{{.GridKey}}" method="post">
  <input type="text" name="label" placeholder="column label"/><input type="submit" value="add column" />
	</form>
</body>
</html>`
