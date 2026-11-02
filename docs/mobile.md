# Mobile app

The Android app is the same SvelteKit SPA as the web client, wrapped in a Capacitor shell. The code lives in `apps/client` and the Android project in `apps/client/android`. iOS comes later with the same code.

## How the shell works

The web build (`bun run build`) is copied into the app and served from `https://localhost` inside a WebView. On the phone the API base is `https://bulle.facile.studio/api`. Set `VITE_API_BASE` at build time to point somewhere else.

The web build is unchanged: cookies, same origin, `X-Facile-CSRF` header. The native build sends `Authorization: Bearer <token>` instead and uses no cookies and no CSRF header. The token lives in Capacitor Preferences. A 401 clears it and sends you to the login page.

## Login

1. The app creates a random verifier and state, and keeps them in memory.
2. It opens the system browser at `/api/auth/oidc?flow=mobile&challenge=<sha256 of verifier>&state=<state>`.
3. After sign-in, porte redirects to `studio.facile.bulle://auth?code=...&state=...`.
4. The app checks the state, then posts `{code, verifier}` to `/api/auth/oidc/exchange` and stores the returned token.

The other deep link is `studio.facile.bulle://sessions/<id>`. A notification tap uses it to open a session.

## Push notifications

Push goes through ntfy, with no Google services. The Account page shows the server and topic when the server has push enabled. Install the ntfy app, subscribe to that topic, then use "Send test". Jobs with delivery set to "Push notification" publish to the topic when they finish.

## Build the APK

Needs the Android SDK (`ANDROID_HOME`) and Java 21.

```
mise run //apps/client:mobile:sync
mise run //apps/client:mobile:apk
```

The APK lands in `apps/client/android/app/build/outputs/apk/release/`. It is one universal APK, with no ABI splits. For a quick debug build, run `./gradlew assembleDebug` in `apps/client/android`.

The Capacitor CLI wants Node 22. The tasks run it through Bun (`bun --bun x cap`), so Node is not needed.

## Environment

| Name | Use | Default |
| --- | --- | --- |
| `BULLE_VERSION_NAME` | `versionName` | `0.0.0` |
| `BULLE_VERSION_CODE` | `versionCode`, an integer | `1` |
| `ANDROID_KEYSTORE_PATH` | Keystore file. Release signing is on only when this is set | unset |
| `ANDROID_KEYSTORE_PASSWORD` | Keystore password | unset |
| `ANDROID_KEY_ALIAS` | Key alias | unset |
| `ANDROID_KEY_PASSWORD` | Key password | unset |
| `VITE_API_BASE` | API base for the web build | `/api` or the bulle.facile.studio URL |

Without a keystore, `assembleRelease` produces an unsigned APK.
