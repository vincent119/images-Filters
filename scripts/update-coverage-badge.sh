#!/bin/bash
set -e

# Generate coverage output if not present or stale
if [ ! -f coverage.out ]; then
    echo "coverage.out not found, running tests..."
    go test -cover -coverprofile=coverage.out ./...
fi

# Extract total coverage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Calculated Coverage: ${COVERAGE}%"

# Color logic based on coverage
COLOR="red"
if (( $(echo "$COVERAGE > 80" | bc -l) )); then
  COLOR="brightgreen"
elif (( $(echo "$COVERAGE > 60" | bc -l) )); then
  COLOR="green"
elif (( $(echo "$COVERAGE > 40" | bc -l) )); then
  COLOR="yellow"
else
  COLOR="red"
fi

# Determine Gist file name from existing URL if possible, or default
GIST_FILE="images-filters-coverage.json"

# Create JSON content
cat <<EOF > $GIST_FILE
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "${COVERAGE}%",
  "color": "${COLOR}"
}
EOF

echo "Generated $GIST_FILE"
echo "JSON Content:"
cat $GIST_FILE

# Auto-update if GIST_ID is present
if [ -n "$GIST_ID" ]; then
    echo ""
    echo "GIST_ID found, attempting to update Gist..."
    if command -v gh &> /dev/null; then
        gh gist edit "$GIST_ID" "$GIST_FILE"
        echo "Gist updated successfully!"
    else
        echo "Error: gh cli not found. Cannot update Gist."
        exit 1
    fi
else
    echo ""
    echo "To update your Gist, run:"
    echo "gh gist edit cca471fced090cd840f0d85a5e876305 $GIST_FILE"
fi
