package vis

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestPackageImportsNoDisplay is the rule that makes this package worth having
// separate from internal/gui: it draws into a plain framebuffer, so the live
// visualiser window and the headless documentation recordings can share it.
// Importing a display here would put Ebiten back in tools/demogen's build,
// which cannot run without one.
func TestPackageImportsNoDisplay(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Clean(name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		checked++
		for _, imp := range f.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatalf("%s: bad import %s", name, imp.Path.Value)
			}
			if strings.Contains(path, "hajimehoshi") {
				t.Errorf("%s imports a display (%s); this package must stay display-free", name, path)
			}
			if strings.HasSuffix(path, "/internal/gui") {
				t.Errorf("%s imports internal/gui, which links a display", name)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no source files checked; the rule is not actually being enforced")
	}
}
