//! Organization-specific ID generation functionality.
//!
//! This module provides functions for generating vanity IDs for organizations
//! using their names with intelligent phrase processing and acronym handling.

use once_cell::sync::Lazy;
use regex::Regex;
use std::collections::{HashMap, HashSet};

use crate::id::{assign_id, EntityKind, IdError, ReservedWordsChecker};

/// Prefix rule for organization name processing
struct PrefixRule {
    prefix: &'static str,
    min_len: usize,
    replacement: &'static str,
}

/// Organization prefix rules - checked in order (more specific first)
static ORG_PREFIX_RULES: &[PrefixRule] = &[
    PrefixRule { prefix: "ELECTRO", min_len: 7, replacement: "EL" },
    PrefixRule { prefix: "QUANTUM", min_len: 7, replacement: "Q" },
    PrefixRule { prefix: "CRYPTO", min_len: 6, replacement: "CRYP" },
    PrefixRule { prefix: "PHARMA", min_len: 5, replacement: "PH" },
    PrefixRule { prefix: "NEURO", min_len: 5, replacement: "NEUR" },
    PrefixRule { prefix: "MICRO", min_len: 5, replacement: "M" },
    PrefixRule { prefix: "CYBER", min_len: 5, replacement: "CY" },
    PrefixRule { prefix: "GENO", min_len: 4, replacement: "GEN" },
    PrefixRule { prefix: "GENE", min_len: 4, replacement: "GEN" },
    PrefixRule { prefix: "NANO", min_len: 4, replacement: "N" },
    PrefixRule { prefix: "TELE", min_len: 4, replacement: "TEL" },
    PrefixRule { prefix: "AUTO", min_len: 4, replacement: "AU" },
    PrefixRule { prefix: "AERO", min_len: 4, replacement: "AIR" },
    PrefixRule { prefix: "BIO", min_len: 3, replacement: "BIO" },
    PrefixRule { prefix: "ECO", min_len: 3, replacement: "ECO" },
];

/// Common words to exclude from organization names
static ORG_EXCLUDE_WORDS: Lazy<HashSet<&'static str>> = Lazy::new(|| {
    [
        // Articles
        "THE", "A", "AN",
        // Prepositions (various languages)
        "DE", "DU", "LA", "LE", "LES", "DES",
        "VAN", "VON", "DER", "DIE", "DAS", "OF",
        "EL", "LOS", "LAS", "DEL", "AL",
        "DA", "DO", "DOS",
        // Common business suffixes that might appear at start
        "AND", "&",
        // Generic qualifiers to ignore
        "NEW", "FIRST", "REAL", "GENERAL", "PROFESSIONAL",
    ]
    .into_iter()
    .collect()
});

/// Common long words to convert to acronyms
static ORG_ACRONYM_WORDS: Lazy<HashMap<&'static str, &'static str>> = Lazy::new(|| {
    [
        // Institutions
        ("UNIVERSITY", "U"), ("COLLEGE", "C"), ("INSTITUTE", "I"), ("ACADEMY", "A"),
        ("SCHOOL", "S"), ("FOUNDATION", "F"), ("ASSOCIATION", "A"),
        // Geographic/Political
        ("INTERNATIONAL", "I"), ("NATIONAL", "N"), ("FEDERAL", "F"), ("MINISTERIE", "MI"), ("MINISTRY", "MI"),
        ("EUROPEAN", "E"), ("AMERICAN", "A"), ("BRITISH", "B"), ("CANADIAN", "C"),
        ("AUSTRALIAN", "A"), ("GLOBAL", "G"), ("WORLDWIDE", "G"), ("DUTCH", "D"), ("UNITED", "UN"),
        ("CENTRAL", "C"), ("ADVANCED", "A"), ("INNOVATION", "IN"),
        // Cities and locations
        ("AMSTERDAM", "AMS"), ("LONDON", "LON"), ("CALIFORNIA", "CA"), ("CITY", "C"),
        // Geographic regions
        ("ASIAN", "AS"), ("AFRICAN", "AF"), ("LATIN", "LA"), ("PACIFIC", "PAC"), ("ATLANTIC", "ATL"),
        ("SCANDINAVIAN", "SC"), ("MEDITERRANEAN", "MED"), ("CARIBBEAN", "CAR"), ("MIDDLE", "MID"),
        ("EASTERN", "E"), ("WESTERN", "W"), ("NORTHERN", "N"), ("SOUTHERN", "S"), ("ARCTIC", "ARC"),
        // Business types
        ("CORPORATION", "C"), ("CORP", "C"), ("COMPANY", "C"), ("LIMITED", "L"), ("VENTURE", "V"),
        ("CAPITAL", "C"), ("INCORPORATED", "I"), ("INC", "I"), ("ENTERPRISES", "E"),
        ("INDUSTRIES", "I"), ("COLLECTIVE", "C"), ("LOBBY", "L"), ("GROUP", "G"),
        ("PARTNERS", "P"), ("AGENCY", "A"),
        ("TECHNOLOGIES", "T"), ("TECHNOLOGY", "T"), ("SYSTEMS", "S"),
        ("SOLUTIONS", "S"), ("SERVICES", "S"), ("SERVICE", "S"), ("CONSULTING", "C"),
        ("MANAGEMENT", "M"), ("DEVELOPMENT", "D"), ("RESEARCH", "R"), ("MEDIA", "M"),
        ("NETWORK", "N"), ("DIGITAL", "D"), ("BUSINESS", "B"), ("WORLD", "G"), ("SECURITY", "S"),
        ("OFFICE", "O"), ("STUDIO", "S"), ("DESIGN", "D"), ("ONLINE", "O"), ("MOBILE", "M"),
        ("CLOUD", "C"), ("FINANCE", "F"), ("MARKET", "M"), ("SCIENCE", "S"), ("HEALTH", "H"),
        ("LEGAL", "L"), ("SOCIAL", "S"), ("SUPPORT", "S"), ("TRAINING", "T"), ("LEARNING", "L"),
        ("MARKETING", "M"), ("PUBLISHING", "P"), ("HOLDINGS", "H"), ("OPERATIONS", "O"),
        // Industry verticals
        ("AUTOMOTIVE", "AU"), ("RETAIL", "RT"), ("HOSPITALITY", "HO"), ("AGRICULTURE", "AG"),
        ("TEXTILE", "TX"), ("FURNITURE", "FU"), ("JEWELRY", "JW"), ("FASHION", "FS"),
        ("LOGISTICS", "LG"), ("SHIPPING", "SH"), ("AVIATION", "AV"), ("MARITIME", "MAR"),
        ("ENERGY", "EN"), ("RENEWABLE", "RN"), ("PETROLEUM", "PT"), ("MINING", "MN"),
        ("ENTERTAINMENT", "ET"), ("GAMING", "GM"), ("SPORTS", "SP"), ("FITNESS", "FT"),
        ("TOURISM", "TM"), ("RESTAURANT", "RS"), ("CATERING", "CT"),
        // Specialized terms
        ("PROJECT", "P"), ("INTELLIGENCE", "I"),
        ("DATA", "D"), ("BLOCKCHAIN", "B"), ("DEFENSE", "DEF"), ("DEFENSIE", "DEF"),
        ("CROSS", "X"), ("SPACE", "S"),
        // Financial/Business terms
        ("INVESTMENT", "INV"), ("BANKING", "BK"), ("INSURANCE", "INS"), ("WEALTH", "WL"),
        ("PENSION", "PN"), ("MORTGAGE", "MG"), ("CREDIT", "CR"), ("EQUITY", "EQ"),
        ("ASSET", "AS"), ("PORTFOLIO", "PF"), ("FUND", "FD"), ("TREASURY", "TR"),
        ("ACCOUNTING", "AC"), ("AUDIT", "AD"), ("COMPLIANCE", "CP"), ("RISK", "RK"),
        // Academic/Research terms
        ("LABORATORY", "LAB"), ("OBSERVATORY", "OBS"), ("LIBRARY", "LIB"), ("MUSEUM", "MUS"),
        ("ARCHIVE", "ARC"), ("STUDY", "ST"), ("EXPERIMENT", "EXP"), ("ANALYSIS", "AN"),
        ("THEORY", "TH"), ("METHODOLOGY", "MET"),
        // Modern business models
        ("PLATFORM", "PL"), ("MARKETPLACE", "MK"), ("SUBSCRIPTION", "SUB"), ("STREAMING", "ST"),
        ("SHARING", "SH"), ("CROWDSOURCING", "CS"), ("FREELANCE", "FL"), ("REMOTE", "RM"),
        ("VIRTUAL", "VR"), ("HYBRID", "HY"), ("ECOSYSTEM", "ECO"),
        ("ACCELERATOR", "ACC"), ("INCUBATOR", "INC"), ("COWORKING", "CW"), ("STARTUP", "SU"),
        // Other long common words
        ("ENVIRONMENTAL", "EN"), ("PHARMACEUTICAL", "PH"), ("PHARMACEUTICALS", "PH"), ("MANUFACTURING", "MF"),
        ("COMMUNICATIONS", "CM"), ("TRANSPORTATION", "TR"), ("CONSTRUCTION", "CN"),
        ("ENGINEERING", "EG"), ("FINANCIAL", "FN"), ("HEALTHCARE", "HC"),
    ]
    .into_iter()
    .collect()
});

/// Multi-word phrases to replace
static ORG_PHRASE_REPLACEMENTS: Lazy<HashMap<&'static str, &'static str>> = Lazy::new(|| {
    [
        ("NEW YORK", "NY"),
        ("CYBER SECURITY", "CS"),
        ("UNITED STATES", "US"),
        ("UNITED KINGDOM", "UK"),
        ("MACHINE LEARNING", "ML"),
        ("ARTIFICIAL INTELLIGENCE", "AI"),
        ("VENTURE CAPITAL", "VC"),
        ("PRIVATE EQUITY", "PE"),
        ("REAL ESTATE", "RE"),
        ("HUMAN RESOURCES", "HR"),
        ("CUSTOMER SERVICE", "CS"),
        ("SUPPLY CHAIN", "SC"),
        ("CLOUD COMPUTING", "CC"),
        ("DATA SCIENCE", "DS"),
    ]
    .into_iter()
    .collect()
});

/// Generates a vanity ID for an organization using its name
pub fn generate_organization_id<F>(
    name: &str,
    checker: Option<&dyn ReservedWordsChecker>,
    put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let seed = build_organization_seed(name);
    assign_id(EntityKind::Organization, &seed, checker, put)
}

/// Builds a seed string from organization name
pub fn build_organization_seed(name: &str) -> String {
    if name.is_empty() {
        return String::new();
    }

    // Handle multi-word phrases first (case-insensitive)
    let mut processed_name = name.trim().to_string();

    // Apply phrase replacements
    for (phrase, replacement) in ORG_PHRASE_REPLACEMENTS.iter() {
        let re = Regex::new(&format!("(?i){}", regex::escape(phrase))).unwrap();
        processed_name = re.replace_all(&processed_name, *replacement).to_string();
    }

    // Split into words and process
    let words: Vec<&str> = processed_name.split_whitespace().collect();
    let mut processed_words: Vec<String> = Vec::new();

    for word in words {
        let upper_word = word.to_uppercase();

        // Skip excluded words
        if ORG_EXCLUDE_WORDS.contains(upper_word.as_str()) {
            continue;
        }

        // Check prefix patterns first (priority over exact matches)
        let mut matched = false;
        for rule in ORG_PREFIX_RULES {
            if upper_word.len() >= rule.min_len && upper_word.starts_with(rule.prefix) {
                processed_words.push(rule.replacement.to_string());
                matched = true;
                break;
            }
        }

        if matched {
            continue;
        }

        // Check if word should be converted to acronym
        if let Some(acronym) = ORG_ACRONYM_WORDS.get(upper_word.as_str()) {
            processed_words.push(acronym.to_string());
        } else {
            processed_words.push(word.to_string());
        }
    }

    // If all words were filtered out, use original name
    if processed_words.is_empty() {
        return name.trim().to_string();
    }

    processed_words.join(" ")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;

    #[test]
    fn test_build_organization_seed() {
        // Test basic organization names
        assert_eq!(build_organization_seed("Cyberdyne Systems"), "CY S");
        assert_eq!(build_organization_seed("Biosyn Corporation"), "BIO C");
    }

    #[test]
    fn test_phrase_replacement() {
        assert!(build_organization_seed("New York Tech").contains("NY"));
        assert!(build_organization_seed("Machine Learning Corp").contains("ML"));
    }

    #[test]
    fn test_generate_organization_id() {
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
        assert!(id.len() <= 9);
    }
}
