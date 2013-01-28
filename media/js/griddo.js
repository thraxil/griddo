var margin = {top: 80, right: 20, bottom: 10, left: 120},
    width = 720,
    height = 720;

var x = d3.scale.ordinal().rangeBands([0, height]),
    y = d3.scale.ordinal().rangeBands([0, width]),
    z = d3.scale.linear().domain([0, 1]).clamp(true),
    c = d3.scale.category10().domain(d3.range(10));

var svg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .style("margin-left", margin.left + "px")
  .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top +
    ")");

var matrix = [];
//rows = [],
//columns = [],
var nr = 0;
var nc = 0;

var addCell = function(link) {
    matrix[link.row][link.col].z += link.value;
};

var setCell = function(link) {
		if (link.row > nr || link.col > nc) {
				console.log("out of bounds!");
				return
		}
    matrix[link.row][link.col].z += link.value;
    var cell = d3.selectAll(".cell")
				.transition()
				.duration(500)
				.style("fill-opacity", function(d) { return z(d.z); })
				.style("fill", function(d) { return c(d.z); });
}

function rowCreate(row) {
    var cell = d3.select(this).selectAll(".cell")
				.data(row);
    cell.enter().append("rect")
				.attr("class", "cell")
				.attr("x", function(d) { return y(d.x); })
				.attr("width", y.rangeBand())
				.attr("height", x.rangeBand())
				.on("mouseover", mouseover)
				.on("click", function(d, i) {
						d.z += 1;
						d.z %= 5;
						cell.style("fill-opacity", function(d) { return z(d.z); });
						cell.style("fill", function(d) { return c(d.z); });
						var xmlHttp = new XMLHttpRequest();
						var url = "/cellupdate/" + gridKey + "/" + d.y + "/" + d.x + "/";
						xmlHttp.open("POST", url, true);
						xmlHttp.setRequestHeader(
								"Content-type",
								"application/x-www-form-urlencoded");
						xmlHttp.send("v=" + d.z);
						console.log("v=" + d.z);
				})
				.on("mouseout", mouseout);
    cell
				.style("fill-opacity", function(d) { return z(d.z); })
				.style("fill", function(d) { return c(d.z); });
}

function mouseover(p) {
    d3.selectAll(".row text")
				.classed("active", function(d, i) {
						return i == p.y; });
    d3.selectAll(".row line")
				.classed("active", function(d, i) {
						return i == p.y || i == p.y + 1; });
    d3.selectAll(".column text")
				.classed("active", function(d, i) {
						return i == p.x; });
    d3.selectAll(".column line")
				.classed("active", function(d, i) {
						return i == p.x || i == p.x + 1; });
}

function mouseout() {
    d3.selectAll("text").classed("active", false);
    d3.selectAll("line").classed("active", false);
}

function dataUpdate() {
		var row = svg.selectAll(".row")
				.data(matrix);

		var g = row.enter().append("g");
		g
				.attr("class", "row")
				.attr("transform", function(d, i) { return "translate(0," + x(i) + ")"; })
				.each(rowCreate);

		g.append("line")
				.attr("x2", width);

		g.append("text")
				.attr("x", -6)
				.attr("y", x.rangeBand() / 2)
				.attr("dy", ".32em")
				.attr("text-anchor", "end")
				.text(function(d, i) { return rows[i]; });

}

function update() {
		nr = rows.length;
		nc = columns.length;

		// Compute index per node.
		rows.forEach(function(node, i) {
				matrix[i] = d3.range(nc).map(function(j) { return {x: j, y: i, z: 0}; });
		});

		x.domain(d3.range(nr));
		y.domain(d3.range(nc));

		svg.append("rect")
				.attr("class", "background")
				.attr("width", width)
				.attr("height", height);

		dataUpdate();

		var column = svg.selectAll(".column")
				.data(function (d, i) {
						return matrix[i];
				})
				.enter().append("g")
				.attr("class", "column")
				.attr("transform", function(d, i) { return "translate(" + y(i) + ")rotate(-90)"; });

		column.append("line")
				.attr("x1", -width);

		column.append("text")
				.attr("x", 6)
				.attr("y", y.rangeBand() / 2)
				.attr("dy", ".32em")
				.attr("text-anchor", "start")
				.attr("transform", "rotate(25)")
				.text(function(d, i) { return columns[i]; });

//		griddata.links.forEach(setCell);
		cells.forEach(setCell);
};


//d3.json("/test/test.json", update);

update();

//setInterval(function () {
//		var row = parseInt(nr * Math.random(), 10);
//		var col = parseInt(nc * Math.random(), 10);
//		var v = parseInt(5 * Math.random());
//		setCell({"row": row, "col": col, "value": v});
//}, 3000);
