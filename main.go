package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/ktr0731/go-fuzzyfinder"
	"gopkg.in/ini.v1"
)

var stages = []string{"development", "staging"}

type Config struct {
	VaultAddress string
	VaultToken   string
}

type VaultClient struct {
	*vault.Client
	Config Config
}

func NewVaultClient(client *vault.Client, config Config) *VaultClient {
	vc := &VaultClient{Client: client, Config: config}
	vc.Client.SetToken(vc.Config.VaultToken)

	return vc
}

func (c *VaultClient) getSecrets(path string) (map[string]interface{}, error) {
	secret, err := c.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	m, err := parseSecret(secret)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *VaultClient) listRepos(path string) ([]string, error) {
	res, err := c.Client.Logical().List(path)
	if err != nil {
		return nil, err
	}

	data, ok := res.Data["keys"]
	if !ok {
		return nil, fmt.Errorf("Could not parse repo data")
	}

	si, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Could not parse repo data: Invalid data type")
	}

	ss := make([]string, len(si))
	for i, v := range si {
		ss[i] = v.(string)
	}

	return ss, nil
}

func parseSecret(secret *vault.Secret) (map[string]interface{}, error) {
	res, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("Could not parse secret")
	}

	m, ok := res.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Could not parse secret: Invalid data type")
	}

	return m, nil
}

func write(m map[string]interface{}) error {
	f, err := os.OpenFile(".env", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range m {
		if _, err := f.WriteString(fmt.Sprintf("%s=%v\n", k, v)); err != nil {
			return err
		}
	}

	return nil
}

func getChoice(choices []string) (string, error) {
	idx, err := fuzzyfinder.Find(choices, func(i int) string {
		return fmt.Sprintf("Stage: %v", choices[i])
	})
	if err != nil {
		return "", err
	}

	return choices[idx], nil
}

type Credentials struct {
	url   string
	token string
}

func newCredentials() *Credentials {
	home, _ := os.UserHomeDir()
	cfg, err := ini.Load(filepath.Join(home, ".vaultconf.ini"))
	if err != nil {
		// File does not exist, prompt user for credentials
		fmt.Print("Enter url: ")
		reader := bufio.NewReader(os.Stdin)
		url, _ := reader.ReadString('\n')
		url = strings.TrimSuffix(url, "\n")

		fmt.Print("Enter token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSuffix(token, "\n")

		// Create config file with entered credentials
		cfg = ini.Empty()
		cfg.Section("credentials").Key("url").SetValue(url)
		cfg.Section("credentials").Key("token").SetValue(token)
		cfg.SaveTo(filepath.Join(home, ".vaultconf.ini"))
	}
	url := cfg.Section("credentials").Key("url").String()
	token := cfg.Section("credentials").Key("token").String()

	return &Credentials{url: url, token: token}
}

func main() {
	creds := newCredentials()
	config := Config{
		VaultAddress: creds.url,
		VaultToken:   creds.token,
	}

	client, err := vault.NewClient(&vault.Config{
		Address: config.VaultAddress,
	})
	if err != nil {
		log.Fatal("Could not initialize vault client", err)
	}

	vc := NewVaultClient(client, config)

	repos, err := vc.listRepos("staging/metadata")
	if err != nil {
		log.Fatal("Could not get repos.\nError: ", err)
	}

	choiceS, err := getChoice(stages)
	if err != nil {
		log.Fatal("Error while choosing stage ", err)
	}

	choiceR, err := getChoice(repos)
	if err != nil {
		log.Fatal("Error while choosing repo ", err)
	}

	path := fmt.Sprintf("%s/data/%s", choiceS, choiceR)
	m, err := vc.getSecrets(path)
	if err != nil {
		log.Fatal("Could not get clients", err)
	}

	err = write(m)
	if err != nil {
		log.Fatal("Error while creating .env file")
	}
}
