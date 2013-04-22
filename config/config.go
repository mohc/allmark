// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"bufio"
	"fmt"
	"github.com/andreaskoch/allmark/themes"
	"github.com/andreaskoch/allmark/util"
	"os"
	"os/user"
	"path/filepath"
)

const (
	MetaDataFolderName    = ".allmark"
	ConfigurationFileName = "config"
	ThemeFolderName       = "theme"
)

func Initialize(repositoryPath string) {
	config := GetConfig(repositoryPath)

	// create config
	if _, err := config.save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating configuration file %q. Error: ", config.Filepath(), err)
	}

	// create theme
	themeFolder := config.ThemeFolder()
	if !util.CreateDirectory(themeFolder) {
		fmt.Fprintf(os.Stderr, "Unable to create theme folder %q.", themeFolder)
	}

	themeFile := filepath.Join(themeFolder, "screen.css")
	file, err := os.Create(themeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create theme file %q.", themeFile)
	}

	defer file.Close()

	file.WriteString(themes.GetTheme())
}

func GetConfig(repositoryPath string) *Config {

	// return the local config
	if localConfig, err := local(repositoryPath).load(); err == nil {
		return localConfig
	}

	// return the global config
	if homeDirectory, homeDirError := getUserHomeDir(); homeDirError == nil {
		if globalConfig, err := global(homeDirectory).load(); err == nil {
			return globalConfig
		}
	}

	// return the default config
	return defaultConfig(repositoryPath)
}

type Http struct {
	Port int
}

type Web struct {
	DefaultLanguage string
}

type Server struct {
	ThemeFolderName string
	Http            Http
}

type Config struct {
	Server Server
	Web    Web

	baseFolder      string
	metaDataFolder  string
	themeFolderBase string
}

func (config *Config) BaseFolder() string {
	return config.baseFolder
}

func (config *Config) MetaDataFolder() string {
	return config.metaDataFolder
}

func (config *Config) Filepath() string {
	return filepath.Join(config.MetaDataFolder(), ConfigurationFileName)
}

func (config *Config) ThemeFolder() string {
	themeFolderName := ThemeFolderName
	if config.Server.ThemeFolderName != "" {
		themeFolderName = config.Server.ThemeFolderName
	}

	return filepath.Join(config.themeFolderBase, themeFolderName)
}

func (config *Config) load() (*Config, error) {

	path := config.Filepath()

	// check if file can be accessed
	fileInfo, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("Cannot read config file %q. Error: %s", path, err)
	}

	defer fileInfo.Close()

	// deserialize config
	serializer := NewJSONSerializer()
	loadedConfig, err := serializer.DeserializeConfig(fileInfo)
	if err != nil {
		return config, fmt.Errorf("Could not deserialize the configuration file %q. Error: %s", path, err)
	}

	// apply values
	config.Server = loadedConfig.Server
	config.Web = loadedConfig.Web

	return config, nil
}

func (config *Config) save() (*Config, error) {

	path := config.Filepath()

	// create or overwrite the config file
	if created, err := util.CreateFile(path); !created {
		return config, fmt.Errorf("Could not create configuration file %q. Error: ", path, err)
	}

	// open the file for writing
	file, err := os.OpenFile(path, os.O_WRONLY, 0776)
	if err != nil {
		return config, fmt.Errorf("Error while opening file %q for writing.", path)
	}

	writer := bufio.NewWriter(file)

	defer func() {
		writer.Flush()
		file.Close()
	}()

	// serialize the config
	serializer := NewJSONSerializer()
	if serializationError := serializer.SerializeConfig(writer, config); serializationError != nil {
		return config, fmt.Errorf("Error while saving configuration %#v to file %q. Error: %v", config, path, serializationError)
	}

	return config, nil
}

func local(baseFolder string) *Config {
	metaDataFolder := filepath.Join(baseFolder, MetaDataFolderName)

	return &Config{
		baseFolder:      baseFolder,
		metaDataFolder:  metaDataFolder,
		themeFolderBase: metaDataFolder,
	}
}

func global(baseFolder string) *Config {
	metaDataFolder := filepath.Join(baseFolder, MetaDataFolderName)

	return &Config{
		baseFolder:      baseFolder,
		metaDataFolder:  metaDataFolder,
		themeFolderBase: baseFolder,
	}
}

func defaultConfig(baseFolder string) *Config {
	defaultConfig := local(baseFolder)
	defaultConfig.Server.ThemeFolderName = ThemeFolderName
	defaultConfig.Server.Http.Port = 8080
	defaultConfig.Web.DefaultLanguage = "en"

	return defaultConfig
}

// Get the current users home directory path
func getUserHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Cannot determine the current users home direcotry. Error: %s", err)
	}

	return usr.HomeDir, nil
}
