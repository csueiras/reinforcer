package filename_test

import (
	"github.com/csueiras/reinforcer/internal/writer/filename"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSnakeCaseFileNameStrategy_GenerateFileName(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{
			name:     "Default",
			typeName: "HelloWorldService",
			want:     "hello_world_service",
		},
		{
			name:     "All Caps",
			typeName: "HTML",
			want:     "html",
		},
		{
			name:     "Abbrv. all caps and multiple words",
			typeName: "SQSEventHandler",
			want:     "sqs_event_handler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := filename.SnakeCaseStrategy()
			got := s.GenerateFileName(tt.typeName)
			require.Equal(t, tt.want, got)
		})
	}
}
