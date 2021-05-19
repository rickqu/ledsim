let width = 1000;
let height = 600;
let paused = false;

let canvas = d3.select('body').append('svg')
	.attr('width', width)
	.attr('height', height)
    .style('background-color', 'rgba(0, 0, 0, 1.0)')
	.append('g');


// canvas.append('circle')
// 	.attr('r', 2)
// 	.style('fill', 'rgba(0, 0, 0, 1.0)')
// 	.attr('transform', 'translate(' + width / 2 + ', ' + height / 2 + ')');

// for (let i = 1; i < 15; i++) {
// 	canvas.append('circle')
// 		.attr('r', 70 * i)
// 		.style('fill', 'rgba(0, 0, 0, 0)')
// 		.style('border', 'rgba(0, 0, 0, 1.0)')
// 		.style('fill', 'rgba(0, 0, 0, 0)')
// 		.style('stroke', 'rgba(0, 0, 0, 0.2)')
// 		.attr('transform', 'translate(' + width / 2 + ', ' + height / 2 + ')');
// }

canvas.append('rect')
	.attr('x', 1)
	.attr('y', 1)
	.attr('width', width - 2)
	.attr('height', height - 2)
	.style('border', 'rgba(255, 255, 255, 1.0)')
	.style('fill', 'rgba(255, 255, 255, 0)')
	.style('stroke', 'rgba(255, 255, 255, 1.0)')
	.style('stroke-width', '1');

const socket = new WebSocket(window.location.origin.replace('http', 'ws') + '/ws');

let mapping = [];

function draw(data) {
	if (mapping.length > 0) {
		for (let i = 0; i < mapping.length; i++) {
			let led = data.leds[i];
			mapping[i].style('fill', 'rgb(' + led.R + ',' + led.G + ',' + led.B + ')');
		}
		return;
	}

	canvas.selectAll('.data-point').remove();

	for (let led of data.leds) {
        let rendered = canvas.append('circle')
            .attr('class', 'data-point')
            .attr('r', 3)
            .style('fill', 'rgb(' + led.R + ',' + led.G + ',' + led.B + ')')
            .attr('transform', 'translate(' + (led.X * 120 + (width/2)) + ', ' + (-led.Y * 120 + (height/2)) + ')');
		mapping.push(rendered);
	}
}


socket.addEventListener('message', function (event) {
	let leds = event.data.split(',').map(v => {
		return {
			R: v >> 16,
			G: (v >> 8) & 0xff,
			B: v & 0xff
		}
	});

	let i = 0;
	for (let y = -73.0; y < 80.0; y += 3.88) {
		for (let x = -100.0; x < 120.0; x += 4.4) {
			leds[i].X = x / 40
			leds[i].Y = y / 40
			i++;
		}
	}

	// var data = JSON.parse(event.data);
	draw({
		leds: leds
	});
});
