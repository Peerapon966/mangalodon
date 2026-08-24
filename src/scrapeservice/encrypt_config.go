//go:build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/Peerapon966/blackbox/scraper/internal/crypto"
)

func main() {
	// IMPORTANT: This script MUST be run from inside the 'scraper' directory!
	// Example: cd scraper && go run encrypt_config.go

	// 1. Read the plaintext config (Ignored by Git)
	plaintext, err := os.ReadFile("internal/crawler/config/configs.json")
	if err != nil {
		fmt.Println("No configs.json found at internal/crawler/config/configs.json, skipping encryption.")
		return
	}

	// 2. Read the local keys file from the project root (Ignored by Git)
	envBytes, err := os.ReadFile("../.env.keys")
	if err != nil {
		fmt.Println("Missing ../.env.keys file. Please create it with DEV_DEK and PROD_DEK to encrypt configs.")
		os.Exit(1)
	}
	env := string(envBytes)

	// 3. Encrypt and save for DEV
	if strings.Contains(env, "DEV_DEK=") {
		dek := extractKey(env, "DEV_DEK=")
		if len(dek) == 32 {
			encryptAndSave(plaintext, dek, "internal/crawler/config/configs.dev.json.enc")
		} else {
			fmt.Println("Warning: DEV_DEK is not exactly 32 bytes!")
		}
	}

	// 4. Encrypt and save for PROD
	if strings.Contains(env, "PROD_DEK=") {
		dek := extractKey(env, "PROD_DEK=")
		if len(dek) == 32 {
			encryptAndSave(plaintext, dek, "internal/crawler/config/configs.prod.json.enc")
		} else {
			fmt.Println("Warning: PROD_DEK is not exactly 32 bytes!")
		}
	}
}

func encryptAndSave(plaintext, dek []byte, filename string) {
	ciphertext, err := crypto.EncryptBlob(plaintext, dek)
	if err != nil {
		panic(fmt.Sprintf("Failed to encrypt %s: %v", filename, err))
	}

	err = os.WriteFile(filename, ciphertext, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to write %s: %v", filename, err))
	}

	fmt.Printf("Successfully encrypted and saved %s\n", filename)
}

func extractKey(env, prefix string) []byte {
	lines := strings.Split(env, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			base64Key := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			// Decode the Base64 string exported by Terraform back into 32 raw bytes!
			decoded, err := base64.StdEncoding.DecodeString(base64Key)
			if err != nil {
				panic(fmt.Sprintf("Key for %s is not valid Base64: %v", prefix, err))
			}
			return decoded
		}
	}
	return nil
}
