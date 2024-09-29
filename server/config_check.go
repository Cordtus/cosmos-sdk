package server

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

func CheckConfigFiles(homeDir string) error {
	configPath := filepath.Join(homeDir, "config")

	// Check config.toml
	if err := CheckAndUpdateFile(configPath, "config.toml", config.DefaultConfig(), false); err != nil {
		return err
	}

	// Check app.toml
	v := viper.New()
	defaultAppConfig, err := config.GetConfig(v)
	if err != nil {
		return fmt.Errorf("error getting default app config: %w", err)
	}
	if err := CheckAndUpdateFile(configPath, "app.toml", defaultAppConfig, false); err != nil {
		return err
	}

	return nil
}

// CheckAndUpdateFile checks and potentially updates a configuration file
func CheckAndUpdateFile(configPath, fileName string, defaultConfig interface{}, autoUpdate bool) error {
	filePath := filepath.Join(configPath, fileName)

	// Read existing file
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", fileName, err)
	}

	// Parse existing config
	var existingMap map[string]interface{}
	if err := toml.Unmarshal(existingContent, &existingMap); err != nil {
		return fmt.Errorf("error parsing existing %s: %w", fileName, err)
	}

	// Generate new config content
	newContent, err := toml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("error generating new %s content: %w", fileName, err)
	}

	// Parse new config
	var newMap map[string]interface{}
	if err := toml.Unmarshal(newContent, &newMap); err != nil {
		return fmt.Errorf("error parsing new %s: %w", fileName, err)
	}

	// Compare structures
	added, removed, modified := compareConfigs(existingMap, newMap)

	if len(added) > 0 || len(removed) > 0 || len(modified) > 0 {
		fmt.Printf("Your %s file structure is outdated.\n", fileName)
		if len(added) > 0 {
			fmt.Println("Added parameters:")
			for _, key := range added {
				fmt.Printf("  + %s\n", key)
			}
		}
		if len(removed) > 0 {
			fmt.Println("Removed parameters:")
			for _, key := range removed {
				fmt.Printf("  - %s\n", key)
			}
		}
		if len(modified) > 0 {
			fmt.Println("Modified parameters:")
			for _, key := range modified {
				fmt.Printf("  * %s\n", key)
			}
		}

		var shouldUpdate bool
		if autoUpdate {
			shouldUpdate = true
		} else {
			fmt.Printf("Do you want to update %s with the new structure? (y/n): ", fileName)
			var response string
			fmt.Scanln(&response)
			shouldUpdate = response == "y" || response == "Y"
		}

		if shouldUpdate {
			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				return fmt.Errorf("error writing updated %s: %w", fileName, err)
			}
			fmt.Printf("%s has been updated with the new structure.\n", fileName)
		} else {
			fmt.Printf("%s was not updated. Please manually update your file structure.\n", fileName)
		}
	}

	return nil
}

func compareConfigs(existing, new map[string]interface{}) (added, removed, modified []string) {
	for key, newValue := range new {
		if existingValue, ok := existing[key]; !ok {
			added = append(added, key)
		} else if !reflect.DeepEqual(existingValue, newValue) {
			if _, ok := existingValue.(map[string]interface{}); ok {
				// Recursively compare nested structures
				subAdded, subRemoved, subModified := compareConfigs(existingValue.(map[string]interface{}), newValue.(map[string]interface{}))
				for _, subKey := range subAdded {
					added = append(added, key+"."+subKey)
				}
				for _, subKey := range subRemoved {
					removed = append(removed, key+"."+subKey)
				}
				for _, subKey := range subModified {
					modified = append(modified, key+"."+subKey)
				}
			} else {
				modified = append(modified, key)
			}
		}
	}

	for key := range existing {
		if _, ok := new[key]; !ok {
			removed = append(removed, key)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(modified)

	return
}
