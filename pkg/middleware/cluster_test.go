package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReadClusterNameCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  ClusterNameHeader,
		Value: "%E7%94%9F%E4%BA%A7%E9%9B%86%E7%BE%A4",
	})
	ctx.Request = req

	if got := ReadClusterNameCookie(ctx); got != "生产集群" {
		t.Fatalf("ReadClusterNameCookie() = %q, want %q", got, "生产集群")
	}
}
