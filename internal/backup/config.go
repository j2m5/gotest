// Package backup creates and restores trusted, locally stored PostgreSQL archives.
package backup

import (
	"encoding/base64"
	"errors"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Connection contains a single explicit database identity. Secrets remain private.
type Connection struct {
	Host, Port, User, Database string
	password                   string
	parameters                 url.Values
}

var supportedParameters = map[string]string{
	"sslmode": "PGSSLMODE", "sslrootcert": "PGSSLROOTCERT", "sslcert": "PGSSLCERT", "sslkey": "PGSSLKEY",
	"connect_timeout": "PGCONNECT_TIMEOUT", "channel_binding": "PGCHANNELBINDING",
}

func controlled(text string) error { return errors.New(text) }
func hasControl(s string) bool     { return strings.IndexFunc(s, unicode.IsControl) >= 0 }

// ParseConnection rejects implicit libpq configuration and options whose semantics
// cannot be enforced consistently by both pgx and PostgreSQL command line tools.
func ParseConnection(raw string, environment []string) (*Connection, error) {
	for _, entry := range environment {
		key, value, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(key), "PG") && value != "" {
			return nil, controlled("PG environment variables must be unset; use the explicit database URL")
		}
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Opaque != "" || u.Fragment != "" || hasControl(raw) || u.User == nil {
		return nil, controlled("invalid PostgreSQL URL")
	}
	password, _ := u.User.Password()
	c := &Connection{Host: u.Hostname(), Port: u.Port(), User: u.User.Username(), Database: strings.TrimPrefix(u.Path, "/"), password: password}
	if c.Port == "" {
		c.Port = "5432"
	}
	port, err := strconv.ParseUint(c.Port, 10, 16)
	if err != nil || port == 0 {
		return nil, controlled("invalid PostgreSQL port")
	}
	c.Port = strconv.FormatUint(port, 10)
	if c.Host == "" || c.User == "" || c.Database == "" || strings.ContainsAny(c.Host, ",/=\\*") || strings.Contains(c.Database, "=") || strings.HasPrefix(strings.ToLower(c.Database), "postgres:") || strings.HasPrefix(strings.ToLower(c.Database), "postgresql:") {
		return nil, controlled("database URL must specify one host, user and plain database name")
	}
	for _, part := range []string{c.Host, c.User, c.Database, c.password} {
		if hasControl(part) {
			return nil, controlled("database URL contains control characters")
		}
	}
	c.parameters, err = url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, controlled("invalid PostgreSQL URL options")
	}
	for key, values := range c.parameters {
		if _, ok := supportedParameters[key]; !ok {
			return nil, controlled("unsupported PostgreSQL URL option")
		}
		if len(values) != 1 || values[0] == "" || hasControl(values[0]) {
			return nil, controlled("invalid or repeated PostgreSQL URL option")
		}
	}
	switch c.parameters.Get("sslmode") {
	case "":
		if c.parameters.Get("sslrootcert") == "system" {
			c.parameters.Set("sslmode", "verify-full")
		} else {
			c.parameters.Set("sslmode", "prefer")
		}
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return nil, controlled("invalid SSL mode")
	}
	mode := c.parameters.Get("sslmode")
	if (mode == "verify-ca" || mode == "verify-full") && c.parameters.Get("sslrootcert") == "" {
		return nil, controlled("certificate verification requires an explicit sslrootcert")
	}
	if c.parameters.Get("sslrootcert") == "system" && mode != "verify-full" {
		return nil, controlled("system trust requires sslmode verify-full")
	}
	switch c.parameters.Get("channel_binding") {
	case "", "disable", "prefer", "require":
	default:
		return nil, controlled("invalid channel binding")
	}
	if v := c.parameters.Get("connect_timeout"); v != "" {
		n, e := strconv.ParseUint(v, 10, 31)
		if e != nil || n == 0 {
			return nil, controlled("connect timeout must be a positive number of seconds")
		}
	}
	if (c.parameters.Get("sslcert") == "") != (c.parameters.Get("sslkey") == "") {
		return nil, controlled("SSL certificate and key must be provided together")
	}
	return c, nil
}

// ParseKey accepts only the configured 256 bit base64 queue key.
func ParseKey(raw string) ([]byte, error) {
	if hasControl(raw) {
		return nil, controlled("MAIL_QUEUE_KEY must be base64 encoding of exactly 32 bytes")
	}
	key, err := base64.StdEncoding.Strict().DecodeString(raw)
	if err != nil || len(key) != 32 {
		return nil, controlled("MAIL_QUEUE_KEY must be base64 encoding of exactly 32 bytes")
	}
	return key, nil
}

// Pgpass escapes delimiters according to libpq's password file format.
func (c *Connection) Pgpass() string {
	escape := func(s string) string { return strings.NewReplacer("\\", "\\\\", ":", "\\:").Replace(s) }
	return escape(c.Host) + ":" + c.Port + ":" + escape(c.Database) + ":" + escape(c.User) + ":" + escape(c.password) + "\n"
}

// ToolEnvironment removes inherited database/mail credentials. Absent certificate
// options point to nonexistent files beside the private password file, disabling
// implicit libpq certificate and CRL loading on both Unix and Windows.
func (c *Connection) ToolEnvironment(environment []string, passfile string) []string {
	var result []string
	for _, entry := range environment {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		if strings.HasPrefix(key, "PG") || strings.HasPrefix(key, "SMTP_") || key == "DATABASE_URL" || key == "RESTORE_DATABASE_URL" || key == "TEST_DATABASE_URL" || key == "MAIL_QUEUE_KEY" || key == "HOME" || key == "APPDATA" || key == "USERPROFILE" {
			continue
		}
		result = append(result, entry)
	}
	result = append(result, "PGHOST="+c.Host, "PGPORT="+c.Port, "PGUSER="+c.User, "PGDATABASE="+c.Database, "PGPASSFILE="+passfile)
	for key, env := range supportedParameters {
		if value := c.parameters.Get(key); value != "" {
			result = append(result, env+"="+value)
		}
	}
	for parameter, env := range map[string]string{"sslrootcert": "PGSSLROOTCERT", "sslcert": "PGSSLCERT", "sslkey": "PGSSLKEY", "sslcrl": "PGSSLCRL"} {
		if c.parameters.Get(parameter) == "" {
			result = append(result, env+"="+filepath.Join(filepath.Dir(passfile), "unused-"+strings.ToLower(env)))
		}
	}
	return result
}
