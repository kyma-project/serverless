import json

def main(event, context):
    data = event['extensions']['request'].headers
    width = data.get("width")
    height = data.get("height")
    iterations = data.get("iterations")
    zoom = data.get("zoom")

    print(f"Received request with width={width}, height={height}, iterations={iterations}, zoom={zoom}")

    return generate_mandelbrot_json(width, height, iterations, zoom)


def generate_mandelbrot_json(width, height, iterations, zoom):
    safe_width = int(str(width).strip())
    safe_height = int(str(height).strip())
    max_iterations = int(str(iterations).strip())
    safe_zoom = float(str(zoom).strip())
    values = []

    x_offset = -0.7
    y_offset = 0.0

    for y in range(safe_height):
        row = []
        ci = (y / safe_height - 0.5) / safe_zoom + y_offset

        for x in range(safe_width):
            cr = (x / safe_width - 0.5) / safe_zoom + x_offset
            row.append(mandelbrot_escape_iterations(cr, ci, max_iterations))

        values.append(row)

    return {
        "type": "mandelbrot",
        "width": safe_width,
        "height": safe_height,
        "maxIterations": max_iterations,
        "zoom": safe_zoom,
        "values": values,
    }


def mandelbrot_escape_iterations(cr, ci, max_iterations):
    zr = 0.0
    zi = 0.0
    it = 0

    while it < max_iterations and (zr * zr + zi * zi) < 4:
        next_zr = zr * zr - zi * zi + cr
        zi = 2 * zr * zi + ci
        zr = next_zr
        it += 1

    return it
