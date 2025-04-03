package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Voice     string `json:"voice,omitempty"`
	Speed     int    `json:"speed,omitempty"`
	Pitch     int    `json:"pitch,omitempty"`
	Amplitude int    `json:"amplitude,omitempty"`
}

var synthesisConfig = Config{
	Voice:     "en-us",
	Speed:     175,
	Pitch:     50,
	Amplitude: 100,
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Usage: orator <text> | -s <text_file> -o <output_file> | -<param> <value>")
		os.Exit(1)
	}

	if strings.HasPrefix(args[0], "-") && len(args) > 1 {
		setParameter(args[0][1:], args[1])
		return
	}

	text, outputFile := parseArgs(args)
	if text == "" {
		fmt.Println("Error: No text provided for synthesis")
		os.Exit(1)
	}

	jobID := sendRequest(text)
	pollForResult(jobID, outputFile)
}

func parseArgs(args []string) (string, string) {
	var text, outputFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "-s":
			if i+1 < len(args) {
				content, err := os.ReadFile(args[i+1])
				if err != nil {
					fmt.Println("Error reading file:", err)
					os.Exit(1)
				}
				text = string(content)
				i++
			}
		default:
			if text == "" {
				text = args[i]
			} else {
				text += " " + args[i]
			}
		}
	}
	return text, outputFile
}

func setParameter(param, value string) {
	paramMap := map[string]interface{}{}
	switch param {
	case "speed", "pitch", "amplitude":
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err != nil {
			fmt.Printf("Invalid value for %s. Must be an integer.\n", param)
			os.Exit(1)
		}
		paramMap[param] = intValue
	case "voice":
		paramMap["voice"] = value
	default:
		fmt.Println("Invalid parameter:", param)
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(paramMap)
	req, err := http.NewRequest("PATCH", "http://localhost:8080/api/v1/set_parameter", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Parameter updated successfully.")
	} else {
		fmt.Println("Failed to update parameter.")
	}
}

func sendRequest(text string) string {
	requestBody := map[string]interface{}{
		"text":      text,
		"voice":     synthesisConfig.Voice,
		"speed":     synthesisConfig.Speed,
		"pitch":     synthesisConfig.Pitch,
		"amplitude": synthesisConfig.Amplitude,
	}

	reqBody, _ := json.Marshal(requestBody)
	resp, err := http.Post("http://localhost:8080/api/v1/to_speech", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var response map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	return response["job_id"]
}

func pollForResult(jobID, outputFile string) {
	for {
		resp, _ := http.Get(fmt.Sprintf("http://localhost:8080/job_status?job_id=%s", jobID))
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if outputFile != "" {
				file, _ := os.Create(outputFile)
				io.Copy(file, resp.Body)
				fmt.Printf("Speech saved to %s\n", outputFile)
			} else {
				fmt.Println("Speech synthesis completed.")
			}
			return
		}

		time.Sleep(2 * time.Second)
	}
}
