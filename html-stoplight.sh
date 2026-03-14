#!/bin/bash
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

python3 << PYEOF
import base64
import json

with open("$INPUT_JSON", "r") as f:
    spec = json.load(f)

spec_json = json.dumps(spec)
spec_b64 = base64.b64encode(spec_json.encode()).decode()

html = f'''<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  <style>
    body {{
      margin: 0;
      padding: 0;
    }}
    elements-api {{
      display: block;
      height: 100vh;
    }}
  </style>
</head>
<body>
  <elements-api router="hash" layout="sidebar"></elements-api>
  <script type="module">
    import 'https://unpkg.com/@stoplight/elements/web-components.min.js?module';

    const spec = JSON.parse(atob('{spec_b64}'));

    customElements.whenDefined('elements-api').then(() => {{
      const el = document.querySelector('elements-api');
      el.apiDescriptionDocument = spec;
    }});
  </script>
</body>
</html>
'''

with open("$OUTPUT_HTML", "w") as f:
    f.write(html)

print("✅ Generated $OUTPUT_HTML with Stoplight Elements", flush=True)
PYEOF
