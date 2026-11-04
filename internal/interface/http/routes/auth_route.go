package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/handlers"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/middleware"
)

func AuthRoutes(r *gin.RouterGroup, authHandler handlers.AuthHandler) {
	api := r.Group("/auth")
	{
		//auth
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/refresh-token", authHandler.RefreshToken)
		api.GET("/oauth/github/url", authHandler.GetGitHubAuthURL)
		api.POST("/oauth/github", authHandler.GitHubOAuthCallback)
		api.GET("/oauth/github/callback", authHandler.GitHubOAuthCallback)

		// GitHub Repository & Issue operations (Authenticated)
		secretKey := "secret"
		if authHandler.AuthService.GetConfig() != nil && authHandler.AuthService.GetConfig().Server.Jwt.SecretKey != "" {
			secretKey = authHandler.AuthService.GetConfig().Server.Jwt.SecretKey
		}
		authProtected := api.Group("", middleware.AuthMiddleware(secretKey))
		{
			authProtected.GET("/oauth/github/repos", authHandler.GetGitHubRepositories)
			authProtected.POST("/oauth/github/repos/:owner/:repo/issues", authHandler.CreateGitHubIssue)
		}

		api.GET("/actions", authHandler.ListActions)
		api.GET("/resources", authHandler.ListResources)

		// policies
		api.GET("/policies", authHandler.ListPolicies)
		api.POST("/policies", authHandler.CreatePolicy)
		api.GET("/policies/:id", authHandler.GetPolicyById)
		api.PUT("/policies/:id", authHandler.UpdatePolicyById)
		api.DELETE("/policies/:id", authHandler.DeletePolicyById)

		// permissions
		api.GET("/permissions", authHandler.ListPermissions)
		api.POST("/permissions", authHandler.CreatePermission)
		api.GET("/permissions/:id", authHandler.GetPermissionById)
		api.PUT("/permissions/:id", authHandler.UpdatePermissionById)
		api.DELETE("/permissions/:id", authHandler.DeletePermissionById)

		// roles
		api.GET("/roles", authHandler.ListRoles)
		api.POST("/roles", authHandler.CreateRole)
		api.PUT("/roles/:id", authHandler.UpdateRoleById)
		api.PATCH("/roles/:id/status", authHandler.UpdateRoleStatus)
		api.PATCH("/roles/:id/deactivate", authHandler.DeactivateRoleById)
		api.DELETE("/roles/:id", authHandler.DeleteRoleById)

		// users
		api.GET("/users", authHandler.ListUsers)
		api.POST("/users", authHandler.CreateUser)
		api.GET("/users/:id", authHandler.GetUserById)
		api.PUT("/users/:id", authHandler.UpdateUserById)
		api.DELETE("/users/:id", authHandler.DeleteUserById)
	}
}
