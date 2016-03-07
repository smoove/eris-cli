package util

import (
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/golang.org/x/oauth2"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/digitalocean/godo"
)

func ExampleWaitForActive() {
	// build client
	pat := "mytoken"
	token := &oauth2.Token{AccessToken: pat}
	t := oauth2.StaticTokenSource(token)

	oauthClient := oauth2.NewClient(oauth2.NoContext, t)
	client := godo.NewClient(oauthClient)

	// create your droplet and retrieve the create action uri
	uri := "https://api.digitalocean.com/v2/actions/xxxxxxxx"

	// block until until the action is complete
	err := WaitForActive(client, uri)
	if err != nil {
		panic(err)
	}
}
