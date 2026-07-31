package import (
	"testing"
)

func TestConcurrentParser_ParseConcurrentBlock(t *testing.T) {
	parser := NewConcurrentParser()

	code := `concurrent {
    async fetchUserData();
    async fetchOrderHistory();
}`

	ast, err := parser.ParseConcurrentBlock(code)
	if err != nil {
		t.Fatalf("ParseConcurrentBlock failed: %v", err)
	}

	if len(ast.Tasks) != 2 {
		t.Fatalf("expected 2 async tasks parsed, got %d", len(ast.Tasks))
	}

	if ast.Tasks[0].CallExpr != "fetchUserData()" || ast.Tasks[1].CallExpr != "fetchOrderHistory()" {
		t.Errorf("unexpected call expressions: %+v", ast.Tasks)
	}
}
