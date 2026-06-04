package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func getVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
			return "", fmt.Errorf("Node.js is not installed or not found in PATH")
		} else {
			return "", fmt.Errorf("error executing node command: %v", err)
		}
	}
	v := strings.TrimSpace(string(output))
	return strings.TrimPrefix(v, "v"), nil
}

func getOverwriteConfirmation() bool {
	// var response string
	// fmt.Scanln(&response)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return len(response) > 0 && response[0] == 'y'
}

func getConfirmationIfFileExists(filePath string, parsedExecVersion *semver.Version, parsedVersionStr string) (bool, error) {
	fileVersion, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("error reading %s: %v", filePath, err)
		}
		return true, nil // file doesn't exist, so we can create it
	}

	existingVersion := strings.TrimSpace(string(fileVersion))

	// skip parsing if existing version is exactly the same as the current version (with or without "v" prefix)
	if existingVersion == parsedVersionStr || existingVersion == "v"+parsedVersionStr {
		return false, nil
	}

	parsedExistingVersion, err := semver.NewVersion(existingVersion)
	if err != nil {
		if strings.HasSuffix(filePath, ".nvmrc") {
			// likely non-semver format in .nvmrc (e.g., "lts/*", "node", "14", etc.)
			fmt.Printf("Version (%s) in %s is not a valid semantic version.\nDo you want to overwrite it with version %s? (y/n): ", existingVersion, filePath, parsedExecVersion)
			overwrite := getOverwriteConfirmation()
			if !overwrite {
				fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
				return false, nil
			}
			return true, nil
		}
		fmt.Printf("Error parsing version in %s: %v\n", filePath, err)
		fmt.Printf("Do you want to overwrite it with version %s? (y/n): ", parsedExecVersion)
		overwrite := getOverwriteConfirmation()
		if !overwrite {
			fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
			return false, nil
		}
	}

	if !parsedExistingVersion.Equal(parsedExecVersion) {
		fmt.Printf("File %s already exists with version %s.\nDo you want to overwrite it with version %s? (y/n): ", filePath, parsedExistingVersion, parsedExecVersion)
		overwrite := getOverwriteConfirmation()
		if !overwrite {
			fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
			return false, nil
		}
	} else {
		// fmt.Printf("%s already set to version %s. No changes needed.\n", filePath, existingVersion)
		return false, nil
	}
	return true, nil
}

func processVersionFile(filePath string, parsedVersion *semver.Version, parsedVersionStr string, forceOverwrite bool) (bool, error) {
	overwrite := forceOverwrite
	if !forceOverwrite {
		var err error
		overwrite, err = getConfirmationIfFileExists(filePath, parsedVersion, parsedVersionStr)
		if err != nil {
			return false, err
		}
	}
	if overwrite {
		err := os.WriteFile(filePath, []byte(parsedVersionStr+"\n"), 0644)
		if err != nil {
			return false, fmt.Errorf("error writing to %s: %v", filePath, err)
		}
		return true, nil
	}
	return false, nil
}

func run(versionFlag string, forceOverwrite bool, createNvmrc bool, modifyPackageJson bool) error {
	start := time.Now()
	version := versionFlag
	if version == "" {
		var err error
		version, err = getVersion()
		if err != nil {
			return err
		}
	}

	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("error parsing Node.js version: %v", err)
	}

	nvmrcSet := false
	nodeversionSet := false
	packageJsonSet := false

	nodeversionSet, err = processVersionFile(".node-version", parsedVersion, version, forceOverwrite)
	if err != nil {
		return err
	}

	if createNvmrc {
		nvmrcSet, err = processVersionFile(".nvmrc", parsedVersion, version, forceOverwrite)
		if err != nil {
			return err
		}
	}

	if modifyPackageJson {
		if _, err := os.Stat("package.json"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "package.json not found. Skipping engines.node update.")
		} else {
			cmd := exec.Command("npm", "pkg", "set", fmt.Sprintf("engines.node=%s", version))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("error updating package.json engines.node field: %v", err)
			}
			packageJsonSet = true
		}
	}

	elapsed := time.Since(start)
	if nodeversionSet || (createNvmrc && nvmrcSet) || (modifyPackageJson && packageJsonSet) {
		// fmt.Printf("Version files updated successfully. Done in %s.\n", elapsed.Round(time.Millisecond))
		fmt.Printf("Version files updated successfully. Done in %s.\n", color.BlueString("%s", elapsed.Round(time.Millisecond)))
	} else {
		// fmt.Printf("Version files are already up to date. No changes made. Done in %s.\n", elapsed.Round(time.Millisecond))
		fmt.Printf("Version files are already up to date, no changes made. Done in %s.\n", color.BlueString("%s", elapsed.Round(time.Millisecond)))
	}
	return nil
}

func main() {
	var versionFlag string
	var forceOverwrite bool
	var createNvmrc bool
	var modifyPackageJson bool

	var rootCmd = &cobra.Command{
		Use:   "pin-node [flags]",
		Short: "Pin your node version to your current directory in milliseconds",
		Long: `pin-node helps you pin the Node.js version for your current project directory.

It creates or updates the following files:
  • .node-version (Always)
  • .nvmrc        (Optional)
  • package.json  (Optional, updates the 'engines.node' field)`,
		// Example: `# Pin to the currently installed Node.js version
		// pin-node

		// # Pin a specific Node.js version
		// pin-node -v 18.16.0

		// # Pin and also update .nvmrc and package.json
		// pin-node -v 18.16.0 --nvmrc --engines

		// # Force overwrite without interactive prompts
		// pin-node -v 20.0.0 --force --nvmrc --engines`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(versionFlag, forceOverwrite, createNvmrc, modifyPackageJson)
		},
	}

	rootCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Specify a custom Node.js version to pin (defaults to current version)")
	rootCmd.Flags().BoolVar(&createNvmrc, "nvmrc", false, "Also create or update .nvmrc file")
	rootCmd.Flags().BoolVarP(&modifyPackageJson, "engines", "e", false, "Update engines.node field in package.json")
	rootCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Force overwrite existing version files without confirmation")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
