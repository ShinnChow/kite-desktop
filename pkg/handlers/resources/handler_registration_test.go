package resources

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesIncludesResourceQuotas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	RegisterRoutes(api)

	handler, err := GetHandler("resourcequotas")
	if err != nil {
		t.Fatalf("expected resourcequotas handler to be registered: %v", err)
	}
	if handler.IsClusterScoped() {
		t.Fatalf("expected resourcequotas to be namespace scoped")
	}
}
