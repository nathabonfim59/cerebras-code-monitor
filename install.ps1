# PowerShell installation script for cerebras-code-monitor
param(
    [string]$InstallDir = "$env:USERPROFILE\AppData\Local\Programs\cerebras-monitor",
    [switch]$AddToPath = $true
)

# Script configuration
$RepoOwner = "nathabonfim59"
$RepoName = "cerebras-code-monitor"
$BinaryName = "cerebras-monitor.exe"

# Function to write colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Function to get latest release info
function Get-LatestRelease {
    $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get
        return $response
    }
    catch {
        Write-Error "Failed to fetch release information: $($_.Exception.Message)"
        exit 1
    }
}

# Function to download and extract release
function Install-Release {
    param(
        [string]$TagName,
        [string]$Architecture
    )
    
    $archiveName = "$RepoName-windows-$Architecture.zip"
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$TagName/$archiveName"
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $archivePath = Join-Path $tempDir.FullName $archiveName
    
    try {
        Write-Status "Downloading $archiveName..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath
        
        if (-not (Test-Path $archivePath)) {
            Write-Error "Download failed: $archiveName not found"
            exit 1
        }
        
        Write-Status "Extracting archive..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir.FullName -Force
        
        # Find the binary (it has platform-specific name in archive)
        $archiveBinaryName = "$RepoName-windows-$Architecture.exe"
        $binaryPath = Get-ChildItem -Path $tempDir.FullName -Name $archiveBinaryName -Recurse | Select-Object -First 1
        
        if (-not $binaryPath) {
            Write-Error "Binary $archiveBinaryName not found in archive"
            exit 1
        }
        
        $fullBinaryPath = Join-Path $tempDir.FullName $binaryPath
        
        Write-Status "Installing to $InstallDir..."
        
        # Create install directory if it doesn't exist
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Copy binary
        $destinationPath = Join-Path $InstallDir $BinaryName
        Copy-Item -Path $fullBinaryPath -Destination $destinationPath -Force
        
        Write-Success "$BinaryName installed successfully to $InstallDir"
        
        return $destinationPath
    }
    finally {
        # Cleanup
        Remove-Item -Path $tempDir.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Function to add directory to PATH
function Add-ToPath {
    param([string]$Directory)
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -split ';' -contains $Directory) {
        Write-Status "$Directory is already in PATH"
        return
    }
    
    Write-Status "Adding $Directory to user PATH..."
    $newPath = "$currentPath;$Directory"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    
    # Update current session PATH
    $env:PATH += ";$Directory"
    
    Write-Success "Added $Directory to PATH. Restart your terminal or open a new one for changes to take effect."
}

# Function to check if binary is accessible
function Test-Installation {
    param([string]$BinaryPath)
    
    try {
        $version = & $BinaryPath --version 2>$null
        if ($version) {
            Write-Success "Installation verified! Version: $($version[0])"
            return $true
        }
    }
    catch {
        # Binary exists but might not be in PATH yet
    }
    
    return $false
}

# Main function
function Main {
    Write-Status "Installing $RepoName..."
    
    # Check if running as administrator (optional, but good practice)
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    if (-not $isAdmin) {
        Write-Warning "Not running as administrator. Installing to user directory."
    }
    
    # Detect architecture
    $architecture = Get-Architecture
    Write-Status "Detected architecture: $architecture"
    
    # Get latest release info
    Write-Status "Fetching latest release information..."
    $release = Get-LatestRelease
    
    if (-not $release) {
        Write-Error "Failed to get release information"
        exit 1
    }
    
    $tagName = $release.tag_name
    Write-Status "Latest release: $tagName"
    
    # Check if already installed and up to date
    $binaryPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $binaryPath) {
        try {
            $currentVersion = & $binaryPath --version 2>$null | Select-Object -First 1
            Write-Status "Current version: $currentVersion"
            
            if ($currentVersion -like "*$tagName*") {
                Write-Success "$BinaryName is already up to date ($tagName)"
                exit 0
            }
        }
        catch {
            Write-Status "Could not determine current version, proceeding with installation..."
        }
    }
    
    # Download and install
    $installedBinaryPath = Install-Release -TagName $tagName -Architecture $architecture
    
    # Add to PATH if requested
    if ($AddToPath) {
        Add-ToPath -Directory $InstallDir
    }
    
    # Verify installation
    if (Test-Installation -BinaryPath $installedBinaryPath) {
        Write-Status "Run 'cerebras-monitor --help' to get started."
    }
    else {
        Write-Warning "Installation complete, but binary verification failed."
        Write-Status "You can run it directly: $installedBinaryPath --help"
    }
}

# Run main function
try {
    Main
}
catch {
    Write-Error "Installation failed: $($_.Exception.Message)"
    exit 1
}