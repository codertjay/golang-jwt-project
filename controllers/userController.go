package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang-jwt-project/database"
	helper "golang-jwt-project/helpers"
	"golang-jwt-project/models"
	"log"
	"time"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword() {

}

func varifyPassword(userPassword string, providedPassword string) bool {

}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}
		validatorErr := validate.Struct(user)
		if validatorErr != nil {
			c.JSON(400, gin.H{"error": validatorErr.Error()})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Error occurred while checking email"})
			log.Panicln(err)
			return
		}
		if count > 0 {
			c.JSON(400, gin.H{"error": "Email already exists"})
		}
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Error occurred while checking phone"})
			log.Panicln(err)
			return
		}
		if count > 0 {
			c.JSON(400, gin.H{"error": "Phone number already exists"})
		}
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserId = user.ID.Hex()
		token, refreshToken, err := helper.GenerateAllTokens(
			*user.Email,
			*user.FirstName,
			*user.LastName,
			*user.UserType,
			user.UserId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			log.Panicln(err)
			return
		}
		user.Token = &token
		user.RefreshToken = &refreshToken
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, &user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(500, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(200, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			log.Panicln(err)
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"error": "email or password is incorrect"})
			return
		}
		paswordIsValid, msg := varifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
	}

}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		if err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user); err != nil {
			c.JSON(500, gin.H{"error": "An error occurred finding user"})
			return
		}
		defer cancel()
		c.JSON(200, user)
	}
}
func GetUsers() {

}
