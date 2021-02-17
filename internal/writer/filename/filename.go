package filename

import (
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// Strategy is a file naming strategy for generating names of files where a type should be code gen into
type Strategy interface {
	// GenerateFileName generates the filename (without an extension) for the given type name
	GenerateFileName(typeName string) string
}

// snakeCaseFileNameStrategy is a file naming strategy that snake-cases the type name
type snakeCaseFileNameStrategy struct {
}

// GenerateFileName generates the filename (without an extension) in snake case
func (s *snakeCaseFileNameStrategy) GenerateFileName(typeName string) string {
	// Taken from: https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
	snake := matchFirstCap.ReplaceAllString(typeName, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// SnakeCaseStrategy is a file naming strategy that uses snake case
func SnakeCaseStrategy() Strategy {
	return &snakeCaseFileNameStrategy{}
}
