#!/usr/bin/env bash
# Rehearses a backup and restore on a throwaway PostgreSQL container: migrate a database, add rows, dump
# it, restore the dump into a second database and compare every table's row count. It never touches a
# real database. Requires Docker and Go. Prints PASS or exits non-zero.
set -euo pipefail
# Git Bash on Windows rewrites paths such as /tmp in arguments; this keeps them as written.
export MSYS_NO_PATHCONV=1

image="postgres:18.6-alpine3.24"
name="vt-restore-rehearsal"
port="55433"

cleanup() { docker rm -f "$name" >/dev/null 2>&1 || true; }
trap cleanup EXIT
cleanup

docker run -d --name "$name" -e POSTGRES_USER=rehearsal -e POSTGRES_PASSWORD=rehearsal \
	-p "127.0.0.1:${port}:5432" "$image" >/dev/null
for _ in $(seq 1 40); do
	docker exec "$name" pg_isready -U rehearsal >/dev/null 2>&1 && break
	sleep 1
done

psql() { docker exec -i "$name" psql -U rehearsal -v ON_ERROR_STOP=1 -qAt "$@"; }
psql -d postgres -c "CREATE DATABASE source" -c "CREATE DATABASE restored"

key="base64:$(head -c 32 /dev/urandom | base64)"
DATABASE_URL="postgres://rehearsal:rehearsal@127.0.0.1:${port}/source?sslmode=disable" APP_KEY="$key" \
	go run ./cmd/app migrate >/dev/null

psql -d source <<'SQL'
INSERT INTO preferences (id, payload, version, updated_at, status)
VALUES ('owner', '{"account":"b3","symbol":"PETR4","watchlist":["PETR4"],"alerts":[],"templates":[]}', 1, now()::text, 'ready');
INSERT INTO audit_events (actor, action, account_id, details, created_at)
VALUES ('rehearsal', 'backup.test', 'b3', 'row that must survive the restore', now()::text);
SQL

start=$(date +%s%N)
docker exec "$name" pg_dump -U rehearsal -Fc -d source -f /tmp/source.dump
mid=$(date +%s%N)
docker exec "$name" pg_restore -U rehearsal -d restored --no-owner /tmp/source.dump
end=$(date +%s%N)
echo "dump: $(( (mid - start) / 1000000 )) ms, restore: $(( (end - mid) / 1000000 )) ms (includes the docker exec overhead)"

counts() {
	local db="$1"
	local tables
	tables=$(psql -d "$db" -c "SELECT table_name FROM information_schema.tables WHERE table_schema='public' ORDER BY 1" </dev/null)
	for table in $tables; do
		echo "${table} $(psql -d "$db" -c "SELECT count(*) FROM \"${table}\"" </dev/null)"
	done
}

counts source >/tmp/rehearsal-source.txt
counts restored >/tmp/rehearsal-restored.txt
echo "tables restored: $(wc -l </tmp/rehearsal-restored.txt)"
cat /tmp/rehearsal-restored.txt
if diff -u /tmp/rehearsal-source.txt /tmp/rehearsal-restored.txt; then
	echo "PASS: every table has the same row count after the restore"
else
	echo "FAIL: the restored database differs from the source" >&2
	exit 1
fi
