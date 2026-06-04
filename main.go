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
)

func getVersion() string {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
			fmt.Fprintln(os.Stderr, "Node.js is not installed or not found in PATH.")
		} else {
			fmt.Fprintln(os.Stderr, "Error executing node command:", err)
		}
		os.Exit(1)

	}
	v := strings.TrimSpace(string(output))
	return strings.TrimPrefix(v, "v")
}

func getOverwriteConfirmation() bool {
	// var response string
	// fmt.Scanln(&response)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return len(response) > 0 && response[0] == 'y'
}

func getConfirmationIfFileExists(filePath string, parsedExecVersion *semver.Version) bool {
	fileVersion, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
			os.Exit(1)
		}
		return true // file doesn't exist, so we can create it
	}

	existingVersion := strings.TrimSpace(string(fileVersion))
	execVersionStr := parsedExecVersion.String()

	// skip parsing if existing version is exactly the same as the current version (with or without "v" prefix)
	if existingVersion == execVersionStr || existingVersion == "v"+execVersionStr {
		return false
	}


	parsedExistingVersion, err := semver.NewVersion(existingVersion)
	if err != nil {
		if strings.HasSuffix(filePath, ".nvmrc") {
			// likely non-semver format in .nvmrc (e.g., "lts/*", "node", "14", etc.)
			fmt.Printf("Version (%s) in %s is not a valid semantic version.\nDo you want to overwrite it with version %s? (y/n): ", existingVersion, filePath, parsedExecVersion)
			overwrite := getOverwriteConfirmation()
			if !overwrite {
				fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
				return false
			}
			return true
		}
		fmt.Printf("Error parsing version in %s: %v\n", filePath, err)
		fmt.Printf("Do you want to overwrite it with version %s? (y/n): ", parsedExecVersion)
		overwrite := getOverwriteConfirmation()
		if !overwrite {
			fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
			return false
		}
	}

	if !parsedExistingVersion.Equal(parsedExecVersion) {
		fmt.Printf("File %s already exists with version %s.\nDo you want to overwrite it with version %s? (y/n): ", filePath, parsedExistingVersion, parsedExecVersion)
		overwrite := getOverwriteConfirmation()
		if !overwrite {
			fmt.Printf("Skipping %s. It will not be overwritten.\n", filePath)
			return false
		}
	} else {
		// fmt.Printf("%s already set to version %s. No changes needed.\n", filePath, existingVersion)
		return false
	}
	return true
}

func processVersionFile(filePath string, parsedVersion *semver.Version, forceOverwrite bool) bool {
	overwrite := forceOverwrite
	if !forceOverwrite {
		overwrite = getConfirmationIfFileExists(filePath, parsedVersion)
	}
	if overwrite {
		err := os.WriteFile(filePath, []byte(parsedVersion.String()+"\n"), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", filePath, err)
			os.Exit(1)
		}
		return true
	}
	return false
}

func main() {
	start := time.Now()
	args := []string{}
	forceOverwrite := false
	createNvmrc := false
	modifyPackageJson := false
	if len(os.Args) > 1 {
		args = os.Args[1:]
		for _, arg := range args {
			if arg == "--force" || arg == "-f" {
				forceOverwrite = true
			}
			if arg == "--nvmrc" {
				createNvmrc = true
			}
			if arg == "--engines" || arg == "-e" {
				modifyPackageJson = true
			}
		}
	}

	version := getVersion()
	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing Node.js version:", err)
		os.Exit(1)
	}

	nvmrcSet := false
	nodeversionSet := false
	packageJsonSet := false

	nodeversionSet = processVersionFile(".node-version", parsedVersion, forceOverwrite)

	if createNvmrc {
		nvmrcSet = processVersionFile(".nvmrc", parsedVersion, forceOverwrite)
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
				fmt.Fprintln(os.Stderr, "Error updating package.json engines.node field:", err)
				os.Exit(1)
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
}
