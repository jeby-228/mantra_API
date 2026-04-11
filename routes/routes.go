package routes

import (
	"net/http"

	"mantra_API/auth"
	"mantra_API/controllers"
	"mantra_API/graphql"

	"github.com/gin-gonic/gin"
)

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
	router.Any("/graphql", func(c *gin.Context) {
		graphqlHandler := graphql.GetHandler()
		if graphqlHandler == nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "GraphQL handler not initialized"},
			)
			return
		}
		// OPTIONS 為 CORS 預檢，不可要求 JWT
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodOptions {
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
