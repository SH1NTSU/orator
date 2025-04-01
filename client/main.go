package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Voice     string `json:"voice"`
	Speed     int    `json:"speed"`
	Pitch     int    `json:"pitch"`
	Amplitude int    `json:"amplitude"`
}

var (
	synthesisConfig Config
	outputFile      string
	textFile        string
)

func main() {
	loadDefaultArgs()
	args := os.Args[1:]
	var text string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-o" && i+1 < len(args):
			outputFile = args[i+1]
			i++
		case arg == "-s" && i+1 < len(args):
			textFile = args[i+1]
			i++
		case strings.HasPrefix(arg, "-"):
			param := strings.TrimPrefix(arg, "-")
			if i+1 < len(args) {
				setParameter(param, args[i+1])
				i++
			} else {
				fmt.Printf("Error: Missing value for parameter %s\n", param)
				os.Exit(1)
			}
		default:
			if text == "" {
				text = arg
			} else {
				text += " " + arg
			}
		}
	}

	if textFile != "" {
		content, err := os.ReadFile(textFile)
		if err != nil {
			fmt.Printf("Error reading text file: %v\n", err)
			os.Exit(1)
		}
		text = string(content)
	}

	if text == "" {
		fmt.Println("Error: No text provided for synthesis")
		os.Exit(1)
	}

	sendRequest(text, outputFile)
}

func loadDefaultArgs() {
	synthesisConfig = Config{
		Voice:     "en-us",
		Speed:     175,
		Pitch:     50,
		Amplitude: 100,
	}
	outputFile = ""
	textFile = ""
}

func setParameter(name, value string) {
	switch name {
	case "voice":
		synthesisConfig.Voice = value
	case "speed":
		var speed int
		_, err := fmt.Sscanf(value, "%d", &speed)
		if err != nil || speed <= 0 {
			fmt.Println("Error: speed must be a positive integer")
			os.Exit(1)
		}
		synthesisConfig.Speed = speed
	case "pitch":
		var pitch int
		_, err := fmt.Sscanf(value, "%d", &pitch)
		if err != nil || pitch < 0 || pitch > 99 {
			fmt.Println("Error: pitch must be between 0 and 99")
			os.Exit(1)
		}
		synthesisConfig.Pitch = pitch
	case "amplitude":
		var amplitude int
		_, err := fmt.Sscanf(value, "%d", &amplitude)
		if err != nil || amplitude < 0 || amplitude > 200 {
			fmt.Println("Error: amplitude must be between 0 and 200")
			os.Exit(1)
		}
		synthesisConfig.Amplitude = amplitude
	default:
		fmt.Printf("Error: Unknown parameter '%s'\n", name)
		os.Exit(1)
	}
}

func sendRequest(text, outputFile string) {
	requestBody := map[string]interface{}{
		"text":      text,
		"voice":     synthesisConfig.Voice,
		"speed":     synthesisConfig.Speed,
		"pitch":     synthesisConfig.Pitch,
		"amplitude": synthesisConfig.Amplitude,
	}

	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	resp, err := http.Post("http://localhost:8080/api/v1/to_speech", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error from server (status %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	if outputFile != "" {
		outFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			fmt.Println("Error saving audio:", err)
			os.Exit(1)
		}
		fmt.Printf("Speech saved to %s\n", outputFile)
	} else {
		fmt.Println("Speech synthesized successfully")
	}
}
