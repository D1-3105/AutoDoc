#!/bin/bash
# build-swagger-ui.sh
# Usage: ./build-swagger-ui.sh path/to/openapi.json output.html

set -e

INPUT_JSON="$1"
OUTPUT_HTML="$2"

if [[ -z "$INPUT_JSON" || -z "$OUTPUT_HTML" ]]; then
  echo "Usage: $0 path/to/openapi.json output.html"
  exit 1
fi

if [[ ! -f "$INPUT_JSON" ]]; then
  echo "Error: File '$INPUT_JSON' does not exist"
  exit 1
fi

# Swagger UI CDN
CSS="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css"
BUNDLE_JS="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"
PRESET_JS="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-standalone-preset.js"


OPENAPI_CONTENT=$(jq -c . "$INPUT_JSON")

cat > "$OUTPUT_HTML" <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Docs</title>
  <link rel="stylesheet" href="$CSS">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="$BUNDLE_JS"></script>
  <script src="$PRESET_JS"></script>
  <script>
    window.onload = function() {
      const spec = $OPENAPI_CONTENT;
      SwaggerUIBundle({
        spec: spec,
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
      });
    }
  </script>
</body>
</html>
EOF

echo "✅ Generated $OUTPUT_HTML with inline OpenAPI spec"
