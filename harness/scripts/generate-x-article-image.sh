#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  generate-x-article-image.sh \
    --prompt-file <path> \
    --output <png> \
    --request-out <json> \
    --response-out <json> \
    [--logo <png>] \
    [--aspect-ratio <ratio>] \
    [--image-size <size>]

Defaults:
  --logo docs/images/claude-harness-logo-with-text.png
  --aspect-ratio 16:9
  --image-size 2K

This script loads GOOGLE_AI_API_KEY or GEMINI_API_KEY from the environment.
If .env exists in the repo root, it is sourced with export enabled.
EOF
}

prompt_file=""
output_file=""
request_out=""
response_out=""
logo_file="docs/images/claude-harness-logo-with-text.png"
aspect_ratio="16:9"
image_size="2K"
tmp_raw=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prompt-file) prompt_file="$2"; shift 2 ;;
    --output) output_file="$2"; shift 2 ;;
    --request-out) request_out="$2"; shift 2 ;;
    --response-out) response_out="$2"; shift 2 ;;
    --logo) logo_file="$2"; shift 2 ;;
    --aspect-ratio) aspect_ratio="$2"; shift 2 ;;
    --image-size) image_size="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 1 ;;
  esac
done

[[ -n "$prompt_file" ]] || { echo "--prompt-file is required" >&2; exit 1; }
[[ -n "$output_file" ]] || { echo "--output is required" >&2; exit 1; }
[[ -n "$request_out" ]] || { echo "--request-out is required" >&2; exit 1; }
[[ -n "$response_out" ]] || { echo "--response-out is required" >&2; exit 1; }
[[ -f "$prompt_file" ]] || { echo "Prompt file not found: $prompt_file" >&2; exit 1; }
[[ -f "$logo_file" ]] || { echo "Logo file not found: $logo_file" >&2; exit 1; }

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env >/dev/null 2>&1 || true
  set +a
fi

api_key="${GOOGLE_AI_API_KEY:-${GEMINI_API_KEY:-}}"
[[ -n "$api_key" ]] || { echo "GOOGLE_AI_API_KEY or GEMINI_API_KEY is required" >&2; exit 1; }

mkdir -p "$(dirname "$output_file")" "$(dirname "$request_out")" "$(dirname "$response_out")"

logo_b64="$(base64 -i "$logo_file" | tr -d '\n')"
prompt_text="$(cat "$prompt_file")"

jq -n \
  --arg logo_b64 "$logo_b64" \
  --arg prompt "$prompt_text" \
  --arg aspect_ratio "$aspect_ratio" \
  --arg image_size "$image_size" \
  '{
    contents: [{
      parts: [
        {
          inlineData: {
            mimeType: "image/png",
            data: $logo_b64
          }
        },
        {
          text: $prompt
        }
      ]
    }],
    generationConfig: {
      responseModalities: ["TEXT", "IMAGE"],
      imageConfig: {
        aspectRatio: $aspect_ratio,
        imageSize: $image_size
      }
    }
  }' > "$request_out"

curl -sS -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent" \
  -H "x-goog-api-key: ${api_key}" \
  -H "Content-Type: application/json" \
  -d @"$request_out" \
  -o "$response_out"

img_b64="$(jq -r '.candidates[0].content.parts[]? | select(.inlineData or .inline_data) | (.inlineData.data // .inline_data.data)' "$response_out" | head -1)"
img_mime="$(jq -r '.candidates[0].content.parts[]? | select(.inlineData or .inline_data) | (.inlineData.mimeType // .inline_data.mime_type // "image/png")' "$response_out" | head -1)"
[[ -n "$img_b64" && "$img_b64" != "null" ]] || {
  echo "Image data not found in response: $response_out" >&2
  jq -r '.error // .candidates[0].content.parts[]?.text // empty' "$response_out" >&2 || true
  exit 1
}

case "$img_mime" in
  image/jpeg) tmp_raw="${output_file}.raw.jpg" ;;
  image/png) tmp_raw="${output_file}.raw.png" ;;
  *) tmp_raw="${output_file}.raw.bin" ;;
esac

trap '[[ -n "${tmp_raw:-}" && -f "${tmp_raw:-}" ]] && rm -f "${tmp_raw:-}"' EXIT
printf '%s' "$img_b64" | base64 -d > "$tmp_raw"

output_ext="${output_file##*.}"
case "$output_ext" in
  png)
    if [[ "$img_mime" == "image/png" ]]; then
      mv "$tmp_raw" "$output_file"
    else
      sips -s format png "$tmp_raw" --out "$output_file" >/dev/null
      rm -f "$tmp_raw"
    fi
    ;;
  jpg|jpeg)
    if [[ "$img_mime" == "image/jpeg" ]]; then
      mv "$tmp_raw" "$output_file"
    else
      sips -s format jpeg "$tmp_raw" --out "$output_file" >/dev/null
      rm -f "$tmp_raw"
    fi
    ;;
  *)
    mv "$tmp_raw" "$output_file"
    ;;
esac

echo "Generated: $output_file"
