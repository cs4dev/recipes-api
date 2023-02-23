// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin.
//
//		Schemes: http
//	 	Host: localhost:8080
//		BasePath: /
//		Version: 1.0.0
//		Contact: CS Lim <cs4dev@gmail.com>
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
//
// swagger:meta
package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"

	"github.com/cs4dev/recipes-api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var recipesHandler *handlers.RecipesHandler
var authHandler *handlers.AuthHandler

func init() {
	ctx = context.Background()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionRecipes := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	recipesHandler = handlers.NewRecipesHandler(ctx, collectionRecipes, redisClient)
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)

}

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		tokenValue := c.GetHeader("Authorization")
// 		claims := &handlers.Claims{}

// 		tkn, err := jwt.ParseWithClaims(tokenValue, claims,
// 			func(token *jwt.Token) (interface{}, error) {
// 				return []byte(os.Getenv("JWT_SECRET")), nil
// 			})
// 		if err != nil {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}
// 		if tkn == nil || !tkn.Valid {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}
// 		c.Next()
// 	}
// }

func main() {
	router := gin.Default()

	store, _ := redisStore.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))

	router.Use(sessions.Sessions("recipes_api", store))

	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/signout", authHandler.SignOutHandler)

	authorized := router.Group("/")

	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.GET("/recipes", recipesHandler.ListRecipesHandler)
		authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
		authorized.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	}

	router.Run()
}
