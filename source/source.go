package source

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/heroku-go"
	vaultapi "github.com/hashicorp/vault/api"
)

var (
	BLACKLISTED_HEROKU_VARS = map[string]bool{
		"LANG":             true,
		"LD_LIBRARY_PATH":  true,
		"LIBRARY_PATH":     true,
		"PATH":             true,
		"PYTHONHASHSEED":   true,
		"PYTHONHOME":       true,
		"PYTHONPATH":       true,
		"PYTHONUNBUFFERED": true,
	}
)

type Source struct {
	herokuClient *heroku.Client
	vaultClient  *vaultapi.Client
}

func (self *Source) sourceHeroku(name string) (string, error) {
	if self.herokuClient == nil {
		usr, err := user.Current()

		if err != nil {
			return "", fmt.Errorf("Could not get the current heroku user: %s", err)
		}

		config, err := netrc.ParseFile(filepath.Join(usr.HomeDir, ".netrc"))

		if err != nil {
			return "", fmt.Errorf("Could not parse ~/.netrc: %s\n", err)
		}

		machineConfig := config.FindMachine("api.heroku.com")

		if machineConfig == nil {
			return "", fmt.Errorf("No entry found for api.heroku.com in ~/.netrc. Please run `heroku login` first.")
		}

		self.herokuClient = &heroku.Client{
			Username: machineConfig.Login,
			Password: machineConfig.Password,
		}
	}

	var buffer bytes.Buffer
	vars, err := self.herokuClient.ConfigVarInfo(name)

	if err != nil {
		return "", err
	}

	for k, v := range vars {
		_, ok := BLACKLISTED_HEROKU_VARS[k]

		if !ok {
			buffer.WriteString(fmt.Sprintf("%s='%s'\n", k, strings.Replace(v, "'", "\\'", -1)))
		}
	}

	return buffer.String(), nil
}

func (self *Source) sourceVault(path string) (string, error) {
	if self.vaultClient == nil {
		config := vaultapi.DefaultConfig()

		err := config.ReadEnvironment()

		if err != nil {
			return "", fmt.Errorf("Could not read environment variables into a vault config: %s", err)
		}

		client, err := vaultapi.NewClient(config)

		if err != nil {
			return "", fmt.Errorf("Could not connect to vault: %s", err)
		}

		self.vaultClient = client
	}

	// TODO: We currently pull all the keys, then we read each key one at a
	// time, resulting in a lot of API calls. Is there a way to read all of
	// the secrets in one API call?
	secrets, err := self.vaultClient.Logical().List(path)

	if err != nil {
		return "", fmt.Errorf("Could not read vault config: %s", err)
	}

	var buffer bytes.Buffer

	if secrets != nil {
		keys := secrets.Data["keys"].([]interface{})

		for _, key := range keys {
			secret, err := self.vaultClient.Logical().Read(fmt.Sprintf("%s/%s", path, key.(string)))

			if err != nil {
				return "", fmt.Errorf("Could not read vault config: %s", err)
			}

			if secret != nil {
				value := secret.Data["value"]
				fmt.Printf("%s=%s\n", key, value)
			}

		}
	}

	return buffer.String(), nil
}

func (self *Source) Execute(w io.Writer, name string, params interface{}) error {
	path := os.Getenv("ENVIRONATOR_PATH")

	if path == "" {
		path = "env"
	}

	filename := fmt.Sprintf("%s.env", name)
	fullpath := filepath.Join(path, filename)

	funcMap := template.FuncMap{
		"source": self.ExecuteString,
		"heroku": self.sourceHeroku,
		"vault":  self.sourceVault,
	}

	tmpl, err := template.New(filename).Funcs(funcMap).ParseFiles(fullpath)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, params)
}

func (self *Source) ExecuteString(name string, params interface{}) (string, error) {
	var buffer bytes.Buffer
	err := self.Execute(&buffer, name, params)

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
