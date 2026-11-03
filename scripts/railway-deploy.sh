#!/usr/bin/env sh
# Runtime tokens are inherited securely; never put secrets in command arguments.
set -eu
# Accept the authorized OAuth CLI session as well as a securely configured token.
: "${RAILWAY_PROJECT_ID:?Set the existing project ID}"
: "${RAILWAY_ENVIRONMENT:?Set the environment ID or name}"
: "${RAILWAY_WEB_SERVICE:?Set the web service ID or name}"
: "${RAILWAY_ENGINE_SERVICE:?Set the engine service ID or name}"
command -v railway >/dev/null
command -v node >/dev/null
if [ -z "${RAILWAY_TOKEN:-}" ] && [ -z "${RAILWAY_API_TOKEN:-}" ]; then
  railway whoami >/dev/null
fi
sh "$(dirname "$0")/railway-configure.sh"
wait_deploy() {
  service="$1"
  deployment_id="$(railway deployment list --project "$RAILWAY_PROJECT_ID" --environment "$RAILWAY_ENVIRONMENT" --service "$service" --limit 1 --json | node -e 'let s="";process.stdin.on("data",d=>s+=d);process.stdin.on("end",()=>{const a=JSON.parse(s);if(!Array.isArray(a)||!a[0]?.id)process.exit(1);process.stdout.write(a[0].id)})')"
  attempt=0
  while [ "$attempt" -lt 120 ]; do
    status="$(railway deployment list --project "$RAILWAY_PROJECT_ID" --environment "$RAILWAY_ENVIRONMENT" --service "$service" --limit 10 --json | node -e 'let s="";process.stdin.on("data",d=>s+=d);process.stdin.on("end",()=>{const a=JSON.parse(s);process.stdout.write(a.find(x=>x.id===process.argv[1])?.status||"UNKNOWN")})' "$deployment_id")"
    case "$status" in
      SUCCESS) return 0 ;;
      FAILED|CRASHED|CANCELED|REMOVED) echo "Deployment failed: $service ($status)" >&2; return 1 ;;
    esac
    attempt=$((attempt + 1))
    sleep 5
  done
  echo "Deployment readiness timed out: $service" >&2
  return 1
}
# Execution settings are applied by railway-configure.sh through the current API.
# --ci exits after the build log stream; it does not itself prove readiness.
railway up --project "$RAILWAY_PROJECT_ID" --environment "$RAILWAY_ENVIRONMENT" --service "$RAILWAY_WEB_SERVICE" --ci
wait_deploy "$RAILWAY_WEB_SERVICE"
railway up --project "$RAILWAY_PROJECT_ID" --environment "$RAILWAY_ENVIRONMENT" --service "$RAILWAY_ENGINE_SERVICE" --ci
wait_deploy "$RAILWAY_ENGINE_SERVICE"
