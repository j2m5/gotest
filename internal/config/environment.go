package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func loadEnvironment(options Options) (func(string) string, string, error) {
	base, err := os.Getwd()
	if err != nil {
		return nil, "", errors.New("cannot resolve startup working directory")
	}
	var fileValues map[string]string
	if options.EnvFile != "" {
		filename, err := filepath.Abs(options.EnvFile)
		if err != nil {
			return nil, "", errors.New("cannot resolve environment file; check -env-file")
		}
		fileValues, err = godotenv.Read(filename)
		if err != nil {
			if !options.OptionalEnvFile || !errors.Is(err, os.ErrNotExist) {
				// Parser and filesystem errors may contain secret lines or private paths.
				return nil, "", errors.New("cannot load environment file; check -env-file, permissions and format")
			}
		} else {
			base = filepath.Dir(filename)
		}
	}
	lookup := func(key string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return fileValues[key]
	}
	return lookup, base, nil
}

func resolveDirectory(base, value, key string) (string, error) {
	if strings.TrimSpace(value) == "" || strings.ContainsRune(value, 0) {
		return "", fmt.Errorf("%s must be a nonempty directory path", key)
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	// Windows drive-relative (C:dir) and root-relative (\dir) names do not
	// specify a stable base. Ordinary relative paths and absolute UNC paths do.
	if filepath.VolumeName(value) != "" || os.IsPathSeparator(value[0]) {
		return "", fmt.Errorf("%s must be absolute or relative to the configuration directory", key)
	}
	return filepath.Join(base, value), nil
}
