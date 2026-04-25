#!/bin/sh
set -eu

public_url="${FRONTEND_PUBLIC_URL:-http://localhost:5173}"

printf '\nFrontend is available at: %s\n\n' "$public_url"
