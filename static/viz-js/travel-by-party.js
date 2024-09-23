var margin = { top: 10, right: 30, bottom: 20, left: 50 },
    width = 430 - margin.left - margin.right,
    height = 400 - margin.top - margin.bottom;




fetch("/json/travel-by-party")
    .then(response => response.json())
    .then(data => {
        var svg = makeSvg("#travel-by-party");
        makeTravelPartyBarChart(svg, data, "Number of Trips by Party");
     });

fetch("/json/days-traveled-by-party")
    .then(response => response.json())
    .then(data => {
        var svg = makeSvg("#days-traveled-by-party");
        makeTravelPartyBarChart(svg, data, "Days Traveled by Party");
     });


function makeSvg(id) {
    var svg = d3.select(id)
    .append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform",
        "translate(" + margin.left + "," + margin.top + ")");
    return svg;
}


function makeTravelPartyBarChart(svg, data, title) {
    // data is a list of objects, each with a year, party and a count
    // we are rendering a grouped bar chart using d3. Year is the x-axis, count is the y-axis, and party is the grouping

    // Add chart title
    svg.append("text")
        .attr("x", (width / 2))
        .attr("y", 0 - (margin.top / 2 - 15))
        .attr("text-anchor", "middle")
        .style("font-size", "16px")
        .style("text-decoration", "underline")
        .text(title);


    var subgroups = ['Republican', 'Democrat', 'Independent', 'Libertarian'];
    var groups = d3.map(data, (d) => d.year).keys();

    var x = d3.scaleBand()
        .domain(data.map(d => d.year))
        .range([0, width])
        .padding(0.2);


    svg.append("g")
        .attr("transform", `translate(0,${height})`)
        .call(d3.axisBottom(x).tickSize(0));


    var y = d3.scaleLinear()
        .domain([0, d3.max(data, d => d3.max(subgroups, key => d[key]))]).nice()
        .range([height, 0]);

    console.log(y);
    svg.append("g")
        .call(d3.axisLeft(y).ticks(null, "s"))

    var xSubgroup = d3.scaleBand()
        .domain(subgroups)
        .range([0, x.bandwidth()])
        .padding(0.05);

    var color = d3.scaleOrdinal()
        .domain(subgroups)
        .range(['#e41a1c', '#377eb8', '#4daf4a', '#984ea3']);

    var tooltip = d3.select("body").append("div")
        .attr("class", "tooltip")
        .style("position", "absolute")
        .style("background-color", "white")
        .style("border", "solid")
        .style("border-width", "1px")
        .style("border-radius", "5px")
        .style("padding", "10px")
        .style("display", "none");

    // Show the bars
    svg.append("g")
        .selectAll("g")
        // Enter in data = loop group per group
        .data(data)
        .enter()
        .append("g")
        .attr("transform", function (d) { return "translate(" + x(d.year) + ",0)"; })
        .selectAll("rect")
        .data(function (d) { return subgroups.map(function (key) { return { key: key, value: d[key] || 0 }; }); })
        .enter().append("rect")
        .attr("x", function (d) { return xSubgroup(d.key); })
        .attr("y", function (d) { return y(d.value); })
        .attr("width", xSubgroup.bandwidth())
        .attr("height", function (d) {
            return height - y(d.value);
        })
        .attr("fill", function (d) { return color(d.key); })
        .on("mouseover", function (event, d) {
            d3.select(this).attr("fill", d3.rgb(color(d.key)).darker(1));
            tooltip.style("display", "block")
                .html(`Party: ${d.key}<br>Count: ${d.value}`)
                .style("left", (event.pageX + 10) + "px")
                .style("top", (event.pageY - 10) + "px");
        })
        .on("mouseout", function (event, d) {
            d3.select(this).attr("fill", color(d.key));
            tooltip.style("display", "none");
        })
        .on("mousemove", function (event) {
            tooltip.style("left", (event.pageX + 10) + "px")
                .style("top", (event.pageY - 10) + "px");
        });
}
