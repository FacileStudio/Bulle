use keyring::Entry;

const SERVICE: &str = "studio.facile.bulle";
const ACCOUNT: &str = "session";

fn entry() -> Result<Entry, String> {
    Entry::new(SERVICE, ACCOUNT).map_err(|err| err.to_string())
}

/// The session token from the OS keychain, or `None` when none is stored.
#[tauri::command]
pub fn token_get() -> Result<Option<String>, String> {
    match entry()?.get_password() {
        Ok(token) => Ok(Some(token)),
        Err(keyring::Error::NoEntry) => Ok(None),
        Err(err) => Err(err.to_string()),
    }
}

/// Stores the session token in the OS keychain (Keychain, Credential Manager, Secret Service).
#[tauri::command]
pub fn token_set(token: String) -> Result<(), String> {
    entry()?.set_password(&token).map_err(|err| err.to_string())
}

/// Removes the session token from the OS keychain.
#[tauri::command]
pub fn token_clear() -> Result<(), String> {
    match entry()?.delete_credential() {
        Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
        Err(err) => Err(err.to_string()),
    }
}
