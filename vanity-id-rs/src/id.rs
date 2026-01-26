//! Core vanity ID generation functionality.
//!
//! This module provides the fundamental ID generation algorithms including:
//! - Vanity ID generation from seed strings
//! - Random Base32 Crockford ID generation
//! - Unicode normalization and cleaning
//! - Configuration management

use once_cell::sync::Lazy;
use parking_lot::RwLock;
use rand::Rng;
use sha2::{Digest, Sha256};
use std::collections::HashSet;
use unicode_normalization::UnicodeNormalization;

/// Custom alphabet that's human-readable and avoids confusing characters.
/// Excludes: 1 (looks like I), 0 (looks like O)
/// Includes: I, L, O (user preference), all other clear letters and numbers
pub const CUSTOM_ALPHABET: &str = "23456789ABCDEFGHIJKLMNOPQRSTUVWXYZ";

/// EntityKind represents different types of entities that can have vanity IDs
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum EntityKind {
    HumanContact,
    Organization,
    Role,
}

impl std::fmt::Display for EntityKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EntityKind::HumanContact => write!(f, "contact"),
            EntityKind::Organization => write!(f, "organization"),
            EntityKind::Role => write!(f, "role"),
        }
    }
}

impl EntityKind {
    /// Returns the minimum vanity ID length for this entity kind
    pub fn get_min_length(&self) -> usize {
        let cfg = get_config();
        match self {
            EntityKind::HumanContact => cfg.contact.min_length,
            EntityKind::Organization => cfg.organization.min_length,
            EntityKind::Role => cfg.role.min_length,
        }
    }

    /// Returns the maximum vanity ID length for this entity kind
    pub fn get_max_length(&self) -> usize {
        let cfg = get_config();
        match self {
            EntityKind::HumanContact => cfg.contact.max_length,
            EntityKind::Organization => cfg.organization.max_length,
            EntityKind::Role => cfg.role.max_length,
        }
    }
}

/// EntityLengthConfig defines min/max length constraints for an entity type
#[derive(Debug, Clone)]
pub struct EntityLengthConfig {
    pub min_length: usize,
    pub max_length: usize,
}

/// Config holds ID generation configuration.
#[derive(Debug, Clone)]
pub struct Config {
    pub max_retries: usize,
    pub contact: EntityLengthConfig,
    pub organization: EntityLengthConfig,
    pub role: EntityLengthConfig,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            max_retries: 128,
            contact: EntityLengthConfig {
                min_length: 4,
                max_length: 9,
            },
            organization: EntityLengthConfig {
                min_length: 3,
                max_length: 9,
            },
            role: EntityLengthConfig {
                min_length: 2,
                max_length: 10,
            },
        }
    }
}

/// Global configuration
static PKG_CONFIG: Lazy<RwLock<Option<Config>>> = Lazy::new(|| RwLock::new(None));

/// Sets the package-level ID generation configuration.
/// Pass None to reset to defaults. This is safe for concurrent use.
pub fn set_config(cfg: Option<Config>) {
    let mut config = PKG_CONFIG.write();
    *config = cfg;
}

/// Gets the current configuration, using defaults if not set.
pub fn get_config() -> Config {
    let config = PKG_CONFIG.read();
    config.clone().unwrap_or_default()
}

/// Returns the maximum number of collision retry attempts
pub fn get_max_retries() -> usize {
    get_config().max_retries
}

/// Error type for ID generation operations
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum IdError {
    EmptySeed,
    NormalizationFailed(String),
    GenerationFailed(String),
    MaxRetriesExceeded(usize),
    InvalidLength(String),
}

impl std::fmt::Display for IdError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            IdError::EmptySeed => write!(f, "seed cannot be empty"),
            IdError::NormalizationFailed(msg) => write!(f, "failed to normalize seed: {}", msg),
            IdError::GenerationFailed(msg) => write!(f, "failed to generate ID: {}", msg),
            IdError::MaxRetriesExceeded(n) => write!(f, "failed to assign ID after {} retries", n),
            IdError::InvalidLength(msg) => write!(f, "invalid length: {}", msg),
        }
    }
}

impl std::error::Error for IdError {}

/// ReservedWordsChecker checks if a seed matches reserved words
pub trait ReservedWordsChecker: Send + Sync {
    fn is_reserved(&self, seed: &str) -> Result<bool, IdError>;
}

/// HardcodedReservedChecker implements ReservedWordsChecker using hardcoded reserved words
pub struct HardcodedReservedChecker {
    reserved_words: HashSet<String>,
}

impl HardcodedReservedChecker {
    /// Creates a new hardcoded reserved words checker
    pub fn new() -> Self {
        let reserved_words: Vec<&str> = vec![
            // System words
            "ADMIN", "ROOT", "TEST", "NULL",
            // Common system terms
            "USER", "API", "CONFIG", "DATA", "SYSTEM", "PUBLIC", "PRIVATE",
            "GUEST", "DEMO", "TEMP",
            // Database terms
            "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "TABLE", "INDEX",
            // HTTP/Web terms
            "GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS",
            // Common reserved terms
            "ABOUT", "HELP", "CONTACT", "SUPPORT", "FAQ", "LOGIN", "LOGOUT",
            "SIGNUP", "REGISTER", "PROFILE", "SETTINGS", "SEARCH",
            // Technical terms
            "ERROR", "SUCCESS", "FAILED", "PENDING", "ACTIVE", "INACTIVE",
            "DELETED", "ARCHIVE",
        ];

        let word_set: HashSet<String> = reserved_words
            .into_iter()
            .map(|s| s.to_uppercase())
            .collect();

        Self {
            reserved_words: word_set,
        }
    }
}

impl Default for HardcodedReservedChecker {
    fn default() -> Self {
        Self::new()
    }
}

impl ReservedWordsChecker for HardcodedReservedChecker {
    fn is_reserved(&self, seed: &str) -> Result<bool, IdError> {
        Ok(self.reserved_words.contains(&seed.to_uppercase()))
    }
}

/// Generates a vanity ID from a seed string using legacy constraints (defaults to HumanContact)
pub fn generate_vanity_id(seed: &str) -> Result<String, IdError> {
    generate_vanity_id_for_entity(seed, EntityKind::HumanContact)
}

/// Generates a vanity ID from a seed string with entity-specific constraints
pub fn generate_vanity_id_for_entity(seed: &str, kind: EntityKind) -> Result<String, IdError> {
    if seed.is_empty() {
        return Err(IdError::EmptySeed);
    }

    let min_length = kind.get_min_length();
    let max_length = kind.get_max_length();

    // Step 1: Normalize and transliterate
    let normalized = normalize_to_ascii(seed)?;

    // Step 2: Clean and filter
    let mut cleaned = clean_seed(&normalized);
    if cleaned.len() < 3 {
        // Step 3: Pad with consonants from hash if too short
        cleaned = pad_with_consonants(&cleaned, seed);
    }

    // Ensure minimum length for this entity type
    if cleaned.len() < min_length {
        cleaned = pad_to_min_length_for_entity(&cleaned, min_length);
    }

    // Ensure maximum length for this entity type
    if cleaned.len() > max_length {
        cleaned = cleaned[..max_length].to_string();
    }

    Ok(cleaned)
}

/// Generates a random Base32 Crockford ID of specified length
pub fn generate_random_id(n: usize) -> Result<String, IdError> {
    if n < 1 {
        return Err(IdError::InvalidLength("length must be positive".to_string()));
    }

    let mut rng = rand::thread_rng();
    let bytes_needed = (n * 5 + 7) / 8;
    let bytes: Vec<u8> = (0..bytes_needed).map(|_| rng.gen()).collect();

    let alphabet = CUSTOM_ALPHABET.as_bytes();
    let mut result = Vec::with_capacity(n);
    let mut bit_buffer: u64 = 0;
    let mut bit_count: u32 = 0;
    let mut byte_idx = 0;

    for _ in 0..n {
        // Fill buffer if needed
        while bit_count < 5 && byte_idx < bytes.len() {
            bit_buffer = (bit_buffer << 8) | (bytes[byte_idx] as u64);
            bit_count += 8;
            byte_idx += 1;
        }

        if bit_count < 5 {
            // Not enough bits, pad with zeros
            bit_buffer <<= 5 - bit_count;
            bit_count = 5;
        }

        // Extract 5 bits
        let idx = ((bit_buffer >> (bit_count - 5)) & 31) as usize;
        result.push(alphabet[idx]);
        bit_count -= 5;
    }

    Ok(String::from_utf8(result).unwrap())
}

/// Tries to assign an ID based on entity kind and seed
pub fn assign_id<F>(
    kind: EntityKind,
    seed: &str,
    checker: Option<&dyn ReservedWordsChecker>,
    mut put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    // For human contacts, organizations, and roles, try vanity ID first
    if !seed.is_empty() {
        if let Ok(vanity_id) = try_vanity_id_for_entity(seed, kind, checker, &mut put) {
            return Ok(vanity_id);
        }
    }

    // Fallback to random ID using entity-specific minimum length
    let min_len = kind.get_min_length();
    for _ in 0..get_max_retries() {
        let random_id = generate_random_id(min_len)?;

        if put(&random_id).is_ok() {
            return Ok(random_id);
        }
        // If it's a conflict error, retry
    }

    Err(IdError::MaxRetriesExceeded(get_max_retries()))
}

/// Normalizes Unicode and transliterates to ASCII
pub fn normalize_to_ascii(s: &str) -> Result<String, IdError> {
    // NFKD normalization - decompose compatibility characters
    let decomposed: String = s.nfkd().collect();

    // Remove combining diacritical marks (Mn category)
    let result: String = decomposed
        .chars()
        .filter(|c| !unicode_normalization::char::is_combining_mark(*c))
        .collect();

    // Recompose using NFC
    let result: String = result.nfc().collect();

    // Simple ASCII transliteration for common cases
    let result = result
        .replace('ß', "ss")
        .replace('æ', "ae")
        .replace('Æ', "AE")
        .replace('œ', "oe")
        .replace('Œ', "OE");

    Ok(result)
}

/// Cleans the seed according to the algorithm
pub fn clean_seed(s: &str) -> String {
    // Convert to uppercase
    let s = s.to_uppercase();

    // Keep only alphanumeric characters
    let mut result = String::new();
    for c in s.chars() {
        if c.is_ascii_alphanumeric() {
            result.push(c);
        }
    }

    // Collapse consecutive repeats
    if result.len() > 1 {
        let mut collapsed = String::new();
        let mut prev: Option<char> = None;
        for c in result.chars() {
            if Some(c) != prev {
                collapsed.push(c);
                prev = Some(c);
            }
        }
        result = collapsed;
    }

    // Strip leading digits
    result = result.trim_start_matches(|c: char| c.is_ascii_digit()).to_string();

    // Now convert to custom alphabet by mapping excluded chars
    convert_to_custom_alphabet(&result)
}

/// Converts characters to valid custom alphabet
pub fn convert_to_custom_alphabet(s: &str) -> String {
    let mut result = String::new();
    for c in s.chars() {
        match c {
            '0' => result.push('O'), // 0 -> O (since we exclude 0 but include O)
            '1' => result.push('I'), // 1 -> I (since we exclude 1 but include I)
            c if CUSTOM_ALPHABET.contains(c) => result.push(c),
            _ => {} // Skip characters not in alphabet
        }
    }
    result
}

/// Normalizes user input for ID lookups.
/// Converts to uppercase, maps confusing characters (0→O, 1→I),
/// and strips invalid characters.
pub fn normalize_for_lookup(input: &str) -> String {
    convert_to_custom_alphabet(&input.to_uppercase())
}

/// Pads the cleaned seed with consonants from the original seed hash
fn pad_with_consonants(cleaned: &str, original: &str) -> String {
    const CONSONANTS: &str = "BCDFGHJKMNPQRSTVWXYZ";

    // Generate deterministic consonants from hash
    let mut hasher = Sha256::new();
    hasher.update(original.as_bytes());
    let hash = hasher.finalize();

    let mut result = cleaned.to_string();
    let consonants: Vec<char> = CONSONANTS.chars().collect();

    for &byte in hash.iter() {
        if result.len() >= 3 {
            break;
        }
        let consonant_idx = (byte as usize) % consonants.len();
        result.push(consonants[consonant_idx]);
    }

    result
}

/// Pads to minimum vanity length (legacy function)
pub fn pad_to_min_length(s: &str) -> String {
    pad_to_min_length_for_entity(s, EntityKind::HumanContact.get_min_length())
}

/// Pads to specified minimum length
pub fn pad_to_min_length_for_entity(s: &str, min_length: usize) -> String {
    if s.len() >= min_length {
        return s.to_string();
    }

    let mut result = s.to_string();

    // Pad with pattern from the string itself or default pattern
    let padding = if !s.is_empty() {
        // Use last character as padding pattern
        let last_char = s.chars().last().unwrap();
        std::iter::repeat(last_char)
            .take(min_length - s.len())
            .collect::<String>()
    } else {
        "ABCD"[..(min_length - s.len()).min(4)].to_string()
    };

    result.push_str(&padding[..(min_length - s.len()).min(padding.len())]);
    result
}

/// Pads to minimum length with names intelligently using letters from original names
pub fn pad_to_min_length_with_names(s: &str, min_length: usize, first_name: &str, last_name: &str) -> String {
    if s.len() >= min_length {
        return s.to_string();
    }

    let needed = min_length - s.len();

    // Create a pool of available characters from the original names
    let mut name_pool: Vec<char> = Vec::new();

    // Add characters from both names, removing what's already used in the seed
    let all_names = format!("{}{}", first_name, last_name).to_uppercase();
    let used_in_seed = s.to_uppercase();

    for c in all_names.chars() {
        // Only add valid Crockford Base32 characters that aren't already heavily used in seed
        if CUSTOM_ALPHABET.contains(c) {
            let count_in_seed = used_in_seed.matches(c).count();
            let count_in_names = all_names.matches(c).count();

            // Add character if it's not over-represented in the seed
            if count_in_seed < count_in_names {
                name_pool.push(c);
            }
        }
    }

    // If no valid characters from names, fall back to original logic
    if name_pool.is_empty() {
        let fallback_padding = if !s.is_empty() {
            let last_char = s.chars().last().unwrap();
            std::iter::repeat(last_char).take(needed).collect::<String>()
        } else {
            "ABCD"[..needed.min(4)].to_string()
        };
        return format!("{}{}", s, &fallback_padding[..needed.min(fallback_padding.len())]);
    }

    // If we have unused characters from names, use them for padding
    let mut padding = String::new();
    let mut pool_index = 0;

    for _ in 0..needed {
        if pool_index < name_pool.len() {
            padding.push(name_pool[pool_index]);
            pool_index += 1;
        } else {
            // Cycle through available characters
            padding.push(name_pool[pool_index % name_pool.len()]);
            pool_index += 1;
        }
    }

    format!("{}{}", s, padding)
}

/// Attempts to assign a vanity ID with entity-specific collision handling
fn try_vanity_id_for_entity<F>(
    seed: &str,
    kind: EntityKind,
    checker: Option<&dyn ReservedWordsChecker>,
    put: &mut F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let base_id = generate_vanity_id_for_entity(seed, kind)?;
    let min_length = kind.get_min_length();
    let max_length = kind.get_max_length();

    // For organizations and roles, try progressively longer lengths first, then numeric suffixes
    if kind == EntityKind::Organization || kind == EntityKind::Role {
        let mut attempted: HashSet<String> = HashSet::new();

        // Phase 1: Try all lengths first without numeric suffixes
        for length in min_length..=max_length {
            let mut candidate_base = base_id.clone();
            if candidate_base.len() < length {
                candidate_base = pad_to_min_length_for_entity(&candidate_base, length);
            }
            if candidate_base.len() > length {
                candidate_base = candidate_base[..length].to_string();
            }

            // Skip if we've already attempted this ID
            if attempted.contains(&candidate_base) {
                continue;
            }
            attempted.insert(candidate_base.clone());

            // Check reserved words
            if let Some(checker) = checker {
                if checker.is_reserved(&candidate_base)? {
                    continue;
                }
            }

            // Try the base candidate
            if put(&candidate_base).is_ok() {
                return Ok(candidate_base);
            }
        }

        // Phase 2: If all base lengths fail, try numeric suffixes
        for length in (min_length..=max_length).rev() {
            let mut candidate_base = base_id.clone();
            if candidate_base.len() < length {
                candidate_base = pad_to_min_length_for_entity(&candidate_base, length);
            }
            if candidate_base.len() > length {
                candidate_base = candidate_base[..length].to_string();
            }

            // Try numeric suffixes
            if candidate_base.len() >= 2 {
                let base_for_suffix = &candidate_base[..candidate_base.len() - 1];
                for i in 2..=9 {
                    let candidate_id = format!("{}{}", base_for_suffix, i);

                    if attempted.contains(&candidate_id) {
                        continue;
                    }
                    attempted.insert(candidate_id.clone());

                    if let Some(checker) = checker {
                        if checker.is_reserved(&candidate_id)? {
                            continue;
                        }
                    }

                    if put(&candidate_id).is_ok() {
                        return Ok(candidate_id);
                    }
                }
            }
        }

        return Err(IdError::GenerationFailed(format!(
            "failed to assign {} vanity ID within length range {}-{}",
            kind, min_length, max_length
        )));
    }

    // For contacts, try progressively longer lengths
    for length in min_length..=base_id.len().min(max_length) {
        let mut candidate_base = base_id.clone();
        if candidate_base.len() > length {
            candidate_base = candidate_base[..length].to_string();
        }

        // Check reserved words
        if let Some(checker) = checker {
            if checker.is_reserved(&candidate_base)? {
                continue;
            }
        }

        // Try the base candidate at this length
        if put(&candidate_base).is_ok() {
            return Ok(candidate_base);
        }

        // Try with vowel-based collision resolution at this length
        let vowel_variations = generate_vowel_variations(&candidate_base, max_length);
        for candidate_id in vowel_variations {
            if let Some(checker) = checker {
                if checker.is_reserved(&candidate_id)? {
                    continue;
                }
            }

            if put(&candidate_id).is_ok() {
                return Ok(candidate_id);
            }
        }
    }

    Err(IdError::GenerationFailed(format!(
        "failed to assign vanity ID after trying all lengths up to {}",
        max_length
    )))
}

/// Generates name-aware variations by reconstructing the original name
fn generate_vowel_variations(base_id: &str, max_length: usize) -> Vec<String> {
    if base_id.is_empty() {
        return vec![];
    }

    let mut variations = Vec::new();

    // Strategy 1: Progressive vowel insertion to reconstruct name-like forms
    let name_reconstruction = generate_name_reconstruction_variations(base_id, max_length);
    variations.extend(name_reconstruction);

    // Strategy 2: Fallback to generic vowel patterns if needed
    if variations.len() < 10 {
        let generic = generate_generic_vowel_variations(base_id, max_length);
        variations.extend(generic);
    }

    variations
}

/// Creates variations that progressively add vowels
fn generate_name_reconstruction_variations(base_id: &str, max_length: usize) -> Vec<String> {
    let mut variations = Vec::new();
    let vowels = ["I", "A", "E", "O", "U"];

    // Strategy 1: Smart insertion at natural positions
    for vowel in &vowels {
        // Insert at position 1 (after first consonant)
        if base_id.len() >= 1 && base_id.len() + 1 <= max_length {
            let variation = format!("{}{}{}", &base_id[..1], vowel, &base_id[1..]);
            if variation.len() <= max_length {
                variations.push(variation);
            }
        }

        // Insert at position 2 (after second consonant)
        if base_id.len() >= 2 && base_id.len() + 1 <= max_length {
            let variation = format!("{}{}{}", &base_id[..2], vowel, &base_id[2..]);
            if variation.len() <= max_length {
                variations.push(variation);
            }
        }
    }

    // Strategy 2: Replace consonants with vowels for shorter forms
    if base_id.len() >= 3 {
        for vowel in vowels.iter().take(3) {
            if base_id.len() > 2 {
                let variation = format!("{}{}{}", &base_id[..2], vowel, &base_id[3..]);
                if variation.len() <= max_length {
                    variations.push(variation);
                }
            }
        }
    }

    // Strategy 3: Create truncated + vowel versions
    if base_id.len() > 4 {
        let truncated = &base_id[..4];
        for vowel in vowels.iter().take(3) {
            let variation = format!("{}{}", truncated, vowel);
            if variation.len() <= max_length && variation.len() >= 5 {
                variations.push(variation);
            }
        }
    }

    variations
}

/// Provides fallback vowel patterns
fn generate_generic_vowel_variations(base_id: &str, max_length: usize) -> Vec<String> {
    let mut variations = Vec::new();
    let primary_vowels = ["A", "E", "I", "O", "U"];

    // Append vowels
    for vowel in &primary_vowels {
        if base_id.len() + 1 <= max_length {
            variations.push(format!("{}{}", base_id, vowel));
        }
    }

    // Insert vowels between consonant clusters
    for vowel in &primary_vowels {
        let insert_variations = insert_vowel_at_natural_positions(base_id, vowel, max_length);
        variations.extend(insert_variations);
    }

    variations
}

/// Finds good places to insert vowels
fn insert_vowel_at_natural_positions(base_id: &str, vowel: &str, max_length: usize) -> Vec<String> {
    if base_id.len() + 1 > max_length {
        return vec![];
    }

    let mut variations = Vec::new();
    let chars: Vec<char> = base_id.chars().collect();

    for i in 1..chars.len() {
        // Check if we have two consonants next to each other
        if !is_vowel_char(chars[i - 1]) && !is_vowel_char(chars[i]) {
            let new_id = format!("{}{}{}", &base_id[..i], vowel, &base_id[i..]);
            if new_id.len() <= max_length {
                variations.push(new_id);
            }
        }
    }

    variations
}

/// Checks if a character is a vowel
pub fn is_vowel(s: &str) -> bool {
    if s.is_empty() {
        return false;
    }
    is_vowel_char(s.chars().next().unwrap().to_ascii_uppercase())
}

/// Checks if a character is a vowel (more efficient for char operations)
pub fn is_vowel_char(c: char) -> bool {
    matches!(c.to_ascii_uppercase(), 'A' | 'E' | 'I' | 'O' | 'U' | 'Y')
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_random_id() {
        let id = generate_random_id(5).unwrap();
        assert_eq!(id.len(), 5);

        // All characters should be in the custom alphabet
        for c in id.chars() {
            assert!(CUSTOM_ALPHABET.contains(c));
        }
    }

    #[test]
    fn test_clean_seed() {
        assert_eq!(clean_seed("Hello World"), "HELOWORLD");
        assert_eq!(clean_seed("AABBCC"), "ABC");
        assert_eq!(clean_seed("123ABC"), "ABC");
        assert_eq!(clean_seed("A0B1C"), "AOBIC");
    }

    #[test]
    fn test_convert_to_custom_alphabet() {
        assert_eq!(convert_to_custom_alphabet("A0B1C"), "AOBIC");
        assert_eq!(convert_to_custom_alphabet("HELLO"), "HELLO");
    }

    #[test]
    fn test_normalize_for_lookup() {
        assert_eq!(normalize_for_lookup("jd0e"), "JDOE");
        assert_eq!(normalize_for_lookup("JD1E"), "JDIE");
    }

    #[test]
    fn test_hardcoded_reserved_checker() {
        let checker = HardcodedReservedChecker::new();
        assert!(checker.is_reserved("ADMIN").unwrap());
        assert!(checker.is_reserved("admin").unwrap());
        assert!(!checker.is_reserved("JDOE").unwrap());
    }

    #[test]
    fn test_generate_vanity_id() {
        let id = generate_vanity_id("John Doe").unwrap();
        assert!(id.len() >= 4);
        assert!(id.len() <= 9);
    }

    #[test]
    fn test_is_vowel() {
        assert!(is_vowel("A"));
        assert!(is_vowel("e"));
        assert!(!is_vowel("B"));
        assert!(!is_vowel(""));
    }
}
