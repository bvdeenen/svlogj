package config

import (
	"bufio"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"svlogj/pkg/svlog"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

func ParseAndStoreConfig() {
	storeConfig(generateConfig())
}

func generateConfig() types.Config {
	facilities := utils.NewSet[string]()
	levels := utils.NewSet[string]()
	what := utils.NewSet[string]()
	entities := utils.NewSet[string]()
	services := utils.NewSet[string]()
	configFiles := make([]types.ConfigFile, 0)
	parse := func(line string) {
		re := regexp.MustCompile(`^([^.]+)\.([^:]+):?(.*)$`)
		if 0 == len(line) || line == "*" {
			return
		}
		m := re.FindStringSubmatch(line)
		facilities.Add(m[1])
		levels.Add(m[2])
		what.Add(m[3])
	}

	root := utils.SocklogDir()
	var confFile *types.ConfigFile
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		cobra.CheckErr(err)
		if d.IsDir() && path != root {
			services.Add(d.Name())
			configFiles = append(configFiles, types.ConfigFile{
				Service: d.Name(),
				Lines:   make([]string, 0),
			})
			confFile = &configFiles[len(configFiles)-1]

		}
		if d.Name() == "config" && !d.IsDir() {
			file, err := os.Open(path)
			cobra.CheckErr(err)
			defer func(file *os.File) {
				err := file.Close()
				cobra.CheckErr(err)
			}(file)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				confFile.Lines = append(confFile.Lines, line)
				switch line[0] {
				case '#':
					break
				case '+':
					parse(strings.TrimSpace(line[1:]))
					break
				case '-':
					parse(strings.TrimSpace(line[1:]))
					break
				case '!':
					break
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})
	svLogger := svlog.SvLogger{
		ParseConfig: types.ParseConfig{
			Follow: false,
		},
		LineHandler: func(info types.Info) {
			facilities.Add(info.Facility)
			levels.Add(info.Level)
			entities.Add(info.Entity)
		},
	}

	// parse through all of the socklog files to find all the entities
	svLogger.ParseLog()
	return types.Config{
		Facilities:  utils.RemoveEmptyStrings(facilities.Entries()),
		Levels:      utils.RemoveEmptyStrings(levels.Entries()),
		Entities:    utils.RemoveEmptyStrings(entities.Entries()),
		Services:    utils.RemoveEmptyStrings(services.Entries()),
		ConfigFiles: configFiles,
	}
}

func LoadConfig() types.Config {
	bytes, err := os.ReadFile(configFile())
	cobra.CheckErr(err)
	var config types.Config
	err = json.Unmarshal(bytes, &config)
	cobra.CheckErr(err)
	return config
}

func storeConfig(config types.Config) {
	serializedConfig, err := json.MarshalIndent(config, "", "  ")
	cobra.CheckErr(err)
	configFile := configFile()
	err = os.MkdirAll(path.Dir(configFile), 0700)
	cobra.CheckErr(err)
	f, err := os.Create(configFile)
	cobra.CheckErr(err)
	defer func(f *os.File) {
		err := f.Close()
		cobra.CheckErr(err)
	}(f)
	_, err = f.Write(serializedConfig)
	cobra.CheckErr(err)
}

func configFile() string {
	configFile, _ := xdg.ConfigFile("svlogj.json")
	return configFile
}
