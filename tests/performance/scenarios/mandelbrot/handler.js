module.exports = {
  main: async function (event, context) {
    const data = event.extensions.request.headers;
    width = data['width'];
    height = data['height'];
    iterations = data['iterations'];
    zoom = data['zoom'];

    console.log(`Received request with width=${width}, height=${height}, iterations=${iterations}, zoom=${zoom}`);

    return generateMandelbrotJson(width, height, iterations, zoom);
  }
}

function generateMandelbrotJson(width, height, iterations, zoom) {
  const safeWidth = Number.parseInt(width);
  const safeHeight = Number.parseInt(height);
  const maxIterations = Number.parseInt(iterations);
  const safeZoom = Number.parseFloat(zoom);
  const values = [];

  const xOffset = -0.7;
  const yOffset = 0;

  for (let y = 0; y < safeHeight; y++) {
    const row = [];
    const ci = (y / safeHeight - 0.5) / safeZoom + yOffset;

    for (let x = 0; x < safeWidth; x++) {
      const cr = (x / safeWidth - 0.5) / safeZoom + xOffset;
      row.push(mandelbrotEscapeIterations(cr, ci, maxIterations));
    }

    values.push(row);
  }

  return {
    type: "mandelbrot",
    width: safeWidth,
    height: safeHeight,
    maxIterations: maxIterations,
    zoom: safeZoom,
    values: values
  };
}

function mandelbrotEscapeIterations(cr, ci, maxIterations) {
  let zr = 0;
  let zi = 0;
  let iter = 0;

  while (iter < maxIterations && zr * zr + zi * zi < 4) {
    const nextZr = zr * zr - zi * zi + cr;
    zi = 2 * zr * zi + ci;
    zr = nextZr;
    iter++;
  }

  return iter;
}
