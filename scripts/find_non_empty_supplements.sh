#!/bin/bash
find source -type f -regex '.*/[0-9]+\.json' -print0 | xargs -0 -I {} jq -r 'select(.supplements | length > 0) | input_filename' {} | wc -l
