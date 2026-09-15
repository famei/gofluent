package acrylic

// vertexShaderSource passes the fixed-function vertices and texture coordinates
// through to the fragment stage. QGLWidget provides a compatibility (legacy)
// context, so GLSL 1.20 built-ins such as ftransform() and gl_MultiTexCoord0
// are available.
const vertexShaderSource = `#version 120

void main() {
    gl_Position = ftransform();
    gl_TexCoord[0] = gl_MultiTexCoord0;
}
`

// fragmentShaderSource performs:
//   - a rounded-rectangle mask (pixels outside the radius are discarded),
//   - a disk-shaped Gaussian blur of the backdrop texture,
//   - an optional tint, grain and border.
const fragmentShaderSource = `#version 120

uniform sampler2D uTex;      // downsampled backdrop, used for the blur
uniform sampler2D uSharp;    // full-resolution backdrop, used outside the corners
uniform vec2      uTexel;    // (1/texWidth, 1/texHeight)
uniform vec4      uTint;     // rgb = tint colour, a = tint strength [0..1]
uniform int       uRadius;   // blur kernel radius 0 .. MAX_R
uniform float     uSigma;    // blur fall-off
uniform vec2      uSize;     // widget size in pixels
uniform float     uCorner;   // corner radius in pixels (0 = square)
uniform vec4      uBorder;   // border colour (rgb + a)
uniform float     uBorderW;  // border width in pixels (0 = none)
uniform float     uNoise;    // grain amount 0..255

const int MAX_R = 8;

// Signed distance to a rounded box centred at the origin.
float roundedBox(vec2 p, vec2 b, float r) {
    vec2 q = abs(p) - b + r;
    return min(max(q.x, q.y), 0.0) + length(max(q, 0.0)) - r;
}

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453);
}

void main() {
    vec2 uv = gl_TexCoord[0].st;

    // Always compute the signed distance: with uCorner == 0 this is a plain box
    // SDF, so square cards still get the acrylic fill and a square border.
    vec2 halfSize = uSize * 0.5;
    float d = roundedBox(uv * uSize - halfSize, halfSize, uCorner);

    vec3  acc  = vec3(0.0);
    float wsum = 0.0;
    float r2   = float(uRadius * uRadius);

    for (int i = -MAX_R; i <= MAX_R; ++i) {
        for (int j = -MAX_R; j <= MAX_R; ++j) {
            float dd = float(i * i + j * j);
            if (dd > r2) {
                continue;
            }
            vec2  o = vec2(float(i), float(j));
            float w = exp(-dd / (2.0 * uSigma * uSigma));
            acc  += texture2D(uTex, uv + o * uTexel).rgb * w;
            wsum += w;
        }
    }

    // wsum is always >= 1 because the centre sample (i=0, j=0) has weight 1.
    vec3 blurred = acc / wsum;
    vec3 colour  = mix(blurred, uTint.rgb, uTint.a);

    if (uNoise > 0.0) {
        float n = (hash(uv * uSize) - 0.5) * 2.0 * (uNoise / 255.0);
        colour += vec3(n);
    }

    if (uBorderW > 0.0 && d > -uBorderW) {
        colour = mix(colour, uBorder.rgb, uBorder.a);
    }

    // Outside the rounded rectangle show the SHARP, untinted backdrop (the
    // content actually behind the card). The 1px anti-aliasing band sits just
    // INSIDE the shape, so the acrylic never bleeds into the corner: for any
    // d >= 0 the output is exactly the backdrop.
    float cover = clamp(-d, 0.0, 1.0);
    vec3  sharp = texture2D(uSharp, uv).rgb;
    gl_FragColor = vec4(mix(sharp, colour, cover), 1.0);
}
`
