package solver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmMetaV1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	ns1Rest "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sRest "k8s.io/client-go/rest"
	"net/http"
	"strings"
)

// Ns1DNSProviderSolver implements the logic needed to 'present' an ACME
// challenge TXT record. To do so, it implements the
// `github.com/cert-manager/cert-manager/pkg/acme/webhook.Solver` interface.
type Ns1DNSProviderSolver struct {
	k8sClient *kubernetes.Clientset
	ns1Client *ns1Rest.Client
}

// Ns1DNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
type ns1DNSProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.
	APIKeySecretRef cmMetaV1.SecretKeySelector `json:"apiKeySecretRef"`
	Endpoint        string                     `json:"endpoint"`
	IgnoreSSL       bool                       `json:"ignoreSSL"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
func (c *Ns1DNSProviderSolver) Name() string {
	return "ns1"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *Ns1DNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	zone, domain, err := c.parseChallenge(ch)
	if err != nil {
		return err
	}

	if c.ns1Client == nil {
		if err := c.setNS1Client(ch, cfg); err != nil {
			return err
		}
	}

	// Create a TXT Record for domain.zone with answer set to DNS challenge key
	// Short TTL is fine, as we delete the record after the challenge is solved.
	record := dns.NewRecord(zone, domain, "TXT")
	record.TTL = 600
	record.AddAnswer(dns.NewTXTAnswer(ch.Key))

	_, err = c.ns1Client.Records.Create(record)
	if err != nil {
		if err != ns1Rest.ErrRecordExists {
			return err
		}
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *Ns1DNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	zone, domain, err := c.parseChallenge(ch)
	if err != nil {
		return err
	}

	if c.ns1Client == nil {
		if err := c.setNS1Client(ch, cfg); err != nil {
			return err
		}
	}

	// Delete the TXT Record we created in Present
	if _, err = c.ns1Client.Records.Delete(
		zone, fmt.Sprintf("%s.%s", domain, zone), "TXT",
	); err != nil {
		return err
	}

	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *Ns1DNSProviderSolver) Initialize(kubeClientConfig *k8sRest.Config, _ <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	c.k8sClient = cl
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *apiExtensionsV1.JSON) (ns1DNSProviderConfig, error) {
	cfg := ns1DNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func (c *Ns1DNSProviderSolver) setNS1Client(ch *v1alpha1.ChallengeRequest, cfg ns1DNSProviderConfig) error {
	ref := cfg.APIKeySecretRef
	if ref.Name == "" {
		return fmt.Errorf(
			"secret for NS1 apiKey not found in '%s'",
			ch.ResourceNamespace,
		)
	}
	if ref.Key == "" {
		return fmt.Errorf(
			"no 'key' set in secret '%s/%s'",
			ch.ResourceNamespace,
			ref.Name,
		)
	}

	secret, err := c.k8sClient.CoreV1().Secrets(ch.ResourceNamespace).Get(
		context.Background(), ref.Name, k8sMetaV1.GetOptions{},
	)
	if err != nil {
		return err
	}
	apiKeyBytes, ok := secret.Data[ref.Key]
	if !ok {
		return fmt.Errorf(
			"no key '%s' in secret '%s/%s'",
			ref.Key,
			ch.ResourceNamespace,
			ref.Name,
		)
	}
	apiKey := string(apiKeyBytes)

	httpClient := &http.Client{}
	if cfg.IgnoreSSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = tr
	}
	c.ns1Client = ns1Rest.NewClient(
		httpClient,
		ns1Rest.SetAPIKey(apiKey),
		ns1Rest.SetEndpoint(cfg.Endpoint),
	)

	return nil
}

// Get the zone and domain we are setting from the challenge request
func (c *Ns1DNSProviderSolver) parseChallenge(ch *v1alpha1.ChallengeRequest) (
	zone string, domain string, err error,
) {

	if zone, err = util.FindZoneByFqdn(
		ch.ResolvedFQDN, util.RecursiveNameservers,
	); err != nil {
		return "", "", err
	}
	zone = util.UnFqdn(zone)

	if idx := strings.Index(ch.ResolvedFQDN, "."+ch.ResolvedZone); idx != -1 {
		domain = ch.ResolvedFQDN[:idx]
	} else {
		domain = util.UnFqdn(ch.ResolvedFQDN)
	}

	return zone, domain, nil
}
