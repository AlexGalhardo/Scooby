package tests

import (
	"context"
	"crypto/rand"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/Scooby/bootstrap"
)

// TestPostgresPosts exercises migrations and the Inertia CRUD on a real database.
// Set TEST_DATABASE_URL to a dedicated test database; never a production database.
func TestPostgresPosts(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": dsn},
	})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	if got := db.Dialect().Name(); got != "postgres" {
		t.Fatalf("expected PostgreSQL, got %s", got)
	}
	title := "Integration " + rand.Text()
	api := app.Client(t).AcceptJSON()
	api.PostJSON("/posts", map[string]any{"title": title, "body": "PostgreSQL round trip"}).
		AssertCreated().AssertJSONPath("title", title)
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM posts WHERE title = $1", title); err != nil {
			t.Errorf("clean up integration post: %v", err)
		}
	})
	var id int64
	if err := db.QueryRowContext(t.Context(), "SELECT id FROM posts WHERE title = $1", title).Scan(&id); err != nil {
		t.Fatal(err)
	}
	path := "/posts/" + strconv.FormatInt(id, 10)
	client := app.Client(t).Inertia("")
	client.Get(path).AssertOK().AssertInertia("Posts/Show").PropEquals("post.title", title)
	client.PutJSON(path, map[string]any{"title": title, "body": "Updated"}).AssertRedirect(path)
	api.GetJSON(path).AssertOK().AssertJSONPath("body", "Updated")
	client.Delete(path).AssertRedirect("/posts")
	api.GetJSON("/posts?q="+url.QueryEscape(title)).AssertOK().AssertJSONCount("data", 0)
}
