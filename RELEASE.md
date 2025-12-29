## Release Management

This project uses Semantic Versioning (Major.Minor.Patch).

### Bumping Version
To create a new release, ensure your working directory is clean (no uncommitted changes), then run one of the following commands:

```bash
# Bump patch version (e.g., 1.0.0 -> 1.0.1)
make bump part=patch

# Bump minor version (e.g., 1.0.0 -> 1.1.0)
make bump part=minor

# Bump major version (e.g., 1.0.0 -> 2.0.0)
make bump part=major
```

This command will:
1. Verify the working directory is clean.
2. Increment the version number in the `VERSION` file.
3. Commit the change.
4. Create a git tag (e.g., `v1.0.1`).
5. Push the commit and the tag to the remote repository.

### Automated Release
Pushing the tag triggers a GitHub Action workflow that:
1. Runs unit tests and configuration tests.
2. Verifies the tag matches the `VERSION` file.
3. Cross-compiles binaries for Linux, Windows, and macOS.
4. Creates a GitHub Release with the binaries and an automatically generated changelog.