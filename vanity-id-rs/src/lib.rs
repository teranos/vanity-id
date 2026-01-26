//! # vanity-id
//!
//! Human-readable, memorable vanity identifiers from names and text.
//!
//! Unlike UUIDs or auto-incrementing integers, vanity IDs are short, professional-looking,
//! and easy to communicate verbally (e.g., "SBVH", "JDOE", "ACME").
//!
//! ## Key Concepts
//!
//! The package supports three types of ID generation:
//!
//! 1. **Vanity IDs**: Generated from entity attributes (names, titles, etc.)
//!    - Contact IDs: "SBVH", "JDOE" from human names
//!    - Organization IDs: "ACME", "NASA" from company names
//!    - Role IDs: "SWE", "PM" from job titles
//!
//! 2. **ASIDs (Application-Scoped IDs)**: Domain-specific structured identifiers
//!    - Job Description: "JD-ACME-SWE-NYC-A3B7"
//!    - Custom formats with prefix and suffix components
//!
//! 3. **Random IDs**: Cryptographically secure random identifiers when no seed is available
//!
//! ## Design Philosophy
//!
//! - **Character exclusions**: Excludes confusing characters (0/O, 1/I) for clear communication
//! - **Collision handling**: Automatically handles collisions with vowel variations or suffixes
//! - **Entity-specific constraints**: Different min/max lengths per entity type
//! - **Normalization**: Converts unicode to ASCII, removes invalid characters
//!
//! ## Thread Safety
//!
//! All exported functions are safe for concurrent use. Config changes via `set_config`
//! are protected by a mutex and affect subsequent ID generation calls.
//!
//! ## Basic Usage
//!
//! ```rust
//! use vanity_id::{generate_contact_id, generate_random_id, normalize_for_lookup};
//! use std::collections::HashSet;
//!
//! // Generate a random ID
//! let random_id = generate_random_id(5).unwrap();
//! println!("Random ID: {}", random_id);
//!
//! // Normalize user input for lookups (handles typos, case)
//! let normalized = normalize_for_lookup("jd0e");
//! assert_eq!(normalized, "JDOE");
//! ```
//!
//! ## Generating Contact IDs
//!
//! ```rust
//! use vanity_id::generate_contact_id;
//! use std::collections::HashSet;
//!
//! let mut assigned: HashSet<String> = HashSet::new();
//! let put = |id: &str| -> Result<(), vanity_id::IdError> {
//!     if assigned.contains(id) {
//!         Err(vanity_id::IdError::GenerationFailed("already exists".to_string()))
//!     } else {
//!         assigned.insert(id.to_string());
//!         Ok(())
//!     }
//! };
//!
//! let id = generate_contact_id("Jane", "Doe", None, put).unwrap();
//! // Result: "JDOE" or similar
//! ```

// Module declarations
mod alternative;
mod asid;
mod contact;
mod fuzzy;
mod id;
mod organization;
mod role;

// Re-export core types and functions from id module
pub use id::{
    // Configuration
    Config,
    EntityKind,
    EntityLengthConfig,
    set_config,
    get_config,
    get_max_retries,

    // Error type
    IdError,

    // Reserved words checking
    ReservedWordsChecker,
    HardcodedReservedChecker,

    // Core ID generation
    generate_vanity_id,
    generate_vanity_id_for_entity,
    generate_random_id,
    assign_id,

    // Normalization and cleaning
    normalize_to_ascii,
    clean_seed,
    convert_to_custom_alphabet,
    normalize_for_lookup,
    pad_to_min_length,
    pad_to_min_length_for_entity,
    pad_to_min_length_with_names,

    // Helpers
    is_vowel,
    is_vowel_char,
    CUSTOM_ALPHABET,
};

// Re-export contact functions
pub use contact::{
    build_contact_seed,
    extract_consonants,
    assign_contact_id,
    assign_contact_id_with_length_constraints,
    generate_contact_id,
    generate_contact_id_with_length_constraints,
};

// Re-export organization functions
pub use organization::{
    build_organization_seed,
    generate_organization_id,
};

// Re-export role functions
pub use role::{
    build_role_seed,
    generate_role_id,
};

// Re-export ASID functions
pub use asid::{
    // JD ASID
    generate_jd_asid,
    generate_jd_asid_with_retry,

    // Generic ASID
    generate_asid,
    generate_asid_with_vanity,
    generate_asid_with_prefix,
    generate_asid_with_retry,
    generate_asid_with_vanity_and_retry,
    generate_asid_with_prefix_and_retry,

    // Job ASID
    generate_job_asid,
    generate_job_asid_with_retry,

    // Execution ID
    generate_execution_id,
    generate_asid_simple,

    // Validation
    is_valid_asid,
    is_vanity_asid,
};

// Re-export alternative ID functions
pub use alternative::{
    AlternativeId,
    add_alternative_id,
    has_alternative_id,
    get_all_ids,
};

// Re-export fuzzy matching functions
pub use fuzzy::{
    is_similar,
    sort_by_similarity,
};

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;

    #[test]
    fn test_library_exports() {
        // Test that all main exports are accessible
        let _ = generate_random_id(5).unwrap();
        let _ = normalize_for_lookup("test");
        let _ = HardcodedReservedChecker::new();
    }

    #[test]
    fn test_contact_id_generation() {
        let mut assigned: HashSet<String> = HashSet::new();
        let put = |id: &str| -> Result<(), IdError> {
            if assigned.contains(id) {
                Err(IdError::GenerationFailed("already exists".to_string()))
            } else {
                assigned.insert(id.to_string());
                Ok(())
            }
        };

        let id = generate_contact_id("John", "Smith", None, put).unwrap();
        assert!(id.len() >= 4);
    }

    #[test]
    fn test_organization_id_generation() {
        let mut assigned: HashSet<String> = HashSet::new();
        let put = |id: &str| -> Result<(), IdError> {
            if assigned.contains(id) {
                Err(IdError::GenerationFailed("already exists".to_string()))
            } else {
                assigned.insert(id.to_string());
                Ok(())
            }
        };

        let id = generate_organization_id("Acme Corporation", None, put).unwrap();
        assert!(id.len() >= 3);
    }

    #[test]
    fn test_role_id_generation() {
        let mut assigned: HashSet<String> = HashSet::new();
        let put = |id: &str| -> Result<(), IdError> {
            if assigned.contains(id) {
                Err(IdError::GenerationFailed("already exists".to_string()))
            } else {
                assigned.insert(id.to_string());
                Ok(())
            }
        };

        let id = generate_role_id("Software Engineer", None, put).unwrap();
        assert!(id.len() >= 2);
    }

    #[test]
    fn test_asid_generation() {
        let id = generate_asid("Alice", "follows", "Bob", "User").unwrap();
        assert_eq!(id.len(), 32);
        assert!(is_valid_asid(&id));
    }

    #[test]
    fn test_fuzzy_matching() {
        assert!(is_similar("JDO", "JDOE"));
        assert!(!is_similar("JDOE", "JDOE")); // Exact match returns false
    }

    #[test]
    fn test_alternative_ids() {
        let ids: Vec<AlternativeId> = vec![];
        let new_ids = add_alternative_id(&ids, "PRIMARY", "ALT1", "merge", "system");
        assert!(has_alternative_id(&new_ids, "ALT1"));

        let all = get_all_ids("PRIMARY", &new_ids);
        assert_eq!(all, vec!["PRIMARY", "ALT1"]);
    }
}
