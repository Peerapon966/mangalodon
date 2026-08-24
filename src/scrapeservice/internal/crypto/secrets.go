package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"

	"golang.org/x/crypto/pbkdf2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Secrets struct {
	KEK []byte
	DEK []byte
}

func GetSecrets(ctx context.Context) (Secrets, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return Secrets{}, &apperr.ScraperError{
			Code:    apperr.EnvVarNotSet,
			Message: "unable to load SDK config.",
			Err:     err.Error(),
		}
	}

	password, err := getSecret(ctx, &cfg, fmt.Sprintf("/blackbox/%s/vault-password", os.Getenv("ENVIRONMENT")))
	if err != nil {
		return Secrets{}, err
	}

	salt, err := getParam(ctx, &cfg, fmt.Sprintf("/blackbox/%s/vault-salt", os.Getenv("ENVIRONMENT")))
	if err != nil {
		return Secrets{}, err
	}

	b64Dek, err := getParam(ctx, &cfg, fmt.Sprintf("/blackbox/%s/dek", os.Getenv("ENVIRONMENT")))
	if err != nil {
		return Secrets{}, err
	}

	dek, err := base64.StdEncoding.DecodeString(b64Dek)
	if err != nil {
		return Secrets{}, &apperr.ScraperError{
			Code:    apperr.DecodeFailed,
			Message: "unable to decode Base64 DEK.",
			Err:     err.Error(),
		}
	}

	return Secrets{
		KEK: deriveKEK(password, []byte(salt)),
		DEK: dek,
	}, nil
}

func getSecret(ctx context.Context, cfg *aws.Config, name string) (string, error) {
	client := secretsmanager.NewFromConfig(*cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return "", &apperr.ScraperError{
			Code:    apperr.GetSecretValueFailed,
			Message: fmt.Sprintf("Error getting secret %s from Secrets Manager", name),
			Err:     err.Error(),
		}
	}

	return *result.SecretString, nil
}

func getParam(ctx context.Context, cfg *aws.Config, name string) (string, error) {
	client := ssm.NewFromConfig(*cfg)

	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	}

	result, err := client.GetParameter(ctx, input)
	if err != nil {
		return "", &apperr.ScraperError{
			Code:    apperr.GetParameterFailed,
			Message: fmt.Sprintf("Error getting parameter %s from SSM Parameter Store", name),
			Err:     err.Error(),
		}
	}

	return *result.Parameter.Value, nil
}

// DeriveKEK uses PBKDF2 to derive a 32-byte AES key from a human-readable password.
// This is strictly used for encrypting/decrypting the library.json.enc index file.
func deriveKEK(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}
