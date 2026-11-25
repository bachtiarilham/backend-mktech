package delivery_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchitecture_UseCaseIsolation(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		normalized := filepath.ToSlash(path)
		if !strings.Contains(normalized, "/usecase/") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath == "net/http" || strings.HasPrefix(importPath, "github.com/gin-gonic/gin") {
				t.Errorf("usecase %s must not import HTTP package %s", normalized, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk internal directory: %v", err)
	}
}

func TestArchitecture_RepositoryIsolation(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		normalized := filepath.ToSlash(path)
		if !strings.Contains(normalized, "/repository/") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(importPath, "ptmuda.tech/backend/internal/") && strings.HasSuffix(importPath, "/repository") {
				t.Errorf("repository %s must not import another repository %s", normalized, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk internal directory: %v", err)
	}
}
