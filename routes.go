// routes.go

package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var sess *session.Session

func initializeRoutes() {

	// Use the setUserStatus middleware for every route to set a flag
	// indicating whether the request was from an authenticated user or not
	router.Use(setUserStatus())

	// Create AWS Session
	conf := &aws.Config{Region: aws.String("eu-west-1")}
	sess, err := session.NewSession(conf)
	if err != nil {
		panic(err)
	}

	// Get value of Cognito Application Client ID
	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion("eu-west-1"))

	ClientIDKey := "CognitoAppClientID"
	withDecryption := true
	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &ClientIDKey,
		WithDecryption: &withDecryption,
	})

	// ClientIDValue stores the value of Cognito Application Client ID
	ClientIDValue := *param.Parameter.Value
	fmt.Println(ClientIDValue)

	// Define object with Cognito Stuff
	ce := CognitoExample{
		CognitoClient: cognito.New(sess),
		RegFlow:       &regFlow{},
		UserPoolID:    "CognitoUnicornUserPool",
		AppClientID:   ClientIDValue,
	}

	// Create DynamoDB Service Session and reference the session
	ddbsvc := dynamodb.New(sess)

	// load info from DynamoDB to Projects array in Memory
	loadProjectsDynamoDB(ddbsvc)

	// Create DynamoDB Service Session for Users Table
	usrsvc := dynamodb.New(sess)

	// Handle the index route
	router.GET("/", showIndexPage)

	//Group Global routes (About, Leaderboard)
	globalRoutes := router.Group("/g")
	{
		// Handle the GET requests at /g/about
		globalRoutes.GET("/about", showAboutPage)

		// Handle POST requests at /g/leaderboard
		// Ensure that the user is logged in by using the middleware
		globalRoutes.GET("/leaderboard", ensureLoggedIn(), showLeaderboardPage)
	}

	// Group User routes together (Login, Register)
	userRoutes := router.Group("/u")
	{
		// Handle the GET requests at /u/login
		// Show the login page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/login", ensureNotLoggedIn(), showLoginPage)

		// Handle POST requests at /u/login
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/login", ensureNotLoggedIn(), performLogin(ce))

		// Handle GET requests at /u/logout
		// Ensure that the user is logged in by using the middleware
		userRoutes.GET("/logout", ensureLoggedIn(), logout)

		// Handle the GET requests at /u/register
		// Show the registration page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/register", ensureNotLoggedIn(), showRegistrationPage)

		// Handle POST requests at /u/register
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/register", ensureNotLoggedIn(), register(ce))

		// Handle the GET requests at /u/otp
		// Show the login page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/otp", ensureNotLoggedIn(), showOTPPage)

		// Handle POST requests at /u/otp
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/otp", ensureNotLoggedIn(), OTP(ce))
	}

	// Group Project routes (View, Create)
	projectRoutes := router.Group("/project")
	{
		// Handle GET requests at /project/view/project_id
		// This is where VOTING needs to happen
		projectRoutes.GET("/view/:project_id", ensureLoggedIn(), getProject)

		// Handle POST requests at /project/view/project_id
		// Ensure that the user is logged in by using the middleware
		projectRoutes.POST("/view/:project_id", ensureLoggedIn(), voteForProject(ddbsvc, usrsvc))

		// Handle the GET requests at /project/create
		// Show the project creation page
		// Ensure that the user is logged in by using the middleware
		projectRoutes.GET("/create", ensureLoggedIn(), showProjectCreationPage)

		// Handle POST requests at /project/create
		// Ensure that the user is logged in by using the middleware
		projectRoutes.POST("/create", ensureLoggedIn(), createProject(ddbsvc, sess))

		// Handle the GET requests at /project/votes
		// Show the page with all the votes
		// Ensure that the user is logged in by using the middleware
		projectRoutes.GET("/votes", ensureLoggedIn(), showUserVotes(usrsvc))

	}
}
