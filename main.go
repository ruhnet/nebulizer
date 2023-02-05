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

type Host struct {
	Hostname string   `json:"hostname"`
	IP       string   `json:"ip"` //with CIDR network suffix
	Groups   []string `json:"groups,omitempty"`
	Duration float64  `json:"duration,omitempty"` //in days
}

func main() {
	var err error
	l := log.New(os.Stderr, "", 0) //set logging to standard error and no timestamp

	//commandline options
	caCertFile := flag.String("c", "./ca.crt", "CA certificate path.")
	caKeyFile := flag.String("k", "./ca.key", "CA key path.")
	binaryPath := flag.String("p", "", "Path to nebula-cert binary file. If not specified, search $PATH and current directory.")
	networkFile := flag.String("f", "-", "Path to network input file. Use '-' for standard input.")
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
	//fileBytes, _ := ioutil.ReadAll(inputFile)

	var input string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		input = input + scanner.Text()
	}

	var network []Host

	err = json.Unmarshal([]byte(input), &network) //read the network config
	if err != nil {
		if *networkFile == "-" {
			*networkFile = "standard input."
		}
		l.Fatal("Could not parse network description from " + *networkFile + "\nError: " + err.Error())
	}

	for _, h := range network {
		groups := strings.Join(h.Groups, ",")
		var cmd *exec.Cmd
		if h.Duration > 0 {
			duration := strconv.Itoa(int(math.Round(h.Duration*24))) + "h"
			cmd = exec.Command(*binaryPath, "sign", "-ca-crt", *caCertFile, "-ca-key", *caKeyFile, "-duration", duration, "-name", h.Hostname, "-ip", h.IP, "-groups", groups)
		} else {
			cmd = exec.Command(*binaryPath, "sign", "-ca-crt", *caCertFile, "-ca-key", *caKeyFile, "-name", h.Hostname, "-ip", h.IP, "-groups", groups)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			l.Fatal(h.Hostname + " " + string(output) + " Error: " + err.Error())
		}
		l.Println(h.Hostname + " OK " + string(output))
	}

}
