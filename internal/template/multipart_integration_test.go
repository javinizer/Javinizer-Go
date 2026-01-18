package template_test

import (
	"fmt"
	"testing"

	"github.com/javinizer/javinizer-go/internal/template"
	"github.com/stretchr/testify/assert"
)

func TestMultipartConditionalTemplate(t *testing.T) {
	engine := template.NewEngine()

	tests := []struct {
		name        string
		template    string
		isMultiPart bool
		partNumber  int
		partSuffix  string
		want        string
	}{
		{
			name:        "single file - no multipart suffix",
			template:    "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			isMultiPart: false,
			partNumber:  0,
			partSuffix:  "",
			want:        "STSK-074-poster.jpg",
		},
		{
			name:        "multipart pt1",
			template:    "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			isMultiPart: true,
			partNumber:  1,
			partSuffix:  "-pt1",
			want:        "STSK-074-pt1-poster.jpg",
		},
		{
			name:        "multipart pt2",
			template:    "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			isMultiPart: true,
			partNumber:  2,
			partSuffix:  "-pt2",
			want:        "STSK-074-pt2-poster.jpg",
		},
		{
			name:        "fanart format single file",
			template:    "<ID><IF:MULTIPART>-pt<PART></IF>-fanart.jpg",
			isMultiPart: false,
			partNumber:  0,
			partSuffix:  "",
			want:        "STSK-074-fanart.jpg",
		},
		{
			name:        "fanart format multipart",
			template:    "<ID><IF:MULTIPART>-pt<PART></IF>-fanart.jpg",
			isMultiPart: true,
			partNumber:  1,
			partSuffix:  "-pt1",
			want:        "STSK-074-pt1-fanart.jpg",
		},
		{
			name:        "using PARTSUFFIX instead of -pt<PART>",
			template:    "<ID><IF:MULTIPART><PARTSUFFIX></IF>-poster.jpg",
			isMultiPart: true,
			partNumber:  1,
			partSuffix:  "-cd1",
			want:        "STSK-074-cd1-poster.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &template.Context{
				ID:          "STSK-074",
				IsMultiPart: tt.isMultiPart,
				PartNumber:  tt.partNumber,
				PartSuffix:  tt.partSuffix,
			}

			result, err := engine.Execute(tt.template, ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)

			// Also print for debugging
			fmt.Printf("Template: %s, IsMultiPart: %v, PartNumber: %d -> Result: %s\n",
				tt.template, tt.isMultiPart, tt.partNumber, result)
		})
	}
}
