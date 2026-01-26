//! Contact-specific ID generation functionality.
//!
//! This module provides functions for generating vanity IDs for human contacts
//! using their first and last names with intelligent name processing.

use std::collections::HashSet;

use crate::id::{
    clean_seed, convert_to_custom_alphabet, generate_random_id, get_max_retries,
    is_vowel_char, pad_to_min_length_with_names, EntityKind, IdError,
    ReservedWordsChecker,
};

/// Name processing thresholds
const VERY_SHORT_NAME_LENGTH: usize = 3;
const SHORT_NAME_LENGTH: usize = 4;
const MEDIUM_NAME_LENGTH: usize = 6;
const MIN_SEGMENT_LENGTH: usize = 2;
const FALLBACK_LENGTH: usize = 3;

/// Common name particles to filter (case-insensitive)
static NAME_PARTICLES: &[&str] = &[
    // Dutch/Germanic
    "VAN", "VON", "VAN DER", "VAN DEN", "VON DER",
    // Romance languages
    "DA", "DE", "DI", "DEL", "DELLA", "DELLE",
    "DU", "DES", "LA", "LE", "LOS", "LAS",
    // Arabic/Spanish articles
    "AL", "EL",
    // Religious/Geographic
    "SAINT", "ST", "SAN", "SANTA",
];

/// Filters common name particles/articles from a name
fn filter_name_particles(name: &str) -> String {
    if name.is_empty() {
        return String::new();
    }

    let particles: HashSet<&str> = NAME_PARTICLES.iter().copied().collect();
    let mut result = name.to_string();

    // Handle multi-word particles first (case-insensitive removal)
    for particle in NAME_PARTICLES {
        if particle.contains(' ') {
            // Case-insensitive replacement
            let upper = result.to_uppercase();
            if let Some(pos) = upper.find(particle) {
                result = format!("{}{}", &result[..pos], &result[pos + particle.len()..]);
            }
        }
    }

    // Split into words and filter single-word particles
    let words: Vec<&str> = result.split_whitespace().collect();
    let filtered_words: Vec<&str> = words
        .into_iter()
        .filter(|word| {
            let upper_word = word.to_uppercase();
            !particles.contains(upper_word.as_str()) && !word.trim().is_empty()
        })
        .collect();

    filtered_words.join(" ")
}

/// Builds a seed string from contact first and last name
/// Uses bigram/trigram approach for more intuitive name-like IDs
pub fn build_contact_seed(first_name: &str, last_name: &str) -> String {
    if last_name.is_empty() && first_name.is_empty() {
        return String::new();
    }

    // Filter particles from names before processing
    let filtered_first_name = filter_name_particles(first_name);
    let filtered_last_name = filter_name_particles(last_name);

    let mut seed = String::new();

    // Extract meaningful segments from first name
    if !filtered_first_name.is_empty() {
        let has_last_name = !filtered_last_name.is_empty();
        let first_segments = extract_name_segments(&filtered_first_name, true, has_last_name);
        seed.push_str(&first_segments);
    }

    // Extract meaningful segments from last name
    if !filtered_last_name.is_empty() {
        let last_segments = extract_name_segments(&filtered_last_name, false, false);
        seed.push_str(&last_segments);
    }

    // Special case: if we only have a first name and the seed is too short, use the full name
    let contact_min_length = EntityKind::HumanContact.get_min_length();
    if filtered_last_name.is_empty()
        && seed.len() < contact_min_length
        && filtered_first_name.len() >= contact_min_length
    {
        return filtered_first_name.to_uppercase();
    }

    seed
}

/// Extracts meaningful bigrams/trigrams from a name for intuitive IDs
fn extract_name_segments(name: &str, prioritize_full_name: bool, has_last_name: bool) -> String {
    if name.is_empty() {
        return String::new();
    }

    let name_upper = name.to_uppercase();
    let contact_min_length = EntityKind::HumanContact.get_min_length();

    // SPECIAL STRATEGY: For long first names with a last name, use first+last letter
    if prioritize_full_name && has_last_name && name_upper.len() > contact_min_length {
        let first_letter = name_upper.chars().next().unwrap();
        let last_letter = name_upper.chars().last().unwrap();
        return format!("{}{}", first_letter, last_letter);
    }

    // For short-to-medium names, try to keep the full name if possible
    if prioritize_full_name && name_upper.len() <= contact_min_length {
        return name_upper;
    }

    // For longer names, extract meaningful consonant-vowel patterns
    if name_upper.len() <= MEDIUM_NAME_LENGTH {
        if name_upper.len() <= VERY_SHORT_NAME_LENGTH {
            return name_upper;
        }
        return extract_optimal_segment(&name_upper);
    }

    // For long names, extract the most meaningful bigrams/trigrams
    extract_bigrams_trigrams_strategy(&name_upper)
}

/// Extracts the most meaningful segment from a medium-length name
fn extract_optimal_segment(name: &str) -> String {
    if name.len() <= VERY_SHORT_NAME_LENGTH {
        return name.to_string();
    }

    // Extract meaningful consonant patterns
    let segments = extract_meaningful_bigrams(name);
    if !segments.is_empty() {
        // Prefer segments that start at the beginning of the name
        for segment in &segments {
            if name.starts_with(segment) && segment.len() >= MIN_SEGMENT_LENGTH {
                return segment.clone();
            }
        }
        // Fallback to first segment
        return segments[0].clone();
    }

    // Fallback: first few characters
    if name.len() >= FALLBACK_LENGTH {
        return name[..FALLBACK_LENGTH].to_string();
    }
    name.to_string()
}

/// Finds consonant-vowel or consonant-consonant bigrams that preserve name structure
fn extract_meaningful_bigrams(name: &str) -> Vec<String> {
    let mut segments = Vec::new();
    let chars: Vec<char> = name.chars().collect();

    // Extract bigrams that maintain phonetic structure
    for i in 0..chars.len().saturating_sub(1) {
        let bigram: String = chars[i..i + 2].iter().collect();

        // Prefer consonant-vowel or vowel-consonant patterns
        let is_cv = !is_vowel_char(chars[i]) && is_vowel_char(chars[i + 1]); // Consonant-Vowel
        let is_vc = is_vowel_char(chars[i]) && !is_vowel_char(chars[i + 1]); // Vowel-Consonant
        let is_cc = !is_vowel_char(chars[i]) && !is_vowel_char(chars[i + 1]); // Consonant-Consonant

        if is_cv || is_vc || (is_cc && i == 0) {
            segments.push(bigram);
        }
    }

    // Also try trigrams for better representation
    for i in 0..chars.len().saturating_sub(2) {
        let trigram: String = chars[i..i + 3].iter().collect();
        segments.push(trigram);
    }

    segments
}

/// Handles long names with systematic bigram/trigram extraction
fn extract_bigrams_trigrams_strategy(name: &str) -> String {
    let chars: Vec<char> = name.chars().collect();
    if chars.len() <= SHORT_NAME_LENGTH {
        return name.to_string();
    }

    // Strategy 1: Start + meaningful middle/end segments
    let start: String = chars[..2].iter().collect();

    // Find a good ending segment
    let end = if chars.len() >= SHORT_NAME_LENGTH {
        // Look for good consonant combinations in the latter half
        let remaining: String = chars[2..].iter().collect();
        let segments = extract_meaningful_bigrams(&remaining);
        if !segments.is_empty() {
            let mut e = segments[0].clone();
            if e.len() > 2 {
                e = e[..2].to_string();
            }
            e
        } else {
            // Fallback: last 2 chars
            chars[chars.len() - 2..].iter().collect()
        }
    } else {
        String::new()
    };

    let result = format!("{}{}", start, end);

    // Ensure we don't exceed reasonable length for seed
    let contact_min_length = EntityKind::HumanContact.get_min_length();
    let max_truncate_length = contact_min_length - 1;
    if result.len() > max_truncate_length {
        return result[..max_truncate_length].to_string();
    }

    result
}

/// Extracts consonants from a name for more name-like IDs
pub fn extract_consonants(name: &str) -> String {
    const CONSONANTS: &str = "bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ";

    name.chars()
        .filter(|c| CONSONANTS.contains(*c))
        .collect()
}

/// Assigns a vanity ID for a contact with name-aware collision handling
pub fn assign_contact_id<F>(
    kind: EntityKind,
    seed: &str,
    first_name: &str,
    last_name: &str,
    checker: Option<&dyn ReservedWordsChecker>,
    put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let min_length = kind.get_min_length();
    let max_length = kind.get_max_length();
    assign_contact_id_with_length_constraints(
        kind,
        seed,
        first_name,
        last_name,
        min_length,
        max_length,
        checker,
        put,
    )
}

/// Assigns a vanity ID for a contact with custom length constraints
pub fn assign_contact_id_with_length_constraints<F>(
    _kind: EntityKind,
    seed: &str,
    first_name: &str,
    last_name: &str,
    min_length: usize,
    max_length: usize,
    checker: Option<&dyn ReservedWordsChecker>,
    mut put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    // Clean and validate the seed
    let cleaned_seed = clean_seed(seed);
    if cleaned_seed.is_empty() {
        // Fallback to random ID if no valid seed
        for _ in 0..get_max_retries() {
            let random_id = generate_random_id(min_length)?;
            if put(&random_id).is_ok() {
                return Ok(random_id);
            }
        }
        return Err(IdError::MaxRetriesExceeded(get_max_retries()));
    }

    // Track attempted IDs to prevent duplicates within this collision session
    let mut attempted: HashSet<String> = HashSet::new();

    // VOWEL-FIRST STRATEGY: Try vowel-rich reconstructions first, then fall back to consonant-only
    for length in min_length..=max_length {
        // Pad to target length using intelligent padding from original names
        let mut candidate_base =
            pad_to_min_length_with_names(&cleaned_seed, length, first_name, last_name);

        // Truncate if too long for this attempt
        if candidate_base.len() > length {
            candidate_base = candidate_base[..length].to_string();
        }

        // PRIORITY 1: Try vowel-rich name reconstructions FIRST
        let vowel_variations =
            generate_name_aware_vowel_variations(&candidate_base, first_name, last_name, max_length);
        for candidate_id in &vowel_variations {
            // Skip if doesn't match current length attempt
            if candidate_id.len() != length {
                continue;
            }

            // Skip if we've already attempted this ID
            if attempted.contains(candidate_id) {
                continue;
            }
            attempted.insert(candidate_id.clone());

            // Check reserved words
            if let Some(checker) = checker {
                if checker.is_reserved(candidate_id)? {
                    continue;
                }
            }

            if put(candidate_id).is_ok() {
                return Ok(candidate_id.clone());
            }
        }

        // PRIORITY 2: Fall back to consonant-only base if vowel-rich forms are taken
        if !attempted.contains(&candidate_base) {
            attempted.insert(candidate_base.clone());

            if let Some(checker) = checker {
                if checker.is_reserved(&candidate_base)? {
                    continue;
                }
            }

            if put(&candidate_base).is_ok() {
                return Ok(candidate_base);
            }
        }
    }

    Err(IdError::GenerationFailed(format!(
        "failed to assign vanity ID after trying all lengths up to {}",
        max_length
    )))
}

/// Creates variations using reverse vowel strategy
fn generate_name_aware_vowel_variations(
    base_id: &str,
    first_name: &str,
    last_name: &str,
    max_length: usize,
) -> Vec<String> {
    if base_id.is_empty() {
        return vec![];
    }

    // Extract vowels from the original names
    let original_vowels = extract_vowels_from_names(first_name, last_name);

    // Generate the full vowel-inserted reconstruction first (highest priority)
    generate_vowel_first_reconstruction(base_id, &original_vowels, first_name, last_name, max_length)
}

/// Extracts vowels from first and last name in order of appearance
fn extract_vowels_from_names(first_name: &str, last_name: &str) -> Vec<String> {
    let mut vowels = Vec::new();
    let mut seen: HashSet<String> = HashSet::new();

    // Process full name to maintain vowel order
    let full_name = format!("{}{}", first_name, last_name);
    for c in full_name.chars() {
        if is_vowel_char(c) && c.to_ascii_uppercase() != 'Y' {
            let vowel = c.to_ascii_uppercase().to_string();
            if !seen.contains(&vowel) {
                vowels.push(vowel.clone());
                seen.insert(vowel);
            }
        }
    }

    // If no vowels found, use common English vowels in frequency order
    if vowels.is_empty() {
        vowels = vec![
            "E".to_string(),
            "A".to_string(),
            "I".to_string(),
            "O".to_string(),
            "U".to_string(),
        ];
    }

    vowels
}

/// Creates variations starting with full vowel reconstruction
fn generate_vowel_first_reconstruction(
    base_id: &str,
    original_vowels: &[String],
    first_name: &str,
    last_name: &str,
    max_length: usize,
) -> Vec<String> {
    let mut variations = Vec::new();

    if original_vowels.is_empty() {
        return vec![base_id.to_string()];
    }

    let min_len = EntityKind::HumanContact.get_min_length();

    // PRIORITY 0: Use the baseID first if it's meaningful
    if base_id.len() >= min_len && base_id.len() <= max_length {
        variations.push(base_id.to_string());
    }

    // PRIORITY 1: Create the most vowel-rich reconstruction from combined seed
    let full_name_reconstruction =
        create_full_name_reconstruction(base_id, original_vowels, first_name, last_name, max_length);
    if !full_name_reconstruction.is_empty() && full_name_reconstruction != base_id {
        append_unique(&mut variations, full_name_reconstruction);
    }

    // PRIORITY 2: For very short names, add bigram variations
    if first_name.len() == VERY_SHORT_NAME_LENGTH {
        let first_name_upper = first_name.to_uppercase();
        let valid_name = convert_to_custom_alphabet(&first_name_upper);

        if valid_name.len() >= MIN_SEGMENT_LENGTH {
            let bigrams = vec![
                valid_name[..2].to_string(),
                valid_name[1..].to_string(),
            ];

            for bigram in bigrams {
                let mut extended = bigram.clone();
                let mut suffix = String::new();
                if !last_name.is_empty() && base_id.len() > 2 {
                    let remaining = &base_id[2..];
                    let needed_chars = min_len.saturating_sub(extended.len());
                    if remaining.len() >= needed_chars {
                        suffix = remaining[..needed_chars].to_string();
                    } else {
                        suffix = remaining.to_string();
                    }
                }
                extended.push_str(&suffix);

                // Pad if still needed
                let mut vowel_idx = 0;
                while extended.len() < min_len && vowel_idx < original_vowels.len() {
                    extended.push_str(&original_vowels[vowel_idx]);
                    vowel_idx += 1;
                }
                while extended.len() < min_len {
                    extended.push('G');
                }

                if extended.len() >= min_len && extended.len() <= max_length {
                    append_unique(&mut variations, extended);
                }
            }
        }
    }

    // PRIORITY 3: Create progressively vowel-reduced variations
    let progressive_variations =
        create_progressive_vowel_reduction(base_id, original_vowels, first_name, last_name, max_length);
    for variant in progressive_variations {
        append_unique(&mut variations, variant);
    }

    // PRIORITY 4: Ensure the consonant-only baseID is included as final fallback
    append_unique(&mut variations, base_id.to_string());

    variations
}

/// Attempts to create the most readable name form
fn create_full_name_reconstruction(
    base_id: &str,
    original_vowels: &[String],
    first_name: &str,
    _last_name: &str,
    max_length: usize,
) -> String {
    let min_len = EntityKind::HumanContact.get_min_length();

    // PRIORITY 1: For short names, try meaningful extensions
    if first_name.len() >= VERY_SHORT_NAME_LENGTH && first_name.len() <= min_len {
        let first_name_upper = first_name.to_uppercase();
        let valid_name = convert_to_custom_alphabet(&first_name_upper);

        if valid_name.len() >= min_len && valid_name.len() <= max_length {
            return valid_name;
        } else if valid_name.len() >= VERY_SHORT_NAME_LENGTH && valid_name.len() < min_len {
            let mut extended = valid_name;

            // Strategy A: Add vowels from the original names
            if !original_vowels.is_empty() {
                let mut vowel_index = 0;
                while extended.len() < min_len && vowel_index < original_vowels.len() {
                    extended.push_str(&original_vowels[vowel_index]);
                    vowel_index += 1;
                }
            }

            // Strategy B: Minimal fallback if still too short
            while extended.len() < min_len {
                extended.push('E');
            }

            if extended.len() >= min_len && extended.len() <= max_length {
                return extended;
            }
        }
    }

    // Approach 2: Create systematic vowel insertion based on consonant structure
    if base_id.len() <= 3 && original_vowels.len() >= 2 {
        let consonants: Vec<char> = base_id.chars().collect();
        let mut result = consonants[0].to_string();

        // Insert vowels between consonants
        for i in 1..consonants.len().min(original_vowels.len() + 1) {
            result.push_str(&original_vowels[i - 1]);
            result.push(consonants[i]);
        }

        // Add final vowel if we have more vowels and space
        if result.len() < max_length && original_vowels.len() > consonants.len() - 1 {
            result.push_str(&original_vowels[consonants.len() - 1]);
        }

        if result.len() <= max_length {
            return result;
        }
    }

    // Approach 3: For longer consonant strings, strategic vowel insertion
    if base_id.len() >= SHORT_NAME_LENGTH {
        let chars: Vec<char> = base_id.chars().collect();

        if !original_vowels.is_empty() {
            let rest: String = chars[1..].iter().collect();
            let result = format!("{}{}{}", chars[0], original_vowels[0], rest);
            if result.len() <= max_length {
                // Now try to insert secondary vowel
                if original_vowels.len() > 1 && result.len() + 1 <= max_length {
                    let mid_point = result.len() / 2;
                    if mid_point > 1 && mid_point < result.len() {
                        let final_result = format!(
                            "{}{}{}",
                            &result[..mid_point],
                            original_vowels[1],
                            &result[mid_point..]
                        );
                        if final_result.len() <= max_length {
                            return final_result;
                        }
                    }
                }
                return result;
            }
        }
    }

    String::new()
}

/// Creates variations by progressively removing vowels
fn create_progressive_vowel_reduction(
    base_id: &str,
    original_vowels: &[String],
    _first_name: &str,
    _last_name: &str,
    max_length: usize,
) -> Vec<String> {
    let mut variations = Vec::new();

    if !original_vowels.is_empty() {
        let primary_vowel = &original_vowels[0];

        // Variation 1: Insert primary vowel only
        if !base_id.is_empty() {
            let variation1 = format!("{}{}{}", &base_id[..1], primary_vowel, &base_id[1..]);
            if variation1.len() <= max_length && variation1 != base_id {
                variations.push(variation1);
            }
        }

        // Variation 2: Replace last consonant with primary vowel
        if base_id.len() >= 2 {
            let mut chars: Vec<char> = base_id.chars().collect();
            let last_idx = chars.len() - 1;
            chars[last_idx] = primary_vowel.chars().next().unwrap();
            let variation2: String = chars.into_iter().collect();
            if variation2.len() <= max_length && variation2 != base_id {
                variations.push(variation2);
            }
        }

        // Variation 3: If we have multiple vowels, try inserting the second vowel
        if original_vowels.len() > 1 && base_id.len() >= 2 {
            let second_vowel = &original_vowels[1];
            let variation3 = format!("{}{}{}", &base_id[..2], second_vowel, &base_id[2..]);
            if variation3.len() <= max_length && variation3 != base_id {
                variations.push(variation3);
            }
        }
    }

    variations
}

/// Appends a string to a vector only if it doesn't already exist
fn append_unique(vec: &mut Vec<String>, item: String) {
    if !vec.contains(&item) {
        vec.push(item);
    }
}

/// Generates a vanity ID for a contact using their name
pub fn generate_contact_id<F>(
    first_name: &str,
    last_name: &str,
    checker: Option<&dyn ReservedWordsChecker>,
    put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let min_length = EntityKind::HumanContact.get_min_length();
    let max_length = EntityKind::HumanContact.get_max_length();
    generate_contact_id_with_length_constraints(
        first_name,
        last_name,
        min_length,
        max_length,
        checker,
        put,
    )
}

/// Generates a vanity ID for a contact with custom length constraints
pub fn generate_contact_id_with_length_constraints<F>(
    first_name: &str,
    last_name: &str,
    min_length: usize,
    max_length: usize,
    checker: Option<&dyn ReservedWordsChecker>,
    put: F,
) -> Result<String, IdError>
where
    F: FnMut(&str) -> Result<(), IdError>,
{
    let seed = build_contact_seed(first_name, last_name);
    assign_contact_id_with_length_constraints(
        EntityKind::HumanContact,
        &seed,
        first_name,
        last_name,
        min_length,
        max_length,
        checker,
        put,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_build_contact_seed() {
        assert_eq!(build_contact_seed("Neo", ""), "NEO");
        // Sarah Connor: S + H (first+last of first name) + CO (first 2 of last name) = SHCO
        assert_eq!(build_contact_seed("Sarah", "Connor"), "SHCO");
    }

    #[test]
    fn test_filter_name_particles() {
        assert_eq!(filter_name_particles("Van Gogh"), "Gogh");
        assert_eq!(filter_name_particles("Jean De Valois"), "Jean Valois");
    }

    #[test]
    fn test_extract_consonants() {
        assert_eq!(extract_consonants("Hello"), "Hll");
        assert_eq!(extract_consonants("AEIOU"), "");
    }

    #[test]
    fn test_generate_contact_id() {
        let mut assigned: HashSet<String> = HashSet::new();
        let put = |id: &str| -> Result<(), IdError> {
            if assigned.contains(id) {
                Err(IdError::GenerationFailed("already exists".to_string()))
            } else {
                assigned.insert(id.to_string());
                Ok(())
            }
        };

        let id = generate_contact_id("Jane", "Doe", None, put).unwrap();
        assert!(id.len() >= 4);
        assert!(id.len() <= 9);
    }
}
