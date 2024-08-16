let chamber = "S"
let tagId = "";
let url = "/json/congress-network?" + new URLSearchParams({ chamber: chamber, tag_id: tagId });

console.log("FUNCTION MAIN")
var diameter = 750,
    radius = diameter / 2,
    innerRadius = radius - 120;

var cluster = d3.cluster()
    .size([360, innerRadius]);

var line = d3.radialLine()
    .curve(d3.curveBundle.beta(0.85))
    .radius(function (d) { return d.y; })
    .angle(function (d) { return d.x / 180 * Math.PI; });

var svg = d3.select("#network").append("svg")
    .attr("width", diameter)
    .attr("height", diameter)
    .append("g")
    .attr("transform", "translate(" + radius + "," + radius + ")");

var link = svg.append("g").selectAll(".link");
var node = svg.append("g").selectAll(".node");

console.log("DOING JSON")
d3.json(url).then(data => {
    var root = packageHierarchy(preprocessData(data))
        .sum(function (d) { return d.count; });

    console.log(root)
    
    cluster(root)
    leaves = root.leaves()

    imports = packageImports(leaves, data.edges)
    console.log(imports)

    link = link
        .data(imports)
        .enter().append("path")
        .each(function (d) { console.log(d); d.source = d[0], d.target = d[d.length - 1]; })
        .attr("class", "link")
        .attr("d", line);

    node = node
        .data(root.leaves())
        .enter().append("text")
        .attr("class", "node")
        .attr("dy", "0.31em")
        .attr("transform", function (d) { return "rotate(" + (d.x - 90) + ")translate(" + (d.y + 8) + ",0)" + (d.x < 180 ? "" : "rotate(180)"); })
        .attr("text-anchor", function (d) { return d.x < 180 ? "start" : "end"; })
        .text(function (d) { return d.data.RenderName; })

    
})

function packageImports(nodes, edges) {
    var map = {},
        imports = [];

    // Compute a map from name to node.
    nodes.forEach(function (d) {
        map[d.data.name] = d;
    });
    console.log('m', map)

    // For each import, construct a link from the source to target node.
    edges.forEach(function (d) {
        imports.push(
            map[d.source].path(map[d.target])
        )
    });

    return imports;
}


function preprocessData(data) {
    var nodes = data.nodes;
    var nodeMap = {};
    nodes.forEach(function (x) { nodeMap[x.BioGuideId] = {
        name: x.BioGuideId,
        ...x
    }; });
    console.log(nodeMap);
    return Object.values(nodeMap);
}

// function packageHierarchy(classes) {
//     console.log(classes)
//     var map = {};

//     function find(name, data) {
//         var node = map[name], i;
//         if (!node) {
//             node = map[name] = data || { name: name, children: [] };
//             if (name.length) {
//                 node.parent = find(name.substring(0, i = name.lastIndexOf(".")));
//                 node.parent.children.push(node);
//                 node.key = name.substring(i + 1);
//             }
//         }
//         return node;
//     }

//     classes.forEach(function (d) {
//         find(d.name, d);
//     });

//     return d3.hierarchy(map[""]);
// }

function packageHierarchy(senators) {
    var map = {};

    senators.forEach(function (d) {
        state = map[d.State] || { name: d.State, children: [] };
        map[d.State] = state;
        state.children.push({
            name: d.RenderName,
            count: d.Count,
            children: [],
            ...d
        });
    });

    console.log(map)

    return d3.hierarchy({
        name: "",
        children: Object.values(map)
    });
}