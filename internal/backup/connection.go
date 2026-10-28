package backup

import (
	"net"
	"net/url"
	"os"
)

func (c *Connection) connectionString() string {
	parameters := url.Values{}
	for key, values := range c.parameters {
		parameters[key] = append([]string(nil), values...)
	}
	// Empty explicit values override pgx's implicit ~/.postgresql certificates.
	for _, key := range []string{"sslrootcert", "sslcert", "sslkey"} {
		if _, ok := parameters[key]; !ok {
			parameters.Set(key, "")
		}
	}
	parameters.Set("passfile", os.DevNull)
	result := url.URL{Scheme: "postgres", User: url.UserPassword(c.User, c.password), Host: net.JoinHostPort(c.Host, c.Port), Path: "/" + c.Database, RawQuery: parameters.Encode()}
	return result.String()
}
