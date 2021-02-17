package writer

import (
	"github.com/csueiras/reinforcer/internal/generator"
	"github.com/csueiras/reinforcer/internal/writer/filename"
	wio "github.com/csueiras/reinforcer/internal/writer/io"
	"path"
)

// Writer is responsible for unloading the generated code into the output
type Writer struct {
	fileNameStrategy filename.Strategy
	outputProvider   wio.OutputProvider
}

// New is a constructor for Writer
func New(outputProvider wio.OutputProvider, fileNameStrategy filename.Strategy) *Writer {
	return &Writer{
		fileNameStrategy: fileNameStrategy,
		outputProvider:   outputProvider,
	}
}

// Default creates the default provider that writes to the local file system and uses snake case file naming strategy
func Default() *Writer {
	return New(wio.NewFSOutputProvider(), filename.SnakeCaseStrategy())
}

// Write saves the generated contents to the given output location
func (w *Writer) Write(outputDirectory string, generated *generator.Generated) error {
	if err := w.writeTo(path.Join(outputDirectory, "reinforcer_common.go"), generated.Common); err != nil {
		return err
	}

	if err := w.writeTo(path.Join(outputDirectory, "reinforcer_constants.go"), generated.Constants); err != nil {
		return err
	}

	for _, codegen := range generated.Files {
		filePath := path.Join(outputDirectory, w.fileNameStrategy.GenerateFileName(codegen.TypeName)+".go")
		if err := w.writeTo(filePath, codegen.Contents); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writeTo(target string, contents string) error {
	writeTarget, err := w.outputProvider.GetOutputTarget(target)
	if err != nil {
		return err
	}
	defer func() {
		_ = writeTarget.Close()
	}()
	if _, err = writeTarget.Write([]byte(contents)); err != nil {
		return err
	}
	return nil
}
