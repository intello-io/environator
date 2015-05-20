package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dailymuse/environator/source"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func exit(message string) {
	os.Stderr.WriteString(message)
	os.Exit(1)
}

func run(debug bool, dir string, profile string, args []string) error {
	// Disable logging if debug isn't enabled
	if !debug {
		log.SetOutput(ioutil.Discard)
	}

	// Create a temporary file to store the compiled template results in
	tempVarsFile, err := ioutil.TempFile("", "environator-")

	if err != nil {
		return err
	}

	if debug {
		log.Printf("The varsfile will not be discarded due to debug mode: %s\n", tempVarsFile.Name())
	} else {
		defer os.Remove(tempVarsFile.Name())
	}

	src := source.Source{}

	err = src.Execute(tempVarsFile, profile, map[string]interface{}{
		"debug":  debug,
		"dir":    dir,
		"source": profile,
		"cmd":    args,
	})

	closeErr := tempVarsFile.Close()

	if err != nil {
		return err
	} else if closeErr != nil {
		return closeErr
	}

	var bashString bytes.Buffer
	bashString.WriteString(fmt.Sprintf("set -e; set -a; source %s; export ENVIRONATOR_PROFILE=%s; ", tempVarsFile.Name(), profile))

	if dir != "" && dir != "." {
		bashString.WriteString(fmt.Sprintf("cd %s; ", dir))
	}

	bashString.WriteString(strings.Join(args, " "))

	log.Printf("Running: %s\n", bashString.String())

	// Create the command
	cmd := exec.Command("bash", "-c", bashString.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err = cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func main() {
	var debug = flag.Bool("debug", false, "Enable debug mode")
	var dir = flag.String("d", ".", "Sets the working directory")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		exit("Need an argument at position 0 to specify the profile\n")
	} else if len(args) < 2 {
		exit("Need one or more arguments starting at position 1 to specify the command and arguments to run\n")
	}

	if err := run(*debug, *dir, args[0], args[1:]); err != nil {
		exit(fmt.Sprintf("Error: %s\n", err.Error()))
	}
}
