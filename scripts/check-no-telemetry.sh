#!/usr/bin/env sh
# Fails when analytics, crash-reporting, or telemetry endpoints appear in the
# code that ships to users. Honey Ledger is local-first: it must not phone home.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

cd "$ROOT_DIR"

# Hosts and SDK identifiers used for product analytics or crash reporting.
# Keep this list to services that exist to collect user behaviour; providers the
# user explicitly configures (LLM, map, mail, object storage) are out of scope.
PATTERN='google-analytics\.com|googletagmanager\.com|analytics\.google\.com|hm\.baidu\.com|umeng\.com|mixpanel\.com|posthog\.com|amplitude\.com|hotjar\.com|fullstory\.com|segment\.(io|com)|bugsnag\.com|sentry\.io|appcenter\.ms|datadoghq\.com|newrelic\.com|@sentry/|firebase-analytics'

SCAN_PATHS=""
for candidate in src electron pkg cmd main.go; do
    if [ -e "$candidate" ]; then
        SCAN_PATHS="$SCAN_PATHS $candidate"
    fi
done

# shellcheck disable=SC2086
matches=$(grep -rInE \
    --exclude-dir=node_modules \
    --exclude-dir=dist \
    --exclude-dir=vendor \
    --exclude="*.min.js" \
    --exclude="*.map" \
    "$PATTERN" $SCAN_PATHS 2>/dev/null || true)

if [ -n "$matches" ]; then
    echo "Telemetry check failed: analytics or crash-reporting endpoints found."
    echo ""
    echo "$matches"
    echo ""
    echo "Honey Ledger is local-first and must not report usage data. Remove the"
    echo "reference, or extend scripts/check-no-telemetry.sh if it is a false"
    echo "positive (for example a user-configured provider)."
    exit 1
fi

echo "Telemetry check passed: no analytics or crash-reporting endpoints found."
