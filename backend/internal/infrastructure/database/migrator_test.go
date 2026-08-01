package database

import (
	"context"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/muhammadyunus/Restify-Service/migrations"
)

func TestMigrationFilesEmbedded(t *testing.T) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}

	var sqlCount int

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlCount++
		}
	}

	if sqlCount != 26 {
		t.Errorf("embedded sql files = %d, want 26 (13 up + 13 down)", sqlCount)
	}
}

func TestListVersions(t *testing.T) {
	m := &Migrator{}

	versions, err := m.ListVersions(context.Background())
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}

	want := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	if !slices.Equal(versions, want) {
		t.Errorf("versions = %v, want %v", versions, want)
	}
}

func TestParseMigrationVersion(t *testing.T) {
	cases := []struct {
		name string
		want uint
	}{
		{name: "000001_create_users_table.up.sql", want: 1},
		{name: "000013_create_jwt_blacklist_table.down.sql", want: 13},
		{name: "no_version.sql", want: 0},
		{name: "README.md", want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseMigrationVersion(tc.name); got != tc.want {
				t.Errorf("parseMigrationVersion(%q) = %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}
