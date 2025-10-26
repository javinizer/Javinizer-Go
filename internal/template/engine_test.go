package template

import (
	"testing"
	"time"
)

func TestTemplateEngine_Execute(t *testing.T) {
	engine := NewEngine()
	releaseDate := time.Date(2020, 9, 13, 0, 0, 0, 0, time.UTC)

	ctx := &Context{
		ID:          "IPX-535",
		ContentID:   "ipx00535",
		Title:       "Test Movie Title",
		ReleaseDate: &releaseDate,
		Runtime:     120,
		Director:    "Test Director",
		Maker:       "Test Studio",
		Label:       "Test Label",
		Series:      "Test Series",
		Actresses:   []string{"Sakura Momo", "Test Actress"},
		Genres:      []string{"Genre1", "Genre2", "Genre3"},
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "Simple ID",
			template: "<ID>",
			want:     "IPX-535",
		},
		{
			name:     "ID with title",
			template: "<ID> - <TITLE>",
			want:     "IPX-535 - Test Movie Title",
		},
		{
			name:     "Complex format",
			template: "<ID> [<STUDIO>] - <TITLE> (<YEAR>)",
			want:     "IPX-535 [Test Studio] - Test Movie Title (2020)",
		},
		{
			name:     "With runtime",
			template: "<ID> - <TITLE> (<RUNTIME>min)",
			want:     "IPX-535 - Test Movie Title (120min)",
		},
		{
			name:     "Truncated title",
			template: "<ID> - <TITLE:20>",
			want:     "IPX-535 - Test Movie Title",
		},
		{
			name:     "Truncated long title",
			template: "<TITLE:10>",
			want:     "Test Mo...",
		},
		{
			name:     "Date format default",
			template: "<ID> (<RELEASEDATE>)",
			want:     "IPX-535 (2020-09-13)",
		},
		{
			name:     "Date format custom",
			template: "<ID> (<RELEASEDATE:YYYY-MM>)",
			want:     "IPX-535 (2020-09)",
		},
		{
			name:     "Actresses with default delimiter",
			template: "<ID> - <ACTORS>",
			want:     "IPX-535 - Sakura Momo, Test Actress",
		},
		{
			name:     "Actresses with custom delimiter",
			template: "<ID> - <ACTORS: and >",
			want:     "IPX-535 - Sakura Momo and Test Actress",
		},
		{
			name:     "Genres",
			template: "<ID> (<GENRES>)",
			want:     "IPX-535 (Genre1, Genre2, Genre3)",
		},
		{
			name:     "Director and label",
			template: "<DIRECTOR> - <LABEL>",
			want:     "Test Director - Test Label",
		},
		{
			name:     "Multiple tags",
			template: "<YEAR>/<STUDIO>/<ID> - <TITLE>",
			want:     "2020/Test Studio/IPX-535 - Test Movie Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Execute(tt.template, ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Forward slash",
			input: "Test/Movie",
			want:  "Test-Movie",
		},
		{
			name:  "Backslash",
			input: "Test\\Movie",
			want:  "Test-Movie",
		},
		{
			name:  "Colon",
			input: "Test: Movie",
			want:  "Test - Movie",
		},
		{
			name:  "Question mark",
			input: "Test? Movie",
			want:  "Test Movie",
		},
		{
			name:  "Asterisk",
			input: "Test* Movie",
			want:  "Test Movie",
		},
		{
			name:  "Quotes",
			input: `Test "Movie"`,
			want:  "Test 'Movie'",
		},
		{
			name:  "Angle brackets",
			input: "Test<Movie>",
			want:  "Test(Movie)",
		},
		{
			name:  "Pipe",
			input: "Test|Movie",
			want:  "Test-Movie",
		},
		{
			name:  "Multiple spaces",
			input: "Test    Movie",
			want:  "Test Movie",
		},
		{
			name:  "Trailing spaces",
			input: "Test Movie  ",
			want:  "Test Movie",
		},
		{
			name:  "Trailing dots",
			input: "Test Movie...",
			want:  "Test Movie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_IndexFormatting(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		index    int
		want     string
	}{
		{
			name:     "No padding",
			template: "fanart<INDEX>.jpg",
			index:    5,
			want:     "fanart5.jpg",
		},
		{
			name:     "Padding 2",
			template: "fanart<INDEX:2>.jpg",
			index:    5,
			want:     "fanart05.jpg",
		},
		{
			name:     "Padding 3",
			template: "fanart<INDEX:3>.jpg",
			index:    5,
			want:     "fanart005.jpg",
		},
		{
			name:     "Padding 2 with 10",
			template: "fanart<INDEX:2>.jpg",
			index:    10,
			want:     "fanart10.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{Index: tt.index}
			got, err := engine.Execute(tt.template, ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_Conditionals(t *testing.T) {
	engine := NewEngine()
	releaseDate := time.Date(2020, 9, 13, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		template string
		ctx      *Context
		want     string
	}{
		{
			name:     "Simple conditional with value",
			template: "<ID><IF:SERIES> - <SERIES></IF>",
			ctx: &Context{
				ID:     "IPX-535",
				Series: "Test Series",
			},
			want: "IPX-535 - Test Series",
		},
		{
			name:     "Simple conditional without value",
			template: "<ID><IF:SERIES> - <SERIES></IF>",
			ctx: &Context{
				ID:     "IPX-535",
				Series: "",
			},
			want: "IPX-535",
		},
		{
			name:     "Conditional with ELSE - true branch",
			template: "<ID><IF:DIRECTOR> by <DIRECTOR><ELSE> (No Director)</IF>",
			ctx: &Context{
				ID:       "IPX-535",
				Director: "Test Director",
			},
			want: "IPX-535 by Test Director",
		},
		{
			name:     "Conditional with ELSE - false branch",
			template: "<ID><IF:DIRECTOR> by <DIRECTOR><ELSE> (No Director)</IF>",
			ctx: &Context{
				ID:       "IPX-535",
				Director: "",
			},
			want: "IPX-535 (No Director)",
		},
		{
			name:     "Multiple conditionals",
			template: "<ID><IF:SERIES> - <SERIES></IF><IF:LABEL> [<LABEL>]</IF>",
			ctx: &Context{
				ID:     "IPX-535",
				Series: "Test Series",
				Label:  "Test Label",
			},
			want: "IPX-535 - Test Series [Test Label]",
		},
		{
			name:     "Multiple conditionals - partial values",
			template: "<ID><IF:SERIES> - <SERIES></IF><IF:LABEL> [<LABEL>]</IF>",
			ctx: &Context{
				ID:     "IPX-535",
				Series: "Test Series",
				Label:  "",
			},
			want: "IPX-535 - Test Series",
		},
		{
			name:     "Conditional with year",
			template: "<ID><IF:YEAR> (<YEAR>)</IF>",
			ctx: &Context{
				ID:          "IPX-535",
				ReleaseDate: &releaseDate,
			},
			want: "IPX-535 (2020)",
		},
		{
			name:     "Conditional with year - no date",
			template: "<ID><IF:YEAR> (<YEAR>)</IF>",
			ctx: &Context{
				ID:          "IPX-535",
				ReleaseDate: nil,
			},
			want: "IPX-535",
		},
		{
			name:     "Complex conditional with multiple tags",
			template: "<IF:DIRECTOR>Director: <DIRECTOR> | Studio: <STUDIO><ELSE>Studio: <STUDIO></IF>",
			ctx: &Context{
				Director: "John Doe",
				Maker:    "Test Studio",
			},
			want: "Director: John Doe | Studio: Test Studio",
		},
		{
			name:     "Complex conditional - false branch",
			template: "<IF:DIRECTOR>Director: <DIRECTOR> | Studio: <STUDIO><ELSE>Studio: <STUDIO></IF>",
			ctx: &Context{
				Director: "",
				Maker:    "Test Studio",
			},
			want: "Studio: Test Studio",
		},
		{
			name:     "Array conditional - actresses",
			template: "<ID><IF:ACTRESSES> starring <ACTRESSES></IF>",
			ctx: &Context{
				ID:        "IPX-535",
				Actresses: []string{"Actress 1", "Actress 2"},
			},
			want: "IPX-535 starring Actress 1, Actress 2",
		},
		{
			name:     "Array conditional - empty",
			template: "<ID><IF:ACTRESSES> starring <ACTRESSES></IF>",
			ctx: &Context{
				ID:        "IPX-535",
				Actresses: []string{},
			},
			want: "IPX-535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Execute(tt.template, tt.ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Execute() = %q, want %q", got, tt.want)
			}
		})
	}
}
