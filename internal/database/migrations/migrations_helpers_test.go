package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMigrationTestDB(t *testing.T) (*sql.DB, *sql.Conn, context.Context) {
	t.Helper()

	dsn := fmt.Sprintf("file:migrations_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := gdb.DB()
	require.NoError(t, err)

	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		_ = sqlDB.Close()
	})

	return sqlDB, conn, ctx
}

func TestSplitSQLTopLevelCommaList(t *testing.T) {
	input := "id INTEGER, title TEXT DEFAULT 'a,b', note TEXT CHECK(note IN (\"x,y\", 'z,w')), `tag` TEXT"
	parts := splitSQLTopLevelCommaList(input)
	require.Len(t, parts, 4)
	assert.Equal(t, "id INTEGER", strings.TrimSpace(parts[0]))
	assert.Equal(t, "title TEXT DEFAULT 'a,b'", strings.TrimSpace(parts[1]))
	assert.Equal(t, "note TEXT CHECK(note IN (\"x,y\", 'z,w'))", strings.TrimSpace(parts[2]))
	assert.Equal(t, "`tag` TEXT", strings.TrimSpace(parts[3]))
}

func TestIsUniqueSingleColumnDMMConstraintDefinition(t *testing.T) {
	tests := []struct {
		name    string
		segment string
		want    bool
	}{
		{name: "simple unique", segment: "UNIQUE(dmm_id)", want: true},
		{name: "constraint prefix", segment: `CONSTRAINT uq_dmm UNIQUE ( "dmm_id" )`, want: true},
		{name: "quoted constraint name", segment: `CONSTRAINT "uq_dmm" UNIQUE ([dmm_id])`, want: true},
		{name: "multi column unique", segment: "UNIQUE(dmm_id, japanese_name)", want: false},
		{name: "non-unique definition", segment: "PRIMARY KEY(dmm_id)", want: false},
		{name: "malformed constraint", segment: "CONSTRAINT only_name", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isUniqueSingleColumnDMMConstraintDefinition(tt.segment))
		})
	}
}

func TestRemoveInlineUniqueDMMConstraint(t *testing.T) {
	tests := []struct {
		name        string
		segment     string
		wantChanged bool
		contains    string
		notContains string
	}{
		{
			name:        "removes simple unique",
			segment:     "dmm_id INTEGER UNIQUE",
			wantChanged: true,
			contains:    "dmm_id INTEGER",
			notContains: "UNIQUE",
		},
		{
			name:        "removes constraint unique with conflict clause",
			segment:     "dmm_id INTEGER CONSTRAINT uq_dmm UNIQUE ON CONFLICT REPLACE",
			wantChanged: true,
			contains:    "dmm_id INTEGER",
			notContains: "UNIQUE",
		},
		{
			name:        "does not touch other columns",
			segment:     "title TEXT UNIQUE",
			wantChanged: false,
			contains:    "title TEXT UNIQUE",
		},
		{
			name:        "does not touch dmm without unique",
			segment:     "dmm_id INTEGER",
			wantChanged: false,
			contains:    "dmm_id INTEGER",
		},
		{
			name:        "ignores keyword in quoted literal",
			segment:     "dmm_id TEXT DEFAULT 'unique'",
			wantChanged: false,
			contains:    "DEFAULT 'unique'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := removeInlineUniqueDMMConstraint(tt.segment)
			assert.Equal(t, tt.wantChanged, changed)
			assert.Contains(t, got, tt.contains)
			if tt.notContains != "" {
				assert.NotContains(t, got, tt.notContains)
			}
		})
	}
}

func TestIdentifierHelpers(t *testing.T) {
	segment := `CONSTRAINT "uq dmm" UNIQUE (dmm_id)`
	uniquePos := strings.Index(segment, "UNIQUE")
	require.GreaterOrEqual(t, uniquePos, 0)

	nameStart, nameEnd, ok := previousIdentifierTokenBounds(segment, uniquePos)
	require.True(t, ok)
	assert.Equal(t, `"uq dmm"`, segment[nameStart:nameEnd])

	kwStart, kwEnd, ok := previousIdentifierTokenBounds(segment, nameStart)
	require.True(t, ok)
	assert.Equal(t, "CONSTRAINT", segment[kwStart:kwEnd])

	constraintStart, ok := findConstraintPrefixStart(segment, uniquePos)
	require.True(t, ok)
	assert.Equal(t, kwStart, constraintStart)

	_, _, ok = previousIdentifierTokenBounds("   ", 1)
	assert.False(t, ok)

	plainSegment := "dmm_id UNIQUE"
	uniquePos = strings.Index(plainSegment, "UNIQUE")
	require.GreaterOrEqual(t, uniquePos, 0)
	_, ok = findConstraintPrefixStart(plainSegment, uniquePos)
	assert.False(t, ok)
}

func TestHasKeywordAt(t *testing.T) {
	input := "dmm_id UNIQUE ON CONFLICT REPLACE"
	pos := strings.Index(input, "UNIQUE")
	require.GreaterOrEqual(t, pos, 0)
	assert.True(t, hasKeywordAt(input, pos, "unique"))
	assert.False(t, hasKeywordAt(input, pos, "uniq"))
	assert.False(t, hasKeywordAt("xUNIQUEy", 1, "UNIQUE"))
	assert.False(t, hasKeywordAt(input, -1, "UNIQUE"))
}

func TestExtractLeadingIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "double quoted", input: `"dmm_id" INTEGER`, want: "dmm_id", ok: true},
		{name: "backtick quoted", input: "`dmm_id` INTEGER", want: "dmm_id", ok: true},
		{name: "bracket quoted", input: "[dmm_id] INTEGER", want: "dmm_id", ok: true},
		{name: "plain identifier", input: "dmm_id INTEGER", want: "dmm_id", ok: true},
		{name: "unterminated quote", input: `"dmm_id INTEGER`, want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractLeadingIdentifier(tt.input)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDesiredDMMIDPartialIndexSQL(t *testing.T) {
	assert.True(t, isDesiredDMMIDPartialIndexSQL(`
		CREATE UNIQUE INDEX "idx_actresses_dmm_id_positive"
		ON "actresses" ("dmm_id")
		WHERE dmm_id > 0
	`))
	assert.True(t, isDesiredDMMIDPartialIndexSQL("create unique index idx on actresses(dmm_id) where dmm_id>0"))
	assert.False(t, isDesiredDMMIDPartialIndexSQL("CREATE INDEX idx ON actresses(dmm_id) WHERE dmm_id > 0"))
	assert.False(t, isDesiredDMMIDPartialIndexSQL("CREATE UNIQUE INDEX idx ON movies(content_id) WHERE content_id <> ''"))
	assert.False(t, isDesiredDMMIDPartialIndexSQL("CREATE UNIQUE INDEX idx ON actresses(dmm_id) WHERE dmm_id >= 0"))
}

func TestBuildActressesRebuildCreateTableSQL(t *testing.T) {
	_, conn, ctx := newMigrationTestDB(t)

	_, err := conn.ExecContext(ctx, `
		CREATE TABLE actresses (
			id INTEGER PRIMARY KEY,
			dmm_id INTEGER CONSTRAINT uq_dmm UNIQUE ON CONFLICT REPLACE,
			japanese_name TEXT
		)
	`)
	require.NoError(t, err)

	sqlText, err := buildActressesRebuildCreateTableSQL(ctx, conn)
	require.NoError(t, err)
	assert.Contains(t, sqlText, "CREATE TABLE actresses_new")
	assert.Contains(t, sqlText, "japanese_name TEXT")
	assert.NotContains(t, strings.ToUpper(sqlText), "UNIQUE ON CONFLICT")

	_, err = conn.ExecContext(ctx, "DROP TABLE actresses")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
		CREATE TABLE actresses (
			id INTEGER PRIMARY KEY,
			dmm_id INTEGER,
			japanese_name TEXT
		)
	`)
	require.NoError(t, err)

	_, err = buildActressesRebuildCreateTableSQL(ctx, conn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected UNIQUE(dmm_id) constraint was not found")
}

func TestBuildActressesRebuildCopySQL(t *testing.T) {
	_, conn, ctx := newMigrationTestDB(t)

	_, err := conn.ExecContext(ctx, `
		CREATE TABLE actresses (
			id INTEGER PRIMARY KEY,
			dmm_id INTEGER,
			japanese_name TEXT
		)
	`)
	require.NoError(t, err)

	sqlText, err := buildActressesRebuildCopySQL(ctx, conn)
	require.NoError(t, err)
	assert.Contains(t, sqlText, "INSERT INTO actresses_new")
	assert.Contains(t, sqlText, "CASE WHEN dmm_id < 0 THEN 0 ELSE dmm_id END")
	assert.Contains(t, sqlText, `"japanese_name"`)

	_, conn2, ctx2 := newMigrationTestDB(t)
	_, err = buildActressesRebuildCopySQL(ctx2, conn2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load actresses columns for rebuild copy")
}

func TestLoadActressIndexes(t *testing.T) {
	_, conn, ctx := newMigrationTestDB(t)

	_, err := conn.ExecContext(ctx, `
		CREATE TABLE actresses (
			id INTEGER PRIMARY KEY,
			dmm_id INTEGER,
			japanese_name TEXT
		)
	`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "CREATE UNIQUE INDEX idx_actresses_dmm_id_positive ON actresses(dmm_id) WHERE dmm_id > 0")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "CREATE INDEX idx_actresses_japanese_name ON actresses(japanese_name)")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "CREATE INDEX idx_expr ON actresses(lower(japanese_name))")
	require.NoError(t, err)

	indexes, err := loadActressIndexes(ctx, conn)
	require.NoError(t, err)
	require.NotEmpty(t, indexes)

	byName := make(map[string]actressIndexMeta, len(indexes))
	for _, idx := range indexes {
		byName[idx.Name] = idx
	}

	canonical, ok := byName[canonicalPositiveDMMIndexName]
	require.True(t, ok)
	assert.True(t, isUniqueSingleColumnDMMIndex(canonical))
	assert.True(t, isCanonicalPositiveDMMIndex(canonical))

	expr, ok := byName["idx_expr"]
	require.True(t, ok)
	assert.True(t, expr.HasUnsupportedKeyParts)
}

func TestActressesTableExistsAndFindDuplicatePositiveDMMID(t *testing.T) {
	_, conn, ctx := newMigrationTestDB(t)

	exists, err := actressesTableExists(ctx, conn)
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = conn.ExecContext(ctx, `
		CREATE TABLE actresses (
			id INTEGER PRIMARY KEY,
			dmm_id INTEGER
		)
	`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "INSERT INTO actresses (dmm_id) VALUES (1), (1), (0), (-2)")
	require.NoError(t, err)

	exists, err = actressesTableExists(ctx, conn)
	require.NoError(t, err)
	assert.True(t, exists)

	dupID, dupCount, err := findDuplicatePositiveDMMID(ctx, conn)
	require.NoError(t, err)
	assert.Equal(t, 1, dupID)
	assert.Equal(t, 2, dupCount)

	_, err = conn.ExecContext(ctx, "DELETE FROM actresses WHERE id = (SELECT id FROM actresses WHERE dmm_id = 1 LIMIT 1)")
	require.NoError(t, err)

	dupID, dupCount, err = findDuplicatePositiveDMMID(ctx, conn)
	require.NoError(t, err)
	assert.Equal(t, 0, dupID)
	assert.Equal(t, 0, dupCount)
}

func TestEnsureTableColumnsAndTableColumnsByName(t *testing.T) {
	_, conn, ctx := newMigrationTestDB(t)

	_, err := conn.ExecContext(ctx, `CREATE TABLE actresses (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)

	columns, err := tableColumnsByName(ctx, conn, "actresses")
	require.NoError(t, err)
	_, ok := columns["id"]
	assert.True(t, ok)

	err = ensureTableColumns(ctx, conn, "actresses", []columnSpec{
		{Name: "id", Definition: "INTEGER"},
		{Name: "dmm_id", Definition: "INTEGER"},
		{Name: "japanese_name", Definition: "TEXT"},
	})
	require.NoError(t, err)

	columns, err = tableColumnsByName(ctx, conn, "actresses")
	require.NoError(t, err)
	_, ok = columns["dmm_id"]
	assert.True(t, ok)
	_, ok = columns["japanese_name"]
	assert.True(t, ok)

	err = ensureTableColumns(ctx, conn, "missing_table", []columnSpec{{Name: "x", Definition: "TEXT"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add missing column")
}

func TestReconcileLegacyMoviesContentID(t *testing.T) {
	t.Run("backfills from id", func(t *testing.T) {
		_, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE movies (content_id TEXT, id TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `INSERT INTO movies(content_id, id) VALUES ('', 'ABP-123')`)
		require.NoError(t, err)

		require.NoError(t, reconcileLegacyMoviesContentID(ctx, conn))

		var contentID string
		err = conn.QueryRowContext(ctx, `SELECT content_id FROM movies LIMIT 1`).Scan(&contentID)
		require.NoError(t, err)
		assert.Equal(t, "abp123", contentID)
	})

	t.Run("fails when both content_id and id are missing", func(t *testing.T) {
		_, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE movies (content_id TEXT, id TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `INSERT INTO movies(content_id, id) VALUES ('', '')`)
		require.NoError(t, err)

		err = reconcileLegacyMoviesContentID(ctx, conn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing both content_id and id")
	})

	t.Run("fails on duplicate derived content_id", func(t *testing.T) {
		_, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE movies (content_id TEXT, id TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `
			INSERT INTO movies(content_id, id) VALUES
			('', 'ABP-001'),
			('', 'ABP001')
		`)
		require.NoError(t, err)

		err = reconcileLegacyMoviesContentID(ctx, conn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate content_id=")
	})
}

func TestMigrateActressesDMMIDIndexUp(t *testing.T) {
	t.Run("no actresses table is a no-op", func(t *testing.T) {
		sqlDB, _, _ := newMigrationTestDB(t)
		err := migrateActressesDMMIDIndexUp(context.Background(), sqlDB)
		require.NoError(t, err)
	})

	t.Run("returns error on duplicate positive dmm_id", func(t *testing.T) {
		sqlDB, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE actresses (id INTEGER PRIMARY KEY, dmm_id INTEGER, japanese_name TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `INSERT INTO actresses(dmm_id, japanese_name) VALUES (123, 'A'), (123, 'B')`)
		require.NoError(t, err)

		err = migrateActressesDMMIDIndexUp(context.Background(), sqlDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate positive dmm_id=123")
	})

	t.Run("drops legacy index, normalizes negatives, and creates canonical indexes", func(t *testing.T) {
		sqlDB, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE actresses (id INTEGER PRIMARY KEY, dmm_id INTEGER, japanese_name TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `CREATE UNIQUE INDEX legacy_dmm_idx ON actresses(dmm_id)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `INSERT INTO actresses(dmm_id, japanese_name) VALUES (-7, 'A'), (999, 'B')`)
		require.NoError(t, err)

		require.NoError(t, migrateActressesDMMIDIndexUp(context.Background(), sqlDB))

		var negatives int
		err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM actresses WHERE dmm_id < 0`).Scan(&negatives)
		require.NoError(t, err)
		assert.Equal(t, 0, negatives)

		var legacyCount int
		err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='legacy_dmm_idx'`).Scan(&legacyCount)
		require.NoError(t, err)
		assert.Equal(t, 0, legacyCount)

		var canonicalCount int
		err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_actresses_dmm_id_positive'`).Scan(&canonicalCount)
		require.NoError(t, err)
		assert.Equal(t, 1, canonicalCount)

		var jpIndexCount int
		err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_actresses_japanese_name'`).Scan(&jpIndexCount)
		require.NoError(t, err)
		assert.Equal(t, 1, jpIndexCount)
	})
}

func TestRebuildActressesTable_ErrorHandlingAndRollback(t *testing.T) {
	t.Run("begin immediate failure", func(t *testing.T) {
		_, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `CREATE TABLE actresses (id INTEGER PRIMARY KEY, dmm_id INTEGER UNIQUE, japanese_name TEXT)`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE")
		require.NoError(t, err)
		defer func() { _, _ = conn.ExecContext(context.Background(), "ROLLBACK") }()

		err = rebuildActressesTable(ctx, conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "begin actresses rebuild transaction")
	})

	t.Run("rollback on preserved index recreation error", func(t *testing.T) {
		_, conn, ctx := newMigrationTestDB(t)
		_, err := conn.ExecContext(ctx, `
			CREATE TABLE actresses (
				id INTEGER PRIMARY KEY,
				dmm_id INTEGER UNIQUE,
				japanese_name TEXT
			)
		`)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, `INSERT INTO actresses(dmm_id, japanese_name) VALUES (10, 'A')`)
		require.NoError(t, err)

		err = rebuildActressesTable(ctx, conn, []string{`CREATE INDEX idx_bad_preserved ON actresses(does_not_exist)`})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "recreate preserved actresses index")

		var rowCount int
		err = conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM actresses`).Scan(&rowCount)
		require.NoError(t, err)
		assert.Equal(t, 1, rowCount)
	})
}
