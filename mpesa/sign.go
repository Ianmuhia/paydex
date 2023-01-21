package mpesa

import (
	"encoding/base64"
	"errors"
	"fmt"
)

func GeneratePassword(shortCode, passkey, time string) (string, error) {
	if shortCode == "" || passkey == "" || time == "" {
		return "", errors.New("unable to generate password from empty string")
	}
	password := fmt.Sprintf("%s%s%s", shortCode, passkey, time)
	return base64.StdEncoding.EncodeToString([]byte(password)), nil
}

//TODO: mpesa services
//TODO: jenga services
//TODO: currency conversion api
//TODO: store all fainled and passed payments
//TODO: retention periods
