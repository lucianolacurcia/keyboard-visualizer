import './style.css';

// Load the SVG inline so we can manipulate its DOM
fetch('/src/assets/totem.svg')
    .then(r => r.text())
    .then(svgText => {
        const container = document.getElementById('app');
        container.innerHTML = svgText;

        const svg = container.querySelector('svg');

        // Show only the first layer (BASE), hide others
        // Each layer starts with a <text class="label"> and is followed by key groups
        // We'll find all layer sections and manage visibility
        setupLayers(svg);

        // Demo: simulate random keypresses every 500ms
        setInterval(() => {
            const pos = Math.floor(Math.random() * 38);
            highlightKey(svg, pos);
            setTimeout(() => unhighlightKey(svg, pos), 300);
        }, 500);
    });

function setupLayers(svg) {
    // Find all layer labels to identify layer boundaries
    const labels = svg.querySelectorAll('text.label');
    labels.forEach((label, i) => {
        if (i > 0) {
            // Hide non-BASE layers by finding their parent group or siblings
            // For now we'll handle this when we wire real data
        }
    });
}

function highlightKey(svg, position) {
    const key = svg.querySelector(`.keypos-${position}`);
    if (key) {
        const rect = key.querySelector('rect');
        if (rect) rect.classList.add('pressed');
    }
}

function unhighlightKey(svg, position) {
    const key = svg.querySelector(`.keypos-${position}`);
    if (key) {
        const rect = key.querySelector('rect');
        if (rect) rect.classList.remove('pressed');
    }
}
