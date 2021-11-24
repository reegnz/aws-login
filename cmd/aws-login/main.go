package main

import (
	"context"
	"log"
	"time"

	awslogin "github.com/reegnz/aws-login"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	login := awslogin.NewAWSLogin(ctx)
	u, err := login.LoginURL()
	if err != nil {
		log.Fatal(err)
	}
	open.Run(u)
}
