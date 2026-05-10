#!/bin/sh
set -eu

public_url="${FRONTEND_PUBLIC_URL:-http://localhost}"

printf '\nFrontend is available at: %s\n' "$public_url"
printf '\nVK (Vite, same values as compose / runtime env used at image build):\n'
printf '  VITE_VK_APP_ID=%s\n' "${VITE_VK_APP_ID:-<not set>}"
printf '  VITE_VK_REDIRECT_URI=%s\n\n' "${VITE_VK_REDIRECT_URI:-<not set>}"
