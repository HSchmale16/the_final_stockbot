function fetchDataForChamber(chamber, tagId = "") {
    // Clear it before generating new element
    document.getElementById("container").innerHTML = "";

    // load the data
    fetch("/json/congress-network?" + new URLSearchParams({ chamber: chamber, tag_id: tagId }))
        .then(response => response.json())
        .then(data => {
            drawNetwork(data);
        });
}

function updateBioguideTooltipWindow(bioGuideId, tooltip) {
    const url = getCongressPersonDetailsUrl(bioGuideId);
    tooltip.html("Loading...");

    fetch(url)
        .then(response => response.text())
        .then(html => {
            tooltip.html(html);
        });
}



function getCongressPersonDetailsUrl(bioGuideId) {
    return `/congress-member/${bioGuideId}/embed`;
}

function NodeSizeHandler(d) {
    // return 1.5 * d.Count;
    return d.R;
}

// create a continious color scale



function restoreNodeAppearance(nodes) {
    nodes.attr("r", NodeSizeHandler)
        // .attr("stroke", d => colorScale(d.State))
        .attr("stroke", "#FFF")
        .attr("stroke-width", 1)
        .attr("fill", d => PartyColor(d.Party));
}

function drawNetwork(data) {
    // Use d3 to render the nodes

    const width = 1000;
    const height = 800;

    const links = data.edges.map(d => ({ ...d }));
    const nodes = data.nodes.map(d => ({ ...d }));


    const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(links).id(d => d.BioGuideId).distance(0).strength(0))
        .force("charge", d3.forceManyBody().strength(-5))
        // .force("x", d3.forceX())
        // .force("y", d3.forceY())
        .force("x", d3.forceX().strength(0.01))
        .force("y", d3.forceY().strength(0.01))
        .force("cluster", forceCluster())
        .force("collide", forceCollide())
        .on("tick", ticked);



    const svg = d3.create("svg")
        .attr("width", width)
        .attr("height", height)
        .attr("viewBox", [-width / 2, -height / 2, width, height])
        .attr("style", "max-width: 100%; height: auto;");

    // We're going to hide thie links by default.
    // const link = svg.append("g")
    //     .attr("stroke", "#999")
    //     .attr("stroke-opacity", 0.6)
    //     .selectAll()
    //     .data(links)
    //     .join("line")
    //     .attr("stroke-width", d => d.value / 5);

    const node = svg.append("g")
        .selectAll()
        .data(nodes)
        .join("circle")

    restoreNodeAppearance(node);


    // JavaScript: Enhance node hover effect and implement tooltips
    const tooltip = d3.select("#tooltip")
        .style("opacity", 0);

    let clicked = false;

    node.on("mouseover", (event, d) => {
        if (!clicked) {
            // Enhance node appearance
            d3.select(event.currentTarget)
                .attr("r", d => 2 * NodeSizeHandler(d)) // Increase radius
                .attr("stroke", "gold");

            updateBioguideTooltipWindow(d.BioGuideId, tooltip);

            // Show tooltip
            tooltip.transition()
                .duration(400)
                .style("opacity", .9);

        }
    })
        .on("mouseout", (event, d) => {
            if (!clicked) {
                // Reset node appearance
                let node = d3.select(event.currentTarget)

                restoreNodeAppearance(node);

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
                    .attr("r", d => 2 * NodeSizeHandler(d)) // Increase radius
                    .attr("fill", "gold"); // Change color
            } else {
                restoreNodeAppearance(d3.select(event.currentTarget));

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
            d3.selectAll('circle').filter(d => d.RenderName.toLowerCase().includes(searchTerm))
                .attr('r', 10) // Increase size
                .attr('fill', 'green'); // Change color
        }
    });

    // Put my svg in #container
    document.getElementById("container").appendChild(svg.node());

    function ticked() {
        // link.attr("x1", d => d.source.x)
        //     .attr("y1", d => d.source.y)
        //     .attr("x2", d => d.target.x)
        //     .attr("y2", d => d.target.y);

        node.attr("cx", d => d.x)
            .attr("cy", d => d.y);
    }
}


// https://observablehq.com/@d3/clustered-bubbles
// forceCluster and centroid are from the above link
function forceCluster() {
    const strength = 0.2;
    let nodes;

    function force(alpha) {
        const centroids = d3.rollup(nodes, centroid, d => d.Group);
        const l = alpha * strength;
        for (const d of nodes) {
            const { x: cx, y: cy } = centroids.get(d.Group);
            d.vx -= (d.x - cx) * l;
            d.vy -= (d.y - cy) * l;
        }
    }

    force.initialize = _ => nodes = _;

    return force;
}

function centroid(nodes) {
    let x = 0;
    let y = 0;
    let z = 0;
    for (const d of nodes) {
        let k = NodeSizeHandler(d) ** 2;
        x += d.x * k;
        y += d.y * k;
        z += k;
    }
    return { x: x / z, y: y / z };
}

function forceCollide() {
    const alpha = 0.4; // fixed for greater rigidity!
    const padding1 = 2; // separation between same-color nodes
    const padding2 = 6; // separation between different-color nodes
    let nodes;
    let maxRadius;
  
    function force() {
      const quadtree = d3.quadtree(nodes, d => d.x, d => d.y);
      for (const d of nodes) {
        const r = d.r + maxRadius;
        const nx1 = d.x - r, ny1 = d.y - r;
        const nx2 = d.x + r, ny2 = d.y + r;
        quadtree.visit((q, x1, y1, x2, y2) => {
          if (!q.length) do {
            if (q.data !== d) {
              const r = d.r + q.R + (d.Group === q.data.Group? padding1 : padding2);
              let x = d.x - q.data.x, y = d.y - q.data.y, l = Math.hypot(x, y);
              if (l < r) {
                l = (l - r) / l * alpha;
                d.x -= x *= l, d.y -= y *= l;
                q.data.x += x, q.data.y += y;
              }
            }
          } while (q = q.next);
          return x1 > nx2 || x2 < nx1 || y1 > ny2 || y2 < ny1;
        });
      }
    }
  
    force.initialize = _ => maxRadius = d3.max(nodes = _, d => d.r) + Math.max(padding1, padding2);
  
    return force;
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