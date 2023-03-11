package main

import (
	"os"

	"github.com/bpalermo/cert-manager-webhook-ns1/pkg/solver"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

var groupName = os.Getenv("GROUP_NAME")

func main() {
	if groupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our NS1 DNS provider with the webhook serving
	// library, making it available as an API under the provided groupName.
	cmd.RunWebhookServer(groupName, &solver.Ns1DNSProviderSolver{})
}
