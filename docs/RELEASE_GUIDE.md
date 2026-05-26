# Winitrix Release Guide

This document is for project maintainers. Follow these steps to correctly tag and publish a new version of Winitrix using the automated CI/CD pipeline.

## 1. Prepare the Release
1. Ensure all features and bug fixes for this release are merged into the `main` branch.
2. Update the `CHANGELOG.md` file with the new version number, date, and categorized changes (Added, Changed, Deprecated, Removed, Fixed, Security).
3. Commit the changes:
   ```bash
   git add CHANGELOG.md
   git commit -m "chore: prepare for release vX.Y.Z"
   git push origin main
   ```

## 2. Tag the Release
Winitrix uses Semantic Versioning (`vMAJOR.MINOR.PATCH`). 
You **must** use annotated tags so the GitHub Actions pipeline captures the release metadata correctly.

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

## 3. Verify the CI/CD Pipeline
Once the tag is pushed to GitHub:
1. Navigate to the **Actions** tab in the GitHub repository.
2. You will see the `Release` workflow triggered by the `vX.Y.Z` tag.
3. The workflow will automatically:
   - Build a stripped executable (`winitrix.exe`) with `ldflags` injected.
   - Build a Windows installer (`winitrix-setup-vX.X.X.exe`) using Inno Setup.
   - Zip the executable with the `README.md` and `LICENSE`.
   - Generate a SHA256 `checksums.txt`.
   - Create a GitHub Release and upload all assets.

## 4. Finalize the GitHub Release
1. Navigate to the **Releases** section on GitHub.
2. Edit the newly created release (the Action creates it automatically).
3. Copy the relevant section from `CHANGELOG.md` into the release notes description.
4. Publish the release!

## Rollback Guidance
If a critical error is found immediately after tagging:
1. Delete the remote tag: `git push --delete origin vX.Y.Z`
2. Delete the local tag: `git tag -d vX.Y.Z`
3. Delete the drafted GitHub Release from the repository.
4. Fix the code, merge, and re-tag using a new patch version (e.g., `v1.0.1`).
5. BYE
