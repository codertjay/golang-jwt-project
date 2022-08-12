package helpers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"os"
)

var SECRET_KEY = os.Getenv("SECRET_KEY")

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource")
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	// we have actually set the user_type in the authentication and the uuid
	// middleware, so we just get it by the key we use the token to get the user in there
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil
	// verify if the user id is this with the token he provided and also the
	// user is a customer user not an admin
	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource")
	}
	err = CheckUserType(c, userType)
	return err
}
