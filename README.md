# pin-node

**Pin your node version to your current directory in milliseconds.**

`pin-node` is a fast, lightweight CLI tool written in Go that helps you pin the Node.js version for your current project directory. It generates or updates version files that are widely recognized by various Node version managers.

I originally built this as a temporary solution because `fnm` (unlike `nvm`, which I came from previously) doesn't automatically create `.node-version` or `.nvmrc` files when running `fnm use`, nor does it have a built-in way to "pin" a version to the current directory.

With an average runtime of **24-50ms**, it seamlessly integrates into your workflow without slowing you down.

## Installation

### Via Releases (Recommended)

You can download pre-built binaries for macOS and Windows from the [latest release](https://github.com/lkekana/pin-node/releases/latest).

#### macOS

1. Download the appropriate archive for your Mac:
   - **Apple Silicon (M1/M2/M3):** `pin-node_Darwin_arm64.tar.gz`
   - **Intel:** `pin-node_Darwin_x86_64.tar.gz`
   
   *You can do this via your browser or directly in the terminal (Apple Silicon example):*
   ```bash
   curl -L -O https://github.com/lkekana/pin-node/releases/latest/download/pin-node_Darwin_arm64.tar.gz
   ```

2. Extract the archive:
   ```bash
   tar -xzf pin-node_Darwin_arm64.tar.gz
   ```

3. Move the binary to your PATH:
   ```bash
   sudo mv pin /usr/local/bin/
   ```

#### Windows

1. Download the appropriate archive for your system (most likely `pin-node_Windows_x86_64.zip`) from the [latest release](https://github.com/lkekana/pin-node/releases/latest).
2. Extract the `.zip` file.
3. Move `pin.exe` to a directory that is included in your system's `PATH` environment variable. 

*Here is a quick way to do this using PowerShell:*
```powershell
# 1. Create a local bin directory in your user profile
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# 2. Move the executable (assuming you are in the extracted folder)
Move-Item .\pin.exe "$env:USERPROFILE\bin\pin.exe"

# 3. Add the directory to your User PATH (if it's not already there)
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$env:USERPROFILE\bin", "User")
    Write-Host "Added to PATH. Please restart your terminal to use 'pin-node'."
}
```

### Via Go Install
If you have Go installed, you can easily install the CLI globally using:
```bash
go install github.com/lkekana/pin-node@latest
```

### From Source
Alternatively, you can clone the repository and run it directly:
```bash
git clone https://github.com/lkekana/pin-node.git
cd pin-node
go run .
```

*Note: Automated builds and precompiled binaries via GitHub Actions are currently in the works and will be available soon!*

## Compatibility

| Version Manager | Version File Supported |
|-----------------|------------------------|
| nvm             | ✅ .nvmrc                 |
| fnm             | ✅ .node-version          |
| Volta           | ✅ package.json `engines.node` field |
| asdf            | ✅ .nvmrc, .node-version (but not .tool-versions) |
| mise            | ✅ .nvmrc, .node-version (but not .tool-versions or mise.toml) |

## Usage

```bash
pin [flags]
```

### Flags
- `-v, --version`: Specify a custom Node.js version to pin (defaults to the currently installed Node.js version).
- `--nvmrc`: Also create or update the `.nvmrc` file.
- `-e, --engines`: Update the `engines.node` field in your `package.json`.
- `-f, --force`: Force overwrite existing version files without interactive confirmation prompts.

### Examples
```bash
# Pin to the currently installed Node.js version (creates .node-version)
pin

# Pin a specific Node.js version
pin -v 18.16.0

# Pin and also update .nvmrc and package.json
pin -v 18.16.0 --nvmrc --engines

# Force overwrite without interactive prompts
pin -v 20.0.0 --force --nvmrc --engines
```

## Design Choices & Limitations

- **No `.tool-versions` support:** I explicitly did not include support for `.tool-versions` (used by `asdf` or `mise`). Parsing it is tedious, and `asdf` already fully supports `.node-version` and `.nvmrc` files anyway. This means coverage for major version managers remains near 100% without the added complexity.
- **`npm` for `package.json` updates:** Instead of manually parsing/editing `package.json` or creating specific implementations for `pnpm` or `yarn`, `pin-node` uses the `npm` CLI to set the `engines.node` field. It is much easier to maintain, and the `npm` overhead adds less than 1 second to the execution time.
- **Strict Semantic Versioning:** The tool uses `github.com/Masterminds/semver` to parse versions. It expects and generates strict semantic versions (e.g., `18.16.0`). It does **not** parse or support non-semantic aliases that `nvm` sometimes uses (like `lts/*`, `node`, `14`, etc.).
- **Go Version Compatibility:** To maximize compatibility for precompiled binaries, the project dependencies (`fatih/color` >= 1.18, `cobra` >= 1.15, `semver` >= 1.21) allow building with older Go versions. I am adjusting the `go.mod` to target Go 1.21 (instead of 1.22.12) so more users can build and use the tool seamlessly.

## Why?

I migrated from `nvm` to `fnm` for its insanely fast performance (thanks to Rust) and modern features, but I missed the convenience of having a simple command to pin my Node version to the current directory. This tool was born out of that frustration and has since evolved into a standalone utility that I hope others will find useful as well.

An hours' worth of work to save minutes of frustration in the future :)

## Roadmap

- [x] **Automated Releases:** Set up GitHub Actions to automatically build and publish cross-platform binaries to GitHub Releases.
- [ ] **Upstream Contribution:** I'm planning on contributing this logic natively to `fnm` so that `fnm use` can optionally write to version files out-of-the-box, eventually making this standalone tool obsolete!

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.