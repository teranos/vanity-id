//! Role-specific ID generation functionality.
//!
//! This module provides functions for generating vanity IDs for roles/job titles
//! using industry-standard abbreviations and intelligent phrase processing.

use once_cell::sync::Lazy;
use std::collections::{HashMap, HashSet};

use crate::id::{assign_id, EntityKind, IdError, ReservedWordsChecker};

/// Common words to exclude from role titles
static ROLE_EXCLUDE_WORDS: Lazy<HashSet<&'static str>> = Lazy::new(|| {
    [
        // Articles and prepositions
        "THE", "A", "AN", "OF", "AND",
        // Common connectors
        "FOR", "TO", "IN", "AT", "BY",
    ]
    .into_iter()
    .collect()
});

/// Role abbreviations and mappings (case-insensitive)
static ROLE_ABBREVIATIONS: Lazy<HashMap<&'static str, &'static str>> = Lazy::new(|| {
    [
        // Exact role matches (highest priority)
        ("SOFTWARE ENGINEER", "SWE"),
        ("PRODUCT MANAGER", "PM"),
        ("DATA SCIENTIST", "DS"),
        ("ENGINEERING MANAGER", "EM"),
        ("TECHNICAL PROGRAM MANAGER", "TPM"),
        ("PROGRAM MANAGER", "PGM"),
        ("BACKEND DEVELOPER", "BACKEND"),
        ("FRONTEND DEVELOPER", "FRONTEND"),
        ("FULL STACK DEVELOPER", "FULLSTACK"),
        ("DEVOPS ENGINEER", "DEVOPS"),
        ("SITE RELIABILITY ENGINEER", "SRE"),
        ("MACHINE LEARNING ENGINEER", "MLE"),
        ("QUALITY ASSURANCE", "QA"),
        ("BUSINESS ANALYST", "BA"),
        ("SYSTEMS ADMINISTRATOR", "SYSADMIN"),
        ("DATABASE ADMINISTRATOR", "DBA"),
        ("TECHNICAL WRITER", "TECHWRITER"),
        // Function/role components
        ("ENGINEER", "ENG"),
        ("MANAGER", "MGR"),
        ("DIRECTOR", "DIR"),
        ("DEVELOPER", "DEV"),
        ("DESIGNER", "DSGN"),
        ("ARCHITECT", "ARCH"),
        ("ANALYST", "ANLST"),
        ("SPECIALIST", "SPEC"),
        ("COORDINATOR", "COORD"),
        ("CONSULTANT", "CONSULT"),
        ("ADMINISTRATOR", "ADMIN"),
        ("TECHNICIAN", "TECH"),
        ("ASSOCIATE", "ASSOC"),
        ("ASSISTANT", "ASST"),
        // Domain/technology areas
        ("SOFTWARE", "SW"),
        ("HARDWARE", "HW"),
        ("PRODUCT", "PROD"),
        ("ENGINEERING", "ENG"),
        ("TECHNOLOGY", "TECH"),
        ("INFORMATION", "INFO"),
        ("OPERATIONS", "OPS"),
        ("MARKETING", "MKT"),
        ("SALES", "SALES"),
        ("CUSTOMER", "CUST"),
        ("BUSINESS", "BIZ"),
        ("DEVELOPMENT", "DEV"),
        ("RESEARCH", "RES"),
        ("DESIGN", "DSGN"),
        ("QUALITY", "QA"),
        ("SECURITY", "SEC"),
        ("NETWORK", "NET"),
        ("SYSTEMS", "SYS"),
        ("DATABASE", "DB"),
        ("ANALYTICS", "ANALYTICS"),
        ("DATA", "DATA"),
        ("SCIENCE", "SCI"),
        ("MACHINE", "ML"),
        ("LEARNING", "ML"),
        ("ARTIFICIAL", "AI"),
        ("INTELLIGENCE", "AI"),
        ("CLOUD", "CLOUD"),
        ("INFRASTRUCTURE", "INFRA"),
        ("PLATFORM", "PLATFORM"),
        ("MOBILE", "MOBILE"),
        ("WEB", "WEB"),
        ("BACKEND", "BACKEND"),
        ("FRONTEND", "FRONTEND"),
        ("FULLSTACK", "FULLSTACK"),
        ("FULL", "FULL"),
        ("STACK", "STACK"),
        // Experience levels
        ("SENIOR", "SR"),
        ("JUNIOR", "JR"),
        ("STAFF", "STF"),
        ("PRINCIPAL", "PRIN"),
        ("LEAD", "LEAD"),
        ("CHIEF", "CHIEF"),
        ("HEAD", "HEAD"),
        ("VICE", "VP"),
        ("PRESIDENT", "P"),
        ("EXECUTIVE", "EXEC"),
        ("OFFICER", "O"),
    ]
    .into_iter()
    .collect()
});

/// Common role patterns that should be kept together
static ROLE_PATTERNS: &[&str] = &[
    "SOFTWARE ENGINEER", "PRODUCT MANAGER", "DATA SCIENTIST",
    "ENGINEERING MANAGER", "BACKEND DEVELOPER", "FRONTEND DEVELOPER",
    "FULL STACK DEVELOPER", "DEVOPS ENGINEER", "SITE RELIABILITY ENGINEER",
    "MACHINE LEARNING ENGINEER", "QUALITY ASSURANCE", "BUSINESS ANALYST",
    "TECHNICAL PROGRAM MANAGER", "PROGRAM MANAGER",
    "SYSTEMS ADMINISTRATOR", "DATABASE ADMINISTRATOR",
];

/// Generates a vanity ID for a role using its title
pub fn generate_role_id<F>(
    title: &str,
    checker: Option<&dyn ReservedWordsChecker>,
    put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let seed = build_role_seed(title);
    assign_id(EntityKind::Role, &seed, checker, put)
}

/// Builds an optimized seed string from a role title
/// Filters common words and applies role-specific abbreviations
pub fn build_role_seed(title: &str) -> String {
    if title.is_empty() {
        return String::new();
    }

    // Step 1: Check if entire title matches a common role
    let upper_title = title.trim().to_uppercase();

    // Handle empty or whitespace-only input
    if upper_title.is_empty() {
        return String::new();
    }

    if let Some(abbr) = ROLE_ABBREVIATIONS.get(upper_title.as_str()) {
        return abbr.to_string();
    }

    // Step 2: Handle multi-word role phrases within the title
    let mut processed_title = upper_title.clone();
    let mut replacements: HashMap<String, String> = HashMap::new();
    let mut placeholder_counter = 0;

    // Replace multi-word patterns with single-word placeholders
    for pattern in ROLE_PATTERNS {
        if processed_title.contains(pattern) {
            if let Some(abbr) = ROLE_ABBREVIATIONS.get(*pattern) {
                let placeholder = format!("PLACEHOLDER{}", placeholder_counter);
                placeholder_counter += 1;
                replacements.insert(placeholder.clone(), abbr.to_string());
                processed_title = processed_title.replace(pattern, &placeholder);
            }
        }
    }

    // Step 3: Split into words and process remaining words
    let words: Vec<&str> = processed_title.split_whitespace().collect();
    let mut result: Vec<String> = Vec::new();

    for word in words {
        // Check if this is a placeholder
        if word.starts_with("PLACEHOLDER") {
            if let Some(abbr) = replacements.get(word) {
                result.push(abbr.clone());
                continue;
            }
        }

        // Skip excluded words
        if ROLE_EXCLUDE_WORDS.contains(word) {
            continue;
        }

        // Check if word has an abbreviation
        if let Some(abbr) = ROLE_ABBREVIATIONS.get(word) {
            result.push(abbr.to_string());
        } else {
            // Keep the word as-is
            result.push(word.to_string());
        }
    }

    // Join the processed words
    if result.is_empty() {
        // If everything was filtered, return original (will be cleaned by clean_seed)
        return title.to_string();
    }

    result.join(" ")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;

    #[test]
    fn test_build_role_seed() {
        assert_eq!(build_role_seed("Software Engineer"), "SWE");
        assert_eq!(build_role_seed("Product Manager"), "PM");
        assert_eq!(build_role_seed("Data Scientist"), "DS");
    }

    #[test]
    fn test_build_role_seed_with_level() {
        assert_eq!(build_role_seed("Senior Software Engineer"), "SR SWE");
    }

    #[test]
    fn test_generate_role_id() {
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
        assert!(id.len() <= 10);
    }

    #[test]
    fn test_empty_title() {
        assert_eq!(build_role_seed(""), "");
        assert_eq!(build_role_seed("   "), "");
    }
}
