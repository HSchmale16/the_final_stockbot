import * as d3 from "https://cdn.jsdelivr.net/npm/d3@7/+esm";


window.onload = function () {
    fetch("/json/congress-network")
        .then(response => response.json())
        .then(data => {
            drawNetwork(data);
        });
}


function drawNetwork(data) {
    // Use d3 to render the nodes

    const width = 1200;
    const height = 800;

    const links = data.edges.map(d => ({...d}));
    const nodes = data.nodes.map(d => ({...d}));

    console.log(links);
    console.log(nodes)

    const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(links).id(d => d.BioGuideId))
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2))
        .on("tick", ticked);

    const svg = d3.create("svg")
        .attr("width", width)
        .attr("height", height)
        .attr("viewBox", [0, 0, width, height])
        .attr("style", "max-width: 100%; height: auto;");

    const link = svg.append("g")
        .attr("stroke", "#999")
        .attr("stroke-opacity", 0.6)
        .selectAll()
        .data(links)
        .join("line")
        .attr("stroke-width", d => Math.sqrt(d.value));

    const node = svg.append("g")
        .attr("stroke", "#fff")
        .attr("stroke-width", 1.5)
        .selectAll()
        .data(nodes)
        .join("circle")
        .attr("r", 5)
        .attr("fill", d => d.Party === "R" ? "red" : "blue");
    

    // When I hover over a node, show the name
    node.append("title")
        .text(d => `${d.Name} (${d.State} - ${d.Party})`);
    
    
    // Put my svg in #container
    document.getElementById("container").appendChild(svg.node());

    function ticked() {
        link.attr("x1", d => d.source.x)
            .attr("y1", d => d.source.y)
            .attr("x2", d => d.target.x)
            .attr("y2", d => d.target.y);

        node.attr("cx", d => d.x)
            .attr("cy", d => d.y);
    }
}