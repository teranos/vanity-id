//! Fuzzy matching functionality for ID lookups.
//!
//! This module provides functions for fuzzy matching IDs, enabling
//! typo-tolerant and approximate ID searches.

use crate::id::normalize_for_lookup;

/// Determines if two IDs are similar using fuzzy matching
pub fn is_similar(query: &str, candidate: &str) -> bool {
    // Use normalize_for_lookup for consistent normalization (case + character mapping)
    let query_norm = normalize_for_lookup(query);
    let candidate_norm = normalize_for_lookup(candidate);

    // Skip exact matches (those would have been found already)
    if query_norm == candidate_norm {
        return false;
    }

    // 1. Substring match - query appears in candidate
    if candidate_norm.contains(&query_norm) {
        return true;
    }

    // 2. Candidate starts with query (common for partial typing)
    if candidate_norm.starts_with(&query_norm) {
        return true;
    }

    // 3. Edit distance check for close matches (simple version)
    if query.len() >= 3 && candidate.len() >= 3 {
        return has_low_edit_distance(&query_norm, &candidate_norm);
    }

    false
}

/// Checks if two strings have low edit distance (simple version)
fn has_low_edit_distance(s1: &str, s2: &str) -> bool {
    // Simple heuristic: if they share a significant portion of characters
    // and lengths are similar, consider them similar
    let len_diff = (s1.len() as isize - s2.len() as isize).unsigned_abs();

    // If length difference is too big, not similar
    if len_diff > 2 {
        return false;
    }

    // Count common characters at matching positions (simple approach)
    let chars1: Vec<char> = s1.chars().collect();
    let chars2: Vec<char> = s2.chars().collect();

    let mut common_chars = 0;
    let min_len = chars1.len().min(chars2.len());

    for i in 0..min_len {
        if chars1[i] == chars2[i] {
            common_chars += 1;
        }
    }

    // If most characters match, consider similar
    (common_chars as f64 / min_len as f64) > 0.6
}

/// Sorts IDs by how similar they are to the query
pub fn sort_by_similarity(ids: &[String], query: &str) -> Vec<String> {
    let mut sorted = ids.to_vec();
    let query_len = query.len();

    // Simple bubble sort by string length difference (closer lengths = more similar)
    for i in 0..sorted.len() {
        for j in (i + 1)..sorted.len() {
            let i_diff = (sorted[i].len() as isize - query_len as isize).unsigned_abs();
            let j_diff = (sorted[j].len() as isize - query_len as isize).unsigned_abs();

            if j_diff < i_diff {
                sorted.swap(i, j);
            }
        }
    }

    sorted
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_similar_exact_match_returns_false() {
        // Exact matches should return false (not "similar", they're the same)
        assert!(!is_similar("JDOE", "JDOE"));
    }

    #[test]
    fn test_is_similar_substring() {
        assert!(is_similar("JDO", "JDOE")); // Query is substring of candidate
    }

    #[test]
    fn test_is_similar_prefix() {
        assert!(is_similar("JDOE", "JDOE2")); // Candidate starts with query
    }

    #[test]
    fn test_is_similar_character_mapping() {
        // After normalization, JD0E becomes JDOE, so it's an exact match (returns false)
        // The is_similar function returns false for exact matches
        assert!(!is_similar("JD0E", "JDOE"));

        // But with actual different characters, it should still find similarity
        assert!(is_similar("JD0", "JDOE")); // JD0 -> JDO, which is substring of JDOE
    }

    #[test]
    fn test_is_similar_low_edit_distance() {
        assert!(is_similar("JDOA", "JDOE")); // One character different
    }

    #[test]
    fn test_is_similar_too_different() {
        assert!(!is_similar("ABCD", "WXYZ")); // Completely different
    }

    #[test]
    fn test_sort_by_similarity() {
        let ids = vec![
            "JDOE123".to_string(),
            "JDO".to_string(),
            "JDOE".to_string(),
            "JDOEEEE".to_string(),
        ];

        let sorted = sort_by_similarity(&ids, "JDOE");

        // "JDOE" should be first (exact length match)
        assert_eq!(sorted[0], "JDOE");
    }

    #[test]
    fn test_has_low_edit_distance() {
        assert!(has_low_edit_distance("JDOE", "JDOA")); // One char diff
        assert!(has_low_edit_distance("JDOE", "JDOEE")); // Length diff of 1
        assert!(!has_low_edit_distance("JDOE", "JDOEEEEE")); // Length diff > 2
        assert!(!has_low_edit_distance("AAAA", "BBBB")); // Too different
    }
}
