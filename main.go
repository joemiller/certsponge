package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var version = "development"

const (
	sensitiveFileMode    = 0o600
	nonSensitiveFileMode = 0o644
)

type vaultPKIResponse struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`
	Data          struct {
		CaChain        []string `json:"ca_chain"`
		Certificate    string   `json:"certificate"`
		Expiration     int      `json:"expiration"`
		IssuingCa      string   `json:"issuing_ca"`
		PrivateKey     string   `json:"private_key"`
		PrivateKeyType string   `json:"private_key_type"`
		SerialNumber   string   `json:"serial_number"`
	} `json:"data"`
	Warnings interface{} `json:"warnings"`
}

func main() {
	err := run(os.Stdin, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func run(stdin io.Reader, args []string) error {
	var flagSet *flag.FlagSet = flag.NewFlagSet("certsponge", flag.ContinueOnError)

	bundleFile := ""
	caFile := ""
	keyFile := ""
	certFile := ""
	force := false
	printVersion := false

	flagSet.StringVar(&bundleFile, "bundle", "tls.pem", "Path to the PEM bundle file to write containing: private_key, certificate, and ca_chain. Set to \"\" to disable.")
	flagSet.StringVar(&caFile, "ca-cert", "ca.crt", "Path to the CA bundle file to write containing: ca_chain. Set to \"\" to disable.")
	flagSet.StringVar(&keyFile, "key", "", "Path to the file to write the private_key.")
	flagSet.StringVar(&certFile, "cert", "", "Path to the file to write the certificate.")
	flagSet.BoolVar(&force, "f", false, "Force overwriting of existing files.")
	flagSet.BoolVar(&printVersion, "v", false, "Print version and exit.")

	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	if printVersion {
		fmt.Println(version)
		return nil
	}

	input, err := ioutil.ReadAll(stdin)
	if err != nil {
		return err
	}

	resp := vaultPKIResponse{}
	err = json.Unmarshal(input, &resp)
	if err != nil {
		return err
	}

	if resp.Data.PrivateKey == "" || resp.Data.Certificate == "" {
		return errors.New("JSON input is missing data.private_key or data.certificate fields. Aborting")
	}

	if bundleFile != "" {
		if exists(bundleFile) && !force {
			return fmt.Errorf("file %s already exists and -f (force) flag not specified", bundleFile)
		}
		bundle := []string{
			resp.Data.PrivateKey,
			resp.Data.Certificate,
		}
		bundle = append(bundle, resp.Data.CaChain...)
		_, err = writeFile(bundleFile, strings.Join(bundle, "\n"), sensitiveFileMode)
		if err != nil {
			return err
		}
	}

	if caFile != "" {
		if exists(caFile) && !force {
			return fmt.Errorf("file %s already exists and -f (force) flag not specified", caFile)
		}
		_, err = writeFile(caFile, strings.Join(resp.Data.CaChain, "\n"), nonSensitiveFileMode)
		if err != nil {
			return err
		}
	}

	if keyFile != "" {
		if exists(keyFile) && !force {
			return fmt.Errorf("file %s already exists and -f (force) flag not specified", keyFile)
		}
		_, err = writeFile(keyFile, resp.Data.PrivateKey+"\n", sensitiveFileMode)
		if err != nil {
			return err
		}
	}

	if certFile != "" {
		if exists(certFile) && !force {
			return fmt.Errorf("file %s already exists and -f (force) flag not specified", certFile)
		}
		_, err = writeFile(certFile, resp.Data.Certificate+"\n", nonSensitiveFileMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func writeFile(filename string, data string, perm fs.FileMode) (int, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.WriteString(data)
}
