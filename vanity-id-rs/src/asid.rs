//! Application-Scoped ID (ASID) generation functionality.
//!
//! This module provides functions for generating structured 32-character identifiers
//! with semantic vanity components mixed with random segments for uniqueness.
//!
//! ASID Format: prefix(2) + random(2) + subject(5) + random(2) + predicate(7) + random(2) + context(5) + random(4) + actor(3)
//! Example: "AS47ACMEX22ENGINEE33SFBAY5E7AJOB"

use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::id::IdError;

/// Generates a Job Description ASID
/// Format: JD + random(2) + company(5) + random(2) + role(7) + random(2) + location(5) + random(4) + type(3)
/// Returns an ID like "JD47ACMEX22ENGINEE33SFBAY5E7AJOB"
pub fn generate_jd_asid(company: &str, role_title: &str, location: &str) -> Result<String, IdError> {
    let u = Uuid::new_v4();
    let uuid_str: String = u.to_string().replace('-', "").to_uppercase();

    // Extract random segments: 2 + 2 + 2 + 4 = 10 chars total
    let rand1 = &uuid_str[0..2];
    let rand2 = &uuid_str[2..4];
    let rand3 = &uuid_str[4..6];
    let rand4 = &uuid_str[6..10];

    // Extract vanity components
    let company_part = extract_vanity_component(company, 5);
    let role_part = extract_vanity_component(role_title, 7);
    let location_part = extract_vanity_component(location, 5);
    let type_part = "JOB"; // Default type indicator for job descriptions

    // Construct the JD ASID - 32 characters total
    Ok(format!(
        "JD{}{}{}{}{}{}{}{}",
        rand1, company_part, rand2, role_part, rand3, location_part, rand4, type_part
    ))
}

/// Generates a JD ASID with collision detection
pub fn generate_jd_asid_with_retry<F>(
    company: &str,
    role_title: &str,
    location: &str,
    check_exists: F,
) -> Result<String, IdError>
where
    F: Fn(&str) -> bool,
{
    const MAX_RETRIES: usize = 10;

    for _ in 0..MAX_RETRIES {
        let jd_id = generate_jd_asid(company, role_title, location)?;

        if !check_exists(&jd_id) {
            return Ok(jd_id);
        }
        // Collision detected, retry
    }

    Err(IdError::GenerationFailed(format!(
        "failed to generate unique JD ASID after {} attempts",
        MAX_RETRIES
    )))
}

/// Extracts the specified number of characters from a string, using hash-based padding
fn extract_vanity_component(s: &str, length: usize) -> String {
    if s.is_empty() {
        return generate_hash_based_padding("", length);
    }

    // Clean and uppercase the string, keeping only alphanumeric characters
    let cleaned: String = s
        .to_uppercase()
        .chars()
        .filter(|c| c.is_ascii_alphanumeric())
        .collect();

    if cleaned.is_empty() {
        return generate_hash_based_padding(s, length);
    }

    // Take first N characters, pad with hash-based characters if needed
    if cleaned.len() >= length {
        return cleaned[..length].to_string();
    }

    // Need padding - use hash-based approach
    let needed = length - cleaned.len();
    let padding = generate_hash_based_padding(s, needed);
    format!("{}{}", cleaned, padding)
}

/// Generates deterministic padding characters from string hash
fn generate_hash_based_padding(s: &str, length: usize) -> String {
    if length == 0 {
        return String::new();
    }

    // Valid characters for vanity components (alphanumeric, excluding confusing ones)
    const VALID_CHARS: &str = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"; // No I,O,0,1 for clarity

    let mut hasher = Sha256::new();
    hasher.update(s.as_bytes());
    let hash = hasher.finalize();

    let valid_chars: Vec<char> = VALID_CHARS.chars().collect();
    let mut result = String::with_capacity(length);

    for i in 0..length {
        let byte_index = i % hash.len();
        let char_index = (hash[byte_index] as usize) % valid_chars.len();
        result.push(valid_chars[char_index]);
    }

    result
}

/// Generates a new ASID with custom prefix and vanity components
/// Format: prefix(2) + random(2) + subject(5) + random(2) + predicate(7) + random(2) + context(5) + random(4) + actor(3)
pub fn generate_asid_with_prefix(
    prefix: &str,
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
) -> Result<String, IdError> {
    // Validate and normalize prefix
    let prefix = prefix.to_uppercase();
    if prefix.len() != 2 {
        return Err(IdError::InvalidLength(format!(
            "prefix must be exactly 2 characters, got {}",
            prefix.len()
        )));
    }

    let u = Uuid::new_v4();
    let uuid_str: String = u.to_string().replace('-', "").to_uppercase();

    // Extract vanity components with appropriate lengths
    let subject_vanity = extract_vanity_component(subject, 5);
    let predicate_vanity = extract_vanity_component(predicate, 7);
    let context_vanity = extract_vanity_component(context, 5);
    let actor_vanity = extract_vanity_component(actor, 3);

    // Extract random segments: 2 + 2 + 2 + 4 = 10 chars total
    let random1 = &uuid_str[0..2];
    let random2 = &uuid_str[2..4];
    let random3 = &uuid_str[4..6];
    let random4 = &uuid_str[6..10];

    // Construct ASID: 32 chars total
    Ok(format!(
        "{}{}{}{}{}{}{}{}{}",
        prefix,
        random1,
        subject_vanity,
        random2,
        predicate_vanity,
        random3,
        context_vanity,
        random4,
        actor_vanity
    ))
}

/// Generates a new ASID with vanity components (AS prefix)
pub fn generate_asid_with_vanity(
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
) -> Result<String, IdError> {
    generate_asid_with_prefix("AS", subject, predicate, context, actor)
}

/// Generates a new ASID with vanity components (AS prefix)
pub fn generate_asid(
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
) -> Result<String, IdError> {
    generate_asid_with_vanity(subject, predicate, context, actor)
}

/// Generates an ASID with custom prefix, vanity components, and collision detection
pub fn generate_asid_with_prefix_and_retry<F>(
    prefix: &str,
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
    check_exists: F,
) -> Result<String, IdError>
where
    F: Fn(&str) -> bool,
{
    const MAX_RETRIES: usize = 10;

    for _ in 0..MAX_RETRIES {
        let asid = generate_asid_with_prefix(prefix, subject, predicate, context, actor)?;

        if !check_exists(&asid) {
            return Ok(asid);
        }
    }

    // If we get here, we've had too many collisions
    // This is extremely unlikely with UUIDs, so just return the last attempt
    generate_asid_with_prefix(prefix, subject, predicate, context, actor)
}

/// Generates an ASID with vanity components and collision detection
pub fn generate_asid_with_vanity_and_retry<F>(
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
    check_exists: F,
) -> Result<String, IdError>
where
    F: Fn(&str) -> bool,
{
    generate_asid_with_prefix_and_retry("AS", subject, predicate, context, actor, check_exists)
}

/// Generates an ASID with collision detection
pub fn generate_asid_with_retry<F>(
    subject: &str,
    predicate: &str,
    context: &str,
    actor: &str,
    check_exists: F,
) -> Result<String, IdError>
where
    F: Fn(&str) -> bool,
{
    generate_asid_with_vanity_and_retry(subject, predicate, context, actor, check_exists)
}

/// Generates an async job ASID
/// Format: JB + random(2) + jobType(5) + random(2) + process(7) + random(2) + source(5) + random(4) + actor(3)
pub fn generate_job_asid(job_type: &str, source: &str, actor: &str) -> Result<String, IdError> {
    generate_asid_with_prefix("JB", job_type, "process", source, actor)
}

/// Generates a Job ASID with collision detection
pub fn generate_job_asid_with_retry<F>(
    job_type: &str,
    source: &str,
    actor: &str,
    check_exists: F,
) -> Result<String, IdError>
where
    F: Fn(&str) -> bool,
{
    generate_asid_with_prefix_and_retry("JB", job_type, "process", source, actor, check_exists)
}

/// Generates a Pulse Execution ID
pub fn generate_execution_id() -> String {
    generate_asid_simple("PX", "execution", "pulse")
}

/// Generates a simple ASID without retry logic (for non-critical IDs)
pub fn generate_asid_simple(prefix: &str, subject: &str, context: &str) -> String {
    generate_asid_with_prefix(prefix, subject, "id", context, "")
        .unwrap_or_else(|_| format!("{}00000000000000000000000000000", prefix))
}

/// Checks if a string is a valid ASID format
/// Supports any 2-letter prefix, legacy format, and new vanity format
pub fn is_valid_asid(id: &str) -> bool {
    // Must be 32 characters total
    if id.len() != 32 {
        return false;
    }

    let chars: Vec<char> = id.chars().collect();

    // Prefix must be 2 uppercase letters or digits
    for i in 0..2 {
        let c = chars[i];
        if !c.is_ascii_uppercase() && !c.is_ascii_digit() {
            return false;
        }
    }

    // Check if it's the legacy format (all remaining chars are hex)
    let is_legacy_format = id[2..].chars().all(|c| c.is_ascii_hexdigit());

    if is_legacy_format {
        return true;
    }

    // If not legacy format, check if it's valid new vanity format
    // Check random segments (positions: 2-3, 9-10, 18-19, 25-28) - must be hex
    let random_segments = [(2, 4), (9, 11), (18, 20), (25, 29)];
    for (start, end) in random_segments {
        for i in start..end {
            let c = chars[i];
            if !c.is_ascii_hexdigit() {
                return false;
            }
        }
    }

    // Check vanity segments (positions: 4-8, 11-17, 20-24, 29-31) - must be alphanumeric
    let vanity_segments = [(4, 9), (11, 18), (20, 25), (29, 32)];
    for (start, end) in vanity_segments {
        for i in start..end {
            let c = chars[i];
            if !c.is_ascii_alphanumeric() {
                return false;
            }
        }
    }

    true
}

/// Checks if an ASID uses the new vanity format
pub fn is_vanity_asid(id: &str) -> bool {
    if !is_valid_asid(id) {
        return false;
    }

    // Check if it's the legacy format (all remaining chars are hex)
    !id[2..].chars().all(|c| c.is_ascii_hexdigit())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_jd_asid() {
        let id = generate_jd_asid("Acme Corp", "Software Engineer", "San Francisco").unwrap();
        assert_eq!(id.len(), 32);
        assert!(id.starts_with("JD"));
    }

    #[test]
    fn test_generate_asid() {
        let id = generate_asid("Alice", "follows", "Twitter", "User").unwrap();
        assert_eq!(id.len(), 32);
        assert!(id.starts_with("AS"));
    }

    #[test]
    fn test_generate_asid_with_prefix() {
        let id = generate_asid_with_prefix("JB", "import", "process", "url", "sys").unwrap();
        assert_eq!(id.len(), 32);
        assert!(id.starts_with("JB"));
    }

    #[test]
    fn test_is_valid_asid() {
        let id = generate_asid("test", "test", "test", "tst").unwrap();
        assert!(is_valid_asid(&id));
        assert!(!is_valid_asid("short"));
        assert!(!is_valid_asid(""));
    }

    #[test]
    fn test_is_vanity_asid() {
        let id = generate_asid("Alice", "follows", "Bob", "User").unwrap();
        assert!(is_vanity_asid(&id));
    }

    #[test]
    fn test_generate_execution_id() {
        let id = generate_execution_id();
        assert_eq!(id.len(), 32);
        assert!(id.starts_with("PX"));
    }

    #[test]
    fn test_extract_vanity_component_padding() {
        // Short input should be padded deterministically
        let comp1 = extract_vanity_component("AB", 5);
        let comp2 = extract_vanity_component("AB", 5);
        assert_eq!(comp1.len(), 5);
        assert_eq!(comp1, comp2); // Should be deterministic
    }
}
