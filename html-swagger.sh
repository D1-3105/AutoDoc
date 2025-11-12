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
  <style>
    .info .title small.version-stamp {
      display: inline-flex;
      align-items: center;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>

  <script src="$BUNDLE_JS"></script>
  <script src="$PRESET_JS"></script>
  <script>
    window.onload = function() {
      const spec = $OPENAPI_CONTENT;

      const customPlugin = function(system) {
        return {
          wrapComponents: {
            InfoContainer: (Original, system) => (props) => {
              const React = system.React;
              return React.createElement(
                'div',
                null,
                React.createElement(Original, props),
                React.createElement('button', {
                  style: {
                    marginLeft: '15px',
                    padding: '8px 16px',
                    backgroundColor: '#4990e2',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    fontSize: '14px',
                    fontWeight: '500'
                  },
                  onClick: async () => {
                    try {
                      let url = window.location.pathname;
                      url = url.replace(/\/index\.html$|\/$/, '');
                      if (!url.endsWith('.json')) {
                        url = "/api/v1/expand" + url + '.json';
                      }
                      const res = await fetch(url);
                      const data = await res.json();

                      // Create blob and download
                      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
                      const downloadUrl = window.URL.createObjectURL(blob);
                      const a = document.createElement('a');
                      a.href = downloadUrl;

                      // Extract filename from URL
                      const urlPath = url.split('/').pop();
                      a.download = urlPath || 'schema.json';

                      document.body.appendChild(a);
                      a.click();
                      document.body.removeChild(a);
                      window.URL.revokeObjectURL(downloadUrl);

                      alert("Schema downloaded successfully!");
                    } catch(e) {
                      alert("Error: " + e);
                    }
                  }
                }, "Get expanded schema")
              );
            }
          }
        }
      };

      SwaggerUIBundle({
        spec: spec,
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
        plugins: [customPlugin]
      });
    }
  </script>
</body>
</html>
EOF

echo "✅ Generated $OUTPUT_HTML with Swagger UI"
