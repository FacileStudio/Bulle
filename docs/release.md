# Release

Push a `v*` tag. `.github/workflows/release.yml` runs two jobs.

1. `release`: GoReleaser builds the CLI, creates the GitHub release and updates the Homebrew tap.
2. `android`: runs after `release` (`needs: release`), so a failed APK build never blocks the CLI assets. It builds a signed universal APK and uploads two files to the release:
   - `bulle-android.apk`
   - `bulle-android.apk.sha256`

The names are fixed. `https://github.com/FacileStudio/Bulle/releases/latest/download/bulle-android.apk` is a stable URL.

## Version

`BULLE_VERSION_NAME` is the tag without the leading `v` (`v0.95.0` gives `0.95.0`).

`BULLE_VERSION_CODE` is `major * 1000000 + minor * 1000 + patch`, computed from the tag after dropping any `-rc1` or `+build` suffix. `v0.95.0` gives `95000`. It is valid up to minor 999 and patch 999, and it never goes down as long as tags only go up. A prerelease and its final release share a code, so ship finals only, or bump the patch.

## Secrets

GitHub Actions secrets read by the `android` job:

| Secret | Content |
| --- | --- |
| `ANDROID_KEYSTORE_BASE64` | the keystore file, base64 encoded on one line (`base64 -w0 bulle.keystore`) |
| `ANDROID_KEYSTORE_PASSWORD` | keystore password |
| `ANDROID_KEY_ALIAS` | key alias |
| `ANDROID_KEY_PASSWORD` | key password |

The job fails at its first step if any of them is empty. It decodes the keystore to `$RUNNER_TEMP/bulle.keystore`, exports `ANDROID_KEYSTORE_PATH` for gradle, and deletes the file at the end even on failure.

Tiroir keys that hold the same values:

- `BULLE_ANDROID_KEYSTORE_BASE64`
- `BULLE_ANDROID_KEYSTORE_PASSWORD`
- `BULLE_ANDROID_KEY_ALIAS`
- `BULLE_ANDROID_KEY_PASSWORD`

## Rotation

You cannot rotate the key. Android refuses to update an app signed with a different key, so every user would have to uninstall first and lose local data. Use the same keystore forever. Back it up in at least two places (tiroir plus an offline copy). Losing it means a new app id.
