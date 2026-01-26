//! Alternative ID management functionality.
//!
//! This module provides functions for tracking alternative identifiers for entities,
//! such as CardDAV UIDs, merged IDs, or imported IDs.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// AlternativeID represents an alternative ID with timestamp and attestation tracking
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AlternativeId {
    /// The alternative ID (CardDAV UID, vanity ID, etc.)
    pub id: String,
    /// When this ID was added to the entity
    pub added_at: DateTime<Utc>,
    /// Source of the ID (merge, import, carddav-sync, etc.)
    pub source: String,
    /// Who/what is claiming this ID belongs to the entity
    pub attestor: String,
}

impl AlternativeId {
    /// Creates a new AlternativeID with the current timestamp
    pub fn new(id: String, source: String, attestor: String) -> Self {
        Self {
            id,
            added_at: Utc::now(),
            source,
            attestor,
        }
    }
}

/// Adds an ID to the alternative IDs list if it doesn't already exist
pub fn add_alternative_id(
    existing_ids: &[AlternativeId],
    primary_id: &str,
    id: &str,
    source: &str,
    attestor: &str,
) -> Vec<AlternativeId> {
    // Don't add the primary ID as an alternative
    if id == primary_id {
        return existing_ids.to_vec();
    }

    // Check if ID already exists in alternatives
    for existing_alt_id in existing_ids {
        if existing_alt_id.id == id {
            return existing_ids.to_vec(); // Already exists
        }
    }

    // Add the ID with timestamp and attestation
    let mut result = existing_ids.to_vec();
    result.push(AlternativeId::new(
        id.to_string(),
        source.to_string(),
        attestor.to_string(),
    ));
    result
}

/// Checks if the given ID exists in the alternative IDs list
pub fn has_alternative_id(existing_ids: &[AlternativeId], id: &str) -> bool {
    existing_ids.iter().any(|alt_id| alt_id.id == id)
}

/// Returns all IDs for an entity (primary ID + alternative IDs)
pub fn get_all_ids(primary_id: &str, alternative_ids: &[AlternativeId]) -> Vec<String> {
    let mut all_ids = vec![primary_id.to_string()];
    for alt_id in alternative_ids {
        all_ids.push(alt_id.id.clone());
    }
    all_ids
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add_alternative_id() {
        let existing: Vec<AlternativeId> = vec![];
        let result = add_alternative_id(&existing, "PRIMARY", "ALT1", "merge", "system");

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].id, "ALT1");
        assert_eq!(result[0].source, "merge");
        assert_eq!(result[0].attestor, "system");
    }

    #[test]
    fn test_add_alternative_id_no_duplicate() {
        let existing = vec![AlternativeId::new(
            "ALT1".to_string(),
            "merge".to_string(),
            "system".to_string(),
        )];
        let result = add_alternative_id(&existing, "PRIMARY", "ALT1", "import", "user");

        assert_eq!(result.len(), 1); // Should not add duplicate
    }

    #[test]
    fn test_add_alternative_id_no_primary() {
        let existing: Vec<AlternativeId> = vec![];
        let result = add_alternative_id(&existing, "PRIMARY", "PRIMARY", "merge", "system");

        assert_eq!(result.len(), 0); // Should not add primary as alternative
    }

    #[test]
    fn test_has_alternative_id() {
        let existing = vec![AlternativeId::new(
            "ALT1".to_string(),
            "merge".to_string(),
            "system".to_string(),
        )];

        assert!(has_alternative_id(&existing, "ALT1"));
        assert!(!has_alternative_id(&existing, "ALT2"));
    }

    #[test]
    fn test_get_all_ids() {
        let alt_ids = vec![
            AlternativeId::new("ALT1".to_string(), "merge".to_string(), "system".to_string()),
            AlternativeId::new("ALT2".to_string(), "import".to_string(), "user".to_string()),
        ];

        let all_ids = get_all_ids("PRIMARY", &alt_ids);
        assert_eq!(all_ids, vec!["PRIMARY", "ALT1", "ALT2"]);
    }
}
