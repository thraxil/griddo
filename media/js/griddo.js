var margin = {top: 80, right: 0, bottom: 10, left: 120},
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

d3.json("test.json", function(griddata) {
	var matrix = [],
		rows = griddata.rows,
		columns = griddata.columns,
		nr = rows.length,
		nc = columns.length;

  // Compute index per node.
  rows.forEach(function(node, i) {
    matrix[i] = d3.range(nc).map(function(j) { return {x: j, y: i, z: 0}; });
  });

  griddata.links.forEach(function(link) {
    matrix[link.source][link.target].z += link.value;
  });

  x.domain(d3.range(nr));
  y.domain(d3.range(nc));

  svg.append("rect")
      .attr("class", "background")
      .attr("width", width)
      .attr("height", height);

  var row = svg.selectAll(".row")
      .data(matrix)
    .enter().append("g")
      .attr("class", "row")
      .attr("transform", function(d, i) { return "translate(0," + x(i) + ")"; })
      .each(rowCreate);

  row.append("line")
      .attr("x2", width);

  row.append("text")
      .attr("x", -6)
      .attr("y", x.rangeBand() / 2)
      .attr("dy", ".32em")
      .attr("text-anchor", "end")
      .text(function(d, i) { return rows[i].label; });

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
      .text(function(d, i) { return columns[i].label; });

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
			})
      .on("mouseout", mouseout);
     cell
      .style("fill-opacity", function(d) { return z(d.z); })
      .style("fill", function(d) { return c(d.z); });
  }

  function mouseover(p) {
    d3.selectAll(".row text").classed("active", function(d, i) { return i == p.y; });
    d3.selectAll(".row line").classed("active", function(d, i) { return i == p.y || i == p.y + 1; });
    d3.selectAll(".column text").classed("active", function(d, i) { return i == p.x; });
    d3.selectAll(".column line").classed("active", function(d, i) { return i == p.x || i == p.x + 1; });
  }

  function mouseout() {
    d3.selectAll("text").classed("active", false);
  }

});
