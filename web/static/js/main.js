// 拓扑图初始化
function initTopology(data) {
    const container = document.getElementById('topology-graph');
    if (!container) return;

    const nodes = new vis.DataSet(data.nodes);
    const edges = new vis.DataSet(data.edges);

    const network = new vis.Network(container, {
        nodes: nodes,
        edges: edges
    }, {
        nodes: {
            shape: 'dot',
            size: 30,
            font: {
                size: 14
            }
        },
        edges: {
            width: 2,
            smooth: {
                type: 'continuous'
            }
        },
        physics: {
            stabilization: false,
            barnesHut: {
                gravitationalConstant: -80000,
                springConstant: 0.001,
                springLength: 200
            }
        }
    });
}

// 自动刷新处理
document.addEventListener('htmx:afterSettle', function(evt) {
    if (evt.detail.target.id === 'topology-container') {
        initTopology(JSON.parse(evt.detail.target.dataset.topology));
    }
}); 