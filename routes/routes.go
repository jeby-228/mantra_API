package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"mantra_API/auth"
	"mantra_API/controllers"
	"mantra_API/graphql"

	"github.com/gin-gonic/gin"
)

// isIntrospectionQuery checks whether the GraphQL request body is a
// schema-introspection query (e.g. sent by frontend tooling to fetch
// the schema). It returns true only when the top-level query string
// targets __schema or __type and nothing else.
func isIntrospectionQuery(body []byte) bool {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	q := strings.TrimSpace(req.Query)
	if q == "" {
		return false
	}
	// Introspection queries start with "query IntrospectionQuery" or
	// directly with "{ __schema" / "query { __schema".
	// A reliable heuristic: the query must reference __schema or __type
	// and must NOT contain any mutation keyword.
	hasIntrospection := strings.Contains(q, "__schema") || strings.Contains(q, "__type")
	hasMutation := strings.HasPrefix(q, "mutation")
	return hasIntrospection && !hasMutation
}

// SetupRouter registers API routes on the provided Gin engine.
func SetupRouter(router *gin.Engine) {
	// Public - no authentication required
	public := router.Group("/api/v1")
	{
		// Authentication-related routes
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.POST("/auth/line", controllers.LineLogin)
	}

	// GraphQL：GET 為 Playground（無需 JWT），其餘（POST 等）需 JWT
	// Introspection query（__schema / __type）也無需 JWT，方便前端開發工具取得 schema
	router.Any("/graphql", func(c *gin.Context) {
		graphqlHandler := graphql.GetHandler()
		if graphqlHandler == nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "GraphQL handler not initialized"},
			)
			return
		}

		needsAuth := c.Request.Method != http.MethodGet && c.Request.Method != http.MethodOptions

		if needsAuth && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if isIntrospectionQuery(bodyBytes) {
					needsAuth = false
				}
			}
		}

		if needsAuth {
			auth.AuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
		}
		graphqlHandler.ServeHTTP(c.Writer, c.Request)
	})

	// Protected routes - require authentication
	protected := router.Group("/api/v1")
	protected.Use(auth.AuthMiddleware()) // Add authentication middleware
	{
		// LINE Bind/Unbind
		protected.POST("/auth/line/bind", controllers.BindLine)
		protected.POST("/auth/line/unbind", controllers.UnbindLine)
	}
}
