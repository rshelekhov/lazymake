# Release Process

This document explains how to set up and use the automated release system for lazymake.

## Initial Setup (One-time)

### 1. Create Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` in your account:
```bash
# On GitHub, create repository: rshelekhov/homebrew-tap
```

### 2. Set Up GitHub Secrets

Go to your `lazymake` repository settings → Secrets and variables → Actions, and add:

- `HOMEBREW_TAP_GITHUB_TOKEN`: A GitHub Personal Access Token with `repo` scope
  - Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
  - Generate new token with `repo` scope
  - Add it as a repository secret

### 3. Install GoReleaser (for local testing)

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest
```

## Making a Release

### 1. Test locally first
```bash
# Test the release process without publishing
make snapshot

# This creates a dist/ folder with all build artifacts
# Check that binaries work correctly
./dist/lazymake_darwin_amd64_v1/lazymake --help
```

### 2. Create and push a version tag
```bash
# Make sure all changes are committed
git add .
git commit -m "Prepare for release"

# Create a tag (use semantic versioning)
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to GitHub
git push origin v0.1.0
```

### 3. Automatic release happens

Once you push the tag, GitHub Actions will:
1. Build binaries for Linux, macOS, Windows (amd64 and arm64)
2. Create .deb, .rpm, and .apk packages
3. Upload everything to GitHub Releases
4. Update the Homebrew tap with the new formula

### 4. Verify the release

- Check GitHub Actions tab for workflow status
- Check GitHub Releases page for published release
- Check `homebrew-tap` repository for updated formula
- Test installation:
  ```bash
  brew install rshelekhov/tap/lazymake
  lazymake --version
  ```

## What Gets Published

When you create a release, GoReleaser automatically creates:

1. **Binaries** for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)

2. **Packages**:
   - `.deb` (Debian/Ubuntu)
   - `.rpm` (RedHat/Fedora/CentOS)
   - `.apk` (Alpine)

3. **Homebrew Formula**:
   - Automatically pushed to `rshelekhov/homebrew-tap`
   - Users can install with `brew install rshelekhov/tap/lazymake`

4. **Archive files**:
   - `.tar.gz` for Linux/macOS
   - `.zip` for Windows

## Troubleshooting

### Release fails on GitHub Actions
- Check that `HOMEBREW_TAP_GITHUB_TOKEN` secret is set correctly
- Verify the token has `repo` scope
- Check that `homebrew-tap` repository exists

### Homebrew formula not updated
- Ensure the GitHub token has write access to `homebrew-tap` repository
- Check GoReleaser logs in GitHub Actions

### Local snapshot build fails
- Run `go mod tidy` to ensure dependencies are correct
- Check that GoReleaser is installed: `goreleaser --version`

## Release Checklist

- [ ] All tests pass: `make test`
- [ ] Local snapshot works: `make snapshot`
- [ ] CHANGELOG updated (if you have one)
- [ ] README updated with correct version numbers
- [ ] Version tag follows semantic versioning (v0.1.0, v1.2.3, etc.)
- [ ] Tag pushed to GitHub: `git push origin v0.1.0`
- [ ] GitHub Actions workflow completed successfully
- [ ] Release appears on GitHub Releases page
- [ ] Homebrew formula updated in tap repository
- [ ] Installation tested: `brew install rshelekhov/tap/lazymake`
