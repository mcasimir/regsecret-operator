package controller

import "errors"

type Config struct {
	SecretName        string                 `json:"secretName"`
	NamespaceSelector string                 `json:"namespaceSelector"`
	Credentials       RegistryCredentialsMap `json:"credentials"`
}

type RegistryCredentialsMap map[string]RegistryCredentials
type RegistryCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c *Config) Validate() error {
	if c.SecretName == "" {
		return errors.New("secretName is required")
	}

	for _, credentials := range c.Credentials {
		if credentials.Email == "" {
			return errors.New("credentials[*].email is required")
		}

		if credentials.Username == "" {
			return errors.New("credentials[*].username is required")
		}

		if credentials.Password == "" {
			return errors.New("credentials[*].password is required")
		}
	}

	return nil
}
