import * as d3 from "https://cdn.jsdelivr.net/npm/d3@7/+esm";


function fetchDataForChamber(chamber) {
    // Clear it before generating new element
    document.getElementById("container").innerHTML = "";

    // load the data
    fetch("/json/congress-network?" + new URLSearchParams({ chamber: chamber }))
        .then(response => response.json())
        .then(data => {
            drawNetwork(data);
        });
}

function getCongressPersonDetailsUrl(bioGuideId) {
    return `/congress-member/${bioGuideId}/embed`;
}

function NodeSizeHandler(d) {
    return 1.5 * d.Count;
}

function drawNetwork(data) {
    // Use d3 to render the nodes

    const width = 1000;
    const height = 800;

    const links = data.edges.map(d => ({ ...d }));
    const nodes = data.nodes.map(d => ({ ...d }));


    const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(links).id(d => d.BioGuideId).distance(80))
        .force("charge", d3.forceManyBody())
        .force("x", d3.forceX())
        .force("y", d3.forceY())
        .on("tick", ticked);

    const svg = d3.create("svg")
        .attr("width", width)
        .attr("height", height)
        .attr("viewBox", [-width / 2, -height / 2, width, height])
        .attr("style", "max-width: 100%; height: auto;");

    const link = svg.append("g")
        .attr("stroke", "#999")
        .attr("stroke-opacity", 0.6)
        .selectAll()
        .data(links)
        .join("line")
        .attr("stroke-width", d => d.value / 5);

    const node = svg.append("g")
        .attr("stroke", "#fff")
        .attr("stroke-width", 1.5)
        .selectAll()
        .data(nodes)
        .join("circle")
        .attr("r", d => d.Count * 1.2 + 2)
        .attr("fill", d => PartyColor(d.Party));



    // JavaScript: Enhance node hover effect and implement tooltips
    const tooltip = d3.select("#tooltip")
        .style("opacity", 0);

    let clicked = false;

    node.on("mouseover", (event, d) => {
        if (!clicked) {
            // Enhance node appearance
            d3.select(event.currentTarget)
                // .attr("r", 10) // Increase radius
                .attr("stroke", "gold"); // Change color

            // Show tooltip
            tooltip.transition()
                .duration(400)
                .style("opacity", .9);

            tooltip.html(`<div hx-trigger="revealed" hx-get="${getCongressPersonDetailsUrl(d.BioGuideId)}" >${d.Name}</div>`)


            htmx.process(document.getElementById("tooltip"));
        }
    })
        .on("mouseout", (event, d) => {
            if (!clicked) {
                // Reset node appearance
                d3.select(event.currentTarget)
                    .attr("r", NodeSizeHandler) // Reset radius
                    .attr("fill", d => PartyColor(d.Party)) // Reset color
                    .attr("stroke", "#fff"); // Reset color

                // Hide tooltip
                tooltip.transition()
                    .duration(500)
                    .style("opacity", 0);
            }
        })
        .on('click', (event, d) => {
            clicked = !clicked;
            if (clicked) {
                d3.select(event.currentTarget)
                    .attr("r", 10) // Increase radius
                    .attr("fill", "gold"); // Change color
            } else {
                d3.select(event.currentTarget)
                    .attr("r", 5) // Reset radius
                    .attr("fill", d => PartyColor(d.Party)); // Reset color
            }
        });

    node.call(d3.drag()
        .on("start", (event, d) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        })
        .on("drag", (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
        })
        .on("end", (event, d) => {
            if (!event.active) simulation.alphaTarget(0);
            // d.fx = null;
            // d.fy = null;
        }))


    const input = document.getElementById("search");
    input.addEventListener('keyup', (event) => {
        const searchTerm = event.target.value.toLowerCase();
        console.log(searchTerm)
        // Remove any previous highlights
        d3.selectAll('circle').attr('r', 5).attr('fill', d => PartyColor(d.Party));


        if (searchTerm !== "") {
            // Find and highlight the node that matches the search term
            d3.selectAll('circle').filter(d => d.Name.toLowerCase().includes(searchTerm))
                .attr('r', 10) // Increase size
                .attr('fill', 'green'); // Change color
        }
    });

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

function PartyColor(party) {
    switch (party[0]) {
        case 'R':
            return 'red';
        case 'D':
            return 'blue';
        default:
            return 'purple';
    }
}
window.fetchDataForChamber = fetchDataForChamber;