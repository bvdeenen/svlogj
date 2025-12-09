package config

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"svlogj/pkg/svlog"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"

	"github.com/adrg/xdg"
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
	parse := func(line string) {
		re := regexp.MustCompile(`^([^.]+)\.([^:]+)(?::?(.*))$`)
		if 0 == len(line) || line == "*" {
			return
		}
		m := re.FindStringSubmatch(line)
		facilities.Add(m[1])
		levels.Add(m[2])
		what.Add(m[3])
	}

	root := "/var/log/socklog"
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		utils.Check(err)
		if d.IsDir() && path != root {
			services.Add(d.Name())
		}
		if d.Name() == "config" && !d.IsDir() {
			file, err := os.Open(path)
			utils.Check(err)
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
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
		Follow: false,
		LineHandler: func(info types.Info) {
			entities.Add(info.Entity)
		},
	}

	// parse through all of the socklog files to find all the entities
	svLogger.ParseLog()
	return types.Config{
		Facilities: utils.RemoveEmptyStrings(facilities.Entries()),
		Levels:     utils.RemoveEmptyStrings(levels.Entries()),
		Entities:   utils.RemoveEmptyStrings(entities.Entries()),
		Services:   utils.RemoveEmptyStrings(services.Entries()),
	}
}

func LoadConfig() types.Config {
	bytes, err := ioutil.ReadFile(configFile())
	utils.Check(err)
	var config types.Config
	err = json.Unmarshal(bytes, &config)
	utils.Check(err)
	return config
}

func storeConfig(config types.Config) {
	b, err := json.MarshalIndent(config, "", "  ")
	utils.Check(err)
	configFile := configFile()
	err = os.MkdirAll(path.Dir(configFile), 0700)
	utils.Check(err)
	f, err := os.Create(configFile)
	utils.Check(err)
	defer f.Close()
	f.Write(b)
}

func configFile() string {
	configFile, _ := xdg.ConfigFile("svlogj.json")
	return configFile
}
