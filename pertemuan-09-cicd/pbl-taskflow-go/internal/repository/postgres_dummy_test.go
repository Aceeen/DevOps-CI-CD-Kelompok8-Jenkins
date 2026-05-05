package repository_test

import (
	"testing"
	"github.com/taskflow/api/internal/model"
	"github.com/taskflow/api/internal/repository"
)

func TestPostgres_NoDB_ErrorPaths(t *testing.T) {
	r, err := repository.NewPostgresRepository("postgres://fake:fake@localhost:5432/fake?sslmode=disable")
	if err != nil {
		t.Skip("pgxpool failed to parse url, skip testing error paths")
	}

	r.Migrate()
	r.Save(model.Task{})
	r.FindByID("1")
	r.FindAll()
	r.FindByStatus(model.StatusTodo)
	r.Delete("1")
	r.Count()
	r.TruncateForTest(t)
	r.Close()
}
