package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type CA struct {
	Name     string  `json:"name"`
	Duration float64 `json:"duration,omitempty"` //in days
}

type Host struct {
	Hostname string   `json:"hostname"`
	IP       string   `json:"ip"` //with CIDR network suffix
	Groups   []string `json:"groups,omitempty"`
	Duration float64  `json:"duration,omitempty"` //in days
}

type Network struct {
	CA    CA     `json:"ca"`
	Hosts []Host `json:"hosts"`
}

func main() {
	var err error
	l := log.New(os.Stderr, "", 0) //set logging to standard error and no timestamp

	//commandline options
	caCertFile := flag.String("c", "./ca.crt", "CA certificate path.")
	caKeyFile := flag.String("k", "./ca.key", "CA key path.")
	binaryPath := flag.String("p", "", "Path to nebula-cert binary file. If not specified, search $PATH and current directory.")
	networkFile := flag.String("f", "-", "Path to network input file. Use '-' for standard input.")
	overwrite := flag.Bool("o", false, "Overwrite existing files.")
	flag.Parse()

	//Locate binary
	pathFailText := "Executable not found in $PATH or current directory. Specify the path with the '-p' option"
	if *binaryPath == "" {
		*binaryPath, err = exec.LookPath("nebula-cert")
		if err != nil {
			*binaryPath = "./nebula-cert"
		}
	}
	if _, err := os.Stat(*binaryPath); os.IsNotExist(err) { //check if file exists...
		if *binaryPath != "./nebula-cert" {
			pathFailText = "Executable not found at " + *binaryPath
		}
		l.Fatal(pathFailText)
	}

	var inputFile *os.File
	if *networkFile == "-" { //read input from stdin
		inputFile = os.Stdin
	} else { //read input from file
		l.Println("Processing network description file: " + *networkFile)
		inputFile, err = os.Open(*networkFile)
		if err != nil {
			l.Fatal("Could not open network description file: " + *networkFile + "\n" + err.Error())
		}
		defer inputFile.Close()
	}

	var input string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		input = input + scanner.Text()
	}

	var network Network

	err = json.Unmarshal([]byte(input), &network) //read the network config
	if err != nil {
		if *networkFile == "-" {
			*networkFile = "standard input."
		}
		l.Fatal("Could not parse network description from " + *networkFile + "\nError: " + err.Error())
	}

	var cmd *exec.Cmd

	//Create CA if name is specified, AND existing cert doesn't already exist OR overwrite is true.
	if len(network.CA.Name) > 0 {
		if _, err := os.Stat(*caCertFile); os.IsNotExist(err) || *overwrite {
			duration := "8760h" //default 1 year
			if network.CA.Duration > 0 {
				duration = strconv.Itoa(int(math.Round(network.CA.Duration*24))) + "h" //convert days to hours
			}
			cmd := exec.Command(*binaryPath, "ca", "-out-crt", *caCertFile, "-out-key", *caKeyFile, "-name", network.CA.Name, "-duration", duration)
			output, err := cmd.CombinedOutput()
			if err != nil {
				l.Fatal("CA: " + string(output) + " Error: " + err.Error())
			}
			l.Println("Created CA '" + network.CA.Name + "' OK " + string(output))
		} else {
			l.Println("CA certificate '" + *caCertFile + "' already exists. Skipping...")
		}
	}

	for _, h := range network.Hosts {
		if _, err := os.Stat(h.Hostname + ".crt"); err == nil && !*overwrite { //check if host certificate file exists and overwrite not true
			l.Println(h.Hostname + " certificate already exists. Skipping...")
			continue
		}
		groups := strings.Join(h.Groups, ",")
		if h.Duration > 0 {
			duration := strconv.Itoa(int(math.Round(h.Duration*24))) + "h"
			cmd = exec.Command(*binaryPath, "sign", "-ca-crt", *caCertFile, "-ca-key", *caKeyFile, "-duration", duration, "-name", h.Hostname, "-ip", h.IP, "-groups", groups)
		} else {
			cmd = exec.Command(*binaryPath, "sign", "-ca-crt", *caCertFile, "-ca-key", *caKeyFile, "-name", h.Hostname, "-ip", h.IP, "-groups", groups)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			l.Fatal("Host: " + h.Hostname + " " + string(output) + " Error: " + err.Error())
		}
		l.Println(h.Hostname + " OK " + string(output))
	}

}
