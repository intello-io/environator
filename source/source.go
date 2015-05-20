package source

import (
	//"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/heroku-go"
	"io"
	// "io/ioutil"
	"os/user"
	"path/filepath"
	"text/template"
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
}

func (self *Source) sourceHeroku(name string) (string, error) {
	if self.herokuClient == nil {
		usr, err := user.Current()

		if err != nil {
			return "", errors.New(fmt.Sprintf("Could not get the current user: %s", err))
		}

		config, err := netrc.ParseFile(filepath.Join(usr.HomeDir, ".netrc"))

		if err != nil {
			return "", errors.New(fmt.Sprintf("Could not parse ~/.netrc: %s\n", err))
		}

		machineConfig := config.FindMachine("api.heroku.com")

		if machineConfig == nil {
			return "", errors.New(fmt.Sprintf("No entry found for api.heroku.com in ~/.netrc. Please run `heroku login` first."))
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
			buffer.WriteString(k)
			buffer.WriteRune('=')
			buffer.WriteString(v)
			buffer.WriteRune('\n')
		}
	}

	return buffer.String(), nil
}

func (self *Source) Execute(w io.Writer, name string, params interface{}) error {
	funcMap := template.FuncMap{
		"source": self.ExecuteString,
		"heroku": self.sourceHeroku,
	}

	tmpl, err := template.New(fmt.Sprintf("%s.env", name)).Funcs(funcMap).ParseFiles(filepath.Join("env", fmt.Sprintf("%s.env", name)))

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
