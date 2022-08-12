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
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	userHashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panicln(err)
	}
	return string(userHashPassword)

}

func verifyPassword(userPassword string, providedHashedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword(
		[]byte(providedHashedPassword),
		[]byte(userPassword),
	)
	msg := ""
	check := true
	if err != nil {
		msg = fmt.Sprintf("Email or password is incorrrect")
		check = false
		return check, msg
	}
	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		//convert to what golang understands
		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			log.Panicln(err.Error())
			return
		}
		// validating
		validatorErr := validate.Struct(user)
		if validatorErr != nil {
			c.JSON(400, gin.H{"error": validatorErr.Error()})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Error occurred while checking email"})
			log.Panicln(err.Error())
			return
		}
		if count > 0 {
			c.JSON(400, gin.H{"error": "Email already exists"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Error occurred while checking phone"})
			log.Panicln(err.Error())
			return
		}
		if count > 0 {
			c.JSON(400, gin.H{"error": "Phone number already exists"})
			log.Panicln(err.Error())
			return
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
			*&user.UserId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			log.Panicln(err.Error())
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

		passwordIsValid, msg := verifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(500, gin.H{"error": msg})
			return
		}
		if foundUser.Email == nil {
			c.JSON(500, gin.H{"error": "user not found"})
			return
		}
		token, refreshToken, err := helper.GenerateAllTokens(
			*foundUser.Email,
			*foundUser.FirstName,
			*foundUser.LastName,
			*foundUser.UserType,
			foundUser.UserId)
		if err != nil {
			c.JSON(500, gin.H{"Error": "Error generating token"})
			return
		}
		log.Println(*&foundUser.UserId, foundUser.UserId)
		log.Println(token, refreshToken, foundUser.Email, foundUser.FirstName, foundUser.LastName)
		helper.UpdateAllTokens(token, refreshToken, foundUser.UserId)
		err = userCollection.FindOne(ctx, bson.M{"userid": foundUser.UserId}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, foundUser)
	}

}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"userid": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Println(userId, c.GetString("uid"))
		c.JSON(http.StatusOK, user)
	}
}
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(400, gin.H{"error": "Unauthorized access "})
		}
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))
		matchStage := bson.D{{"$match", bson.D{{}}}}
		// ordering of the users
		groupStage := bson.D{
			{"$group", bson.D{
				// ordering user by id
				{"_id", bson.D{
					{"_id", "null"},
				}},
				// summing the unique users together
				{"total_count", bson.D{
					{"$sum", 1},
				}},
				// append all users to return
				{"data", bson.D{
					{"$push", "$$ROOT"},
				}},
			}},
		}
		// fields to be shown in the frontend
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{
					{"$slice", []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"error": "error occurred while listing user items"})
		}
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(200, allUsers[0])

	}
}
